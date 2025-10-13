package queue

import (
	"delivery/models"
	"fmt"

	"gorm.io/gorm"
)

// WhatsAppProducer handles producing WhatsApp messages to the queue
type WhatsAppProducer struct {
	PulsarClient *PulsarClient
	db           *gorm.DB
}

// WhatsAppMessage represents a WhatsApp message in the queue
type WhatsAppMessage struct {
	UUID    string                 `json:"uuid"`
	Message models.WhatsAppMessage `json:"message"`
}

// NewWhatsAppProducer creates a new WhatsApp producer
func NewWhatsAppProducer(pulsarClient *PulsarClient, db *gorm.DB) *WhatsAppProducer {
	return &WhatsAppProducer{
		PulsarClient: pulsarClient,
		db:           db,
	}
}

// ProduceWhatsAppMessage produces a WhatsApp message to the queue
func (p *WhatsAppProducer) ProduceWhatsAppMessage(message *models.WhatsAppMessage, uuid string) error {
	// Create a new message record in the database with ACCEPTED status
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

	dbMessage := models.Message{
		UUID:        uuid,
		Channel:     models.ChannelWhatsApp,
		Identifiers: identifiersJSON,
		Categories:  categoriesJSON,
		RefNo:       message.RefNo,
		Status:      models.StatusAccepted,
	}

	if err := p.db.Create(&dbMessage).Error; err != nil {
		return err
	}

	// Create queue message
	queueMessage := WhatsAppMessage{
		UUID:    uuid,
		Message: *message,
	}

	// Produce the message to the queue
	return p.PulsarClient.ProduceMessage(WhatsAppTopic, queueMessage)
}
