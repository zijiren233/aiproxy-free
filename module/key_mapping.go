package module

import "time"

// KeyMapping key与namespace的映射表
type KeyMapping struct {
	Key       string    `gorm:"primaryKey;size:255"`            // key作为主键
	Namespace string    `gorm:"size:255;not null;index"`        // namespace，建立索引用于反向查询
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (KeyMapping) TableName() string {
	return "key_mappings"
}