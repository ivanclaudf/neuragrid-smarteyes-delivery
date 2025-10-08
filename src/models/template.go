package models

import (
	"time"
)

// Template represents a message template in the database
type Template struct {
	ID          uint      `gorm:"primarykey"`
	UUID        string    `gorm:"type:varchar(36);uniqueIndex;not null"`
	Name        string    `gorm:"type:varchar(255);not null;index"`
	Content     string    `gorm:"type:text;not null"`
	Status      int       `gorm:"type:smallint;default:0;not null;index"` // 0 for inactive, 1 for active
	Channel     Channel   `gorm:"type:varchar(10);not null;index;check:channel IN ('WHATSAPP', 'SMS', 'EMAIL')"`
	TemplateIds JSON      `gorm:"type:jsonb;column:template_ids"` // JSON field to store provider template IDs
	Tenant      string    `gorm:"type:varchar(255);not null;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime;not null;index"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;not null"`
}
