package models

import (
	"delivery/helper"
	"time"

	"gorm.io/gorm"
)

// Provider represents a message provider in the database
type Provider struct {
	ID           uint      `gorm:"primarykey"`
	UUID         string    `gorm:"type:varchar(36);uniqueIndex;not null"`
	Code         string    `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_code_tenant"`
	Provider     string    `gorm:"type:varchar(255);not null;index"` // Implementation class name (e.g. twilio)
	Name         string    `gorm:"type:varchar(255);not null;index"`
	Config       JSON      `gorm:"type:jsonb;not null"`
	SecureConfig JSON      `gorm:"type:jsonb;not null"`
	Status       int       `gorm:"type:smallint;default:0;not null;index"` // 0 for inactive, 1 for active
	Channel      Channel   `gorm:"type:varchar(10);not null;index;check:channel IN ('WHATSAPP', 'SMS', 'EMAIL')"`
	Tenant       string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_code_tenant"`
	CreatedAt    time.Time `gorm:"autoCreateTime;not null;index"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;not null"`
}

// BeforeCreate hook for Provider model to automatically generate a UUID
func (p *Provider) BeforeCreate(tx *gorm.DB) error {
	if p.UUID == "" {
		uuid, err := helper.GenerateUUID()
		if err != nil {
			return err
		}
		p.UUID = uuid
	}
	return nil
}
