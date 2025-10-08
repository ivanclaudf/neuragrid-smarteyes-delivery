package models

import (
	"time"
)

// Template represents a message template in the database
type Template struct {
	ID          uint      `gorm:"primarykey"`
	UUID        string    `gorm:"type:varchar(36);uniqueIndex;not null"`
	Code        string    `gorm:"type:varchar(50);not null;index"` // Non-editable unique code for the template
	Name        string    `gorm:"type:varchar(255);not null;index"`
	Content     string    `gorm:"type:text;not null"`
	Status      int       `gorm:"type:smallint;default:0;not null;index"` // 0 for inactive, 1 for active
	Channel     Channel   `gorm:"type:varchar(10);not null;index;check:channel IN ('WHATSAPP', 'SMS', 'EMAIL')"`
	TemplateIds JSON      `gorm:"type:jsonb;column:template_ids"` // JSON field to store provider template IDs
	Tenant      string    `gorm:"type:varchar(255);not null;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime;not null;index"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;not null"`

	// Define unique constraint: tenant + code + channel must be unique
	_ struct{} `gorm:"uniqueIndex:idx_tenant_code_channel;columns:tenant,code,channel"`
}

// TableName defines the table name for the Template model
func (Template) TableName() string {
	return "templates"
}
