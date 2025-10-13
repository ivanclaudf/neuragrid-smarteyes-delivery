package queue

import (
	"delivery/helper"
	"delivery/models"
	"fmt"

	"gorm.io/gorm"
)

// EmailProducer handles producing email messages to the queue
type EmailProducer struct {
	PulsarClient *PulsarClient
	db           *gorm.DB
}

// EmailMessage represents an email message in the queue
type EmailMessage struct {
	UUID    string              `json:"uuid"`
	Message models.EmailMessage `json:"message"`
}

// NewEmailProducer creates a new email producer
func NewEmailProducer(pulsarClient *PulsarClient, db *gorm.DB) *EmailProducer {
	producer := &EmailProducer{
		PulsarClient: pulsarClient,
		db:           db,
	}

	helper.Log.Info("Email producer created successfully")
	return producer
}

// ProduceEmailMessage sends an email message to the queue
func (p *EmailProducer) ProduceEmailMessage(message *models.EmailMessage, uuid string) error {
	// Create identifiers JSON for the database
	identifiersJSON := models.JSON{
		"tenant":     message.Identifiers.Tenant,
		"eventUuid":  message.Identifiers.EventUUID,
		"actionUuid": message.Identifiers.ActionUUID,
		"actionCode": message.Identifiers.ActionCode,
	}

	// Convert categories array to JSON
	categoriesJSON := models.JSON{}
	for i, category := range message.Categories {
		categoriesJSON[fmt.Sprintf("%d", i)] = category
	}

	// Create a message record in the database
	dbMessage := models.Message{
		UUID:        uuid,
		Channel:     models.ChannelEmail,
		Status:      models.StatusAccepted,
		Identifiers: identifiersJSON,
		RefNo:       message.RefNo,
		Categories:  categoriesJSON,
	}

	// Save to database
	if err := p.db.Create(&dbMessage).Error; err != nil {
		return err
	}

	// Create queue message
	queueMessage := EmailMessage{
		UUID:    uuid,
		Message: *message,
	}

	// Produce the message to the queue
	return p.PulsarClient.ProduceMessage(EmailTopic, queueMessage)
}
