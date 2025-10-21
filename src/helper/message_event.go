package helper

import (
	"delivery/models"
	"fmt"

	"gorm.io/gorm"
)

// InsertMessageEvent inserts a MessageEvent with a generated UUID
// MessageID must be set to the primary key (ID) of the Message
func InsertMessageEvent(db *gorm.DB, event models.MessageEvent) error {
	uuid, err := GenerateUUID()
	if err != nil {
		return fmt.Errorf("failed to generate UUID for MessageEvent: %w", err)
	}
	event.UUID = uuid
	return db.Create(&event).Error
}
