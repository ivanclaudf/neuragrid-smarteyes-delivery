package models

import (
	"time"
)

// MessageEventType represents the status of a message event
type MessageEventType string

const (
	// EventStatusDelivered indicates the message was delivered to the recipient
	EventStatusDelivered MessageEventType = "DELIVERED"

	// EventStatusRead indicates the message was read by the recipient
	EventStatusRead MessageEventType = "READ"

	// EventStatusSent indicates the message was sent by the provider
	EventStatusSent MessageEventType = "SENT"

	// Additional status types that match Status in message.go
	EventStatusAccepted MessageEventType = "ACCEPTED"
	EventStatusRejected MessageEventType = "REJECTED"
)

// MessageEvent represents an event related to a message in the database
type MessageEvent struct {
	ID        uint             `gorm:"primarykey"`
	UUID      string           `gorm:"type:varchar(36);uniqueIndex;not null"`
	MessageID uint             `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:ID"` // Foreign key to Message.ID
	Status    MessageEventType `gorm:"type:varchar(10);not null;index;check:status IN ('DELIVERED', 'FAILED', 'READ', 'SENT', 'ACCEPTED', 'REJECTED')"`
	Reason    string           `gorm:"type:text;column:reason"` // Reason for status change, especially for failures
	Metadata  JSON             `gorm:"type:jsonb"`
	Timestamp time.Time        `gorm:"not null;index"` // Timestamp of when the event occurred
	CreatedAt time.Time        `gorm:"autoCreateTime;not null;index"`
}
