package models

import (
	"time"
)

// Provider represents a message provider in the database
type Provider struct {
	ID           uint      `gorm:"primarykey"`
	UUID         string    `gorm:"type:varchar(36);uniqueIndex;not null"`
	Code         string    `gorm:"type:varchar(255);not null;index"`
	Provider     string    `gorm:"type:varchar(255);not null;index"` // Implementation class name (e.g. twilio)
	Name         string    `gorm:"type:varchar(255);not null;index"`
	Config       JSON      `gorm:"type:jsonb;not null"`
	SecureConfig JSON      `gorm:"type:jsonb;not null"`
	Status       int       `gorm:"type:smallint;default:0;not null;index"` // 0 for inactive, 1 for active
	Channel      Channel   `gorm:"type:varchar(10);not null;index;check:channel IN ('WHATSAPP', 'SMS', 'EMAIL')"`
	TenantID     string    `gorm:"column:tenant_id;type:varchar(255);not null;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime;not null;index"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;not null"`

	// Define unique constraint: code + tenant_id + channel must be unique
	_ struct{} `gorm:"uniqueIndex:idx_code_tenant_channel;columns:code,tenant_id,channel"`
}
