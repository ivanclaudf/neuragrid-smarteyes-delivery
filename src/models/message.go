package models

import (
	"database/sql/driver"
	"delivery/helper"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Channel type for message channels
type Channel string

// Status type for message status
type Status string

const (
	// ChannelWhatsApp represents WhatsApp channel
	ChannelWhatsApp Channel = "WHATSAPP"
	// ChannelSMS represents SMS channel
	ChannelSMS Channel = "SMS"
	// ChannelEmail represents Email channel
	ChannelEmail Channel = "EMAIL"
)

// Message status constants - use MessageEventType from message_event.go instead
const (
	// StatusAccepted represents message is accepted but not yet processed
	StatusAccepted Status = "ACCEPTED"

	// StatusRejected represents message is rejected
	StatusRejected Status = "REJECTED"

	// The following statuses match MessageEventType in message_event.go
	// StatusSent represents message is sent to provider - matches EventStatusSent
	StatusSent Status = "SENT"

	// StatusDelivered represents message is delivered to recipient - matches EventStatusDelivered
	StatusDelivered Status = "DELIVERED"

	// StatusOpened represents message is opened by recipient - matches EventStatusRead
	StatusOpened Status = "READ"
)

// JSON type for storing JSON in database
type JSON map[string]interface{}

// Value converts JSON to value for database
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan converts database value to JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}

// Message represents a delivery message in the database
type Message struct {
	ID          uint      `gorm:"primarykey"`
	UUID        string    `gorm:"type:varchar(36);uniqueIndex;not null"`
	Channel     Channel   `gorm:"type:varchar(10);not null;index;check:channel IN ('WHATSAPP', 'SMS', 'EMAIL')"`
	Identifiers JSON      `gorm:"type:jsonb;not null"`
	Categories  JSON      `gorm:"type:jsonb"`
	RefNo       string    `gorm:"type:varchar(255);not null"`
	Status      Status    `gorm:"type:varchar(10);default:'ACCEPTED';not null;index;check:status IN ('ACCEPTED', 'SENT', 'DELIVERED', 'REJECTED', 'READ', 'FAILED')"`
	CreatedAt   time.Time `gorm:"autoCreateTime;not null;index"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;not null"`
}

// BeforeCreate hook for Message model to automatically generate a UUID
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		uuid, err := helper.GenerateUUID()
		if err != nil {
			return err
		}
		m.UUID = uuid
	}
	return nil
}
