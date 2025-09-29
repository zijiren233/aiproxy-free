package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/labring/aiproxy-free/module"
	"gorm.io/gorm"
)

// AddRequest 插入一个请求记录，返回记录ID
func AddRequest(namespace string) (uint, error) {
	record := &module.RateLimitRecord{
		Namespace:   namespace,
		RequestTime: time.Now().UnixMilli(),
	}

	result := gdb.Create(record)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to add request record: %w", result.Error)
	}

	return record.ID, nil
}

// CountRequestsToday 查询某个namespace在今天的请求数量
func CountRequestsToday(namespace string) (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		UnixMilli()
	endOfDay := startOfDay + 24*60*60*1000 - 1 // 一天的毫秒数减1

	var count int64

	result := gdb.Model(&module.RateLimitRecord{}).
		Where("namespace = ? AND request_time >= ? AND request_time <= ?", namespace, startOfDay, endOfDay).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to count requests today: %w", result.Error)
	}

	return count, nil
}

// DeleteRequestByID 根据ID删除请求记录
func DeleteRequestByID(id uint) error {
	result := gdb.Delete(&module.RateLimitRecord{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete request record: %w", result.Error)
	}

	return nil
}

// GetUsageInfo 获取某个namespace的使用情况信息
func GetUsageInfo(namespace string) (usedToday int64, nextResetTime time.Time, err error) {
	usedToday, err = CountRequestsToday(namespace)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to count requests today: %w", err)
	}

	// 查询一天内最早的请求时间
	nextResetTime, err = getNextResetTime(namespace)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to get next reset time: %w", err)
	}

	return usedToday, nextResetTime, nil
}

// getNextResetTime 获取下一次重置时间（基于最早请求的24小时后）
func getNextResetTime(namespace string) (time.Time, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		UnixMilli()
	endOfDay := startOfDay + 24*60*60*1000 - 1

	var record module.RateLimitRecord
	result := gdb.Where("namespace = ? AND request_time >= ? AND request_time <= ?",
		namespace, startOfDay, endOfDay).
		Order("request_time ASC").
		First(&record)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果没有今天的请求记录，下一次重置时间就是明天00:00
			return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()), nil
		}
		return time.Time{}, fmt.Errorf("failed to get earliest request today: %w", result.Error)
	}

	// 最早请求时间 + 24小时 = 下一次重置时间
	earliestRequestTime := time.UnixMilli(record.RequestTime)
	nextResetTime := earliestRequestTime.Add(24 * time.Hour)

	return nextResetTime, nil
}
