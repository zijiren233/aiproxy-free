// Package module defines data structures for the aiproxy application.
package module

// RateLimitRecord 限流记录表
type RateLimitRecord struct {
	ID          uint   `gorm:"primaryKey"`
	Namespace   string `gorm:"size:255;not null;index:idx_namespace_timestamp"`
	RequestTime int64  `gorm:"not null;index:idx_namespace_timestamp"` // 毫秒时间戳
}

// TableName 指定表名
func (RateLimitRecord) TableName() string {
	return "rate_limit_records"
}
