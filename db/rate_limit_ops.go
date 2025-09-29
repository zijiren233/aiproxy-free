package db

import (
	"fmt"
	"time"

	"github.com/labring/aiproxy-free/module"
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
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
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