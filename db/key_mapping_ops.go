package db

import (
	"fmt"

	"github.com/labring/aiproxy-free/module"
	"gorm.io/gorm"
)

// GetNamespace 查询某个key对应的namespace
func GetNamespace(key string) (string, error) {
	var mapping module.KeyMapping

	result := gdb.Where("key = ?", key).First(&mapping)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("key '%s' not found", key)
		}
		return "", fmt.Errorf("failed to get namespace for key '%s': %w", key, result.Error)
	}

	return mapping.Namespace, nil
}

// SaveMapping 保存key与namespace的对应关系（使用Save保证已存在也不报错）
func SaveMapping(key, namespace string) error {
	mapping := &module.KeyMapping{
		Key:       key,
		Namespace: namespace,
	}

	result := gdb.Save(mapping)
	if result.Error != nil {
		return fmt.Errorf("failed to save mapping for key '%s': %w", key, result.Error)
	}

	return nil
}

// UpdateMapping 更新已存在的key的namespace映射
func UpdateMapping(key, namespace string) error {
	result := gdb.Model(&module.KeyMapping{}).
		Where("key = ?", key).
		Update("namespace", namespace)

	if result.Error != nil {
		return fmt.Errorf("failed to update mapping for key '%s': %w", key, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("key '%s' not found", key)
	}

	return nil
}

// DeleteMapping 删除key映射
func DeleteMapping(key string) error {
	result := gdb.Delete(&module.KeyMapping{}, "key = ?", key)
	if result.Error != nil {
		return fmt.Errorf("failed to delete mapping for key '%s': %w", key, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("key '%s' not found", key)
	}

	return nil
}

// ListMappingsByNamespace 根据namespace查询所有相关的key映射
func ListMappingsByNamespace(namespace string) ([]module.KeyMapping, error) {
	var mappings []module.KeyMapping

	result := gdb.Where("namespace = ?", namespace).Find(&mappings)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list mappings for namespace '%s': %w", namespace, result.Error)
	}

	return mappings, nil
}

// KeyExists 检查key是否存在
func KeyExists(key string) (bool, error) {
	var count int64

	result := gdb.Model(&module.KeyMapping{}).Where("key = ?", key).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check if key '%s' exists: %w", key, result.Error)
	}

	return count > 0, nil
}

// CountKeysByNamespace 统计某个namespace下有多少个key
func CountKeysByNamespace(namespace string) (int64, error) {
	var count int64

	result := gdb.Model(&module.KeyMapping{}).Where("namespace = ?", namespace).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count keys for namespace '%s': %w", namespace, result.Error)
	}

	return count, nil
}

// GetAllMappings 获取所有映射（分页）
func GetAllMappings(limit, offset int) ([]module.KeyMapping, error) {
	var mappings []module.KeyMapping

	result := gdb.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&mappings)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all mappings: %w", result.Error)
	}

	return mappings, nil
}