package queue

import (
	"delivery/helper"
	"delivery/models"
	"fmt"

	"gorm.io/gorm"
)

// SMSProducer handles producing SMS messages to the queue
type SMSProducer struct {
	PulsarClient *PulsarClient
	db           *gorm.DB
}

// SMSMessage represents an SMS message in the queue
type SMSMessage struct {
	UUID    string            `json:"uuid"`
	Message models.SMSMessage `json:"message"`
}

// NewSMSProducer creates a new SMS producer
func NewSMSProducer(pulsarClient *PulsarClient, db *gorm.DB) *SMSProducer {
	producer := &SMSProducer{
		PulsarClient: pulsarClient,
		db:           db,
	}

	helper.Log.Info("SMS producer created successfully")
	return producer
}

// ProduceSMSMessage sends an SMS message to the queue
func (p *SMSProducer) ProduceSMSMessage(message *models.SMSMessage, uuid string) error {
	// Save the Identifiers object as-is
	identifiersJSON := message.Identifiers

	// Convert categories array to JSON
	categoriesJSON := models.JSON{}
	for i, category := range message.Categories {
		categoriesJSON[fmt.Sprintf("%d", i)] = category
	}

	// Create a message record in the database
	dbMessage := models.Message{
		UUID:        uuid,
		Channel:     models.ChannelSMS,
		Status:      models.StatusAccepted,
		Identifiers: identifiersJSON,
		RefNo:       message.RefNo,
		Categories:  categoriesJSON,
		TenantID:    message.TenantID,
	}

	// Save to database
	if err := p.db.Create(&dbMessage).Error; err != nil {
		return err
	}

	// Create queue message
	queueMessage := SMSMessage{
		UUID:    uuid,
		Message: *message,
	}

	// Produce the message to the queue
	return p.PulsarClient.ProduceMessage(SMSTopic, queueMessage)
}
