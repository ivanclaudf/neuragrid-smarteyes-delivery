package queue

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// EmailConsumer handles consuming email messages from the queue
type EmailConsumer struct {
	pulsarClient *PulsarClient
	db           *gorm.DB
	readerDB     *gorm.DB
}

// NewEmailConsumer creates a new email consumer
func NewEmailConsumer(pulsarClient *PulsarClient, db *gorm.DB, readerDB *gorm.DB) (*EmailConsumer, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	return &EmailConsumer{
		pulsarClient: pulsarClient,
		db:           db,
		readerDB:     readerDB,
	}, nil
}

// Start starts the email consumer
func (c *EmailConsumer) Start() error {
	helper.Log.Info("Starting Email consumer")
	subscription := "email-consumer-subscription"
	numWorkers := 1

	return c.pulsarClient.CreateConsumerGroup(EmailTopic, subscription, numWorkers, c.handleMessage)
}

// handleMessage processes a single email message from the queue
func (c *EmailConsumer) handleMessage(data []byte) error {
	var message EmailMessage
	if err := json.Unmarshal(data, &message); err != nil {
		helper.Log.Errorf("Failed to unmarshal Email message: %v", err)
		return err
	}

	logger := helper.Log.WithFields(map[string]interface{}{
		"uuid":     message.UUID,
		"template": message.Message.Template,
		"refNo":    message.Message.RefNo,
	})

	logger.Info("Processing Email message from queue")

	// Ensure the message exists in the database before processing
	if err := c.ensureMessageExists(message); err != nil {
		logger.WithError(err).Error("Failed to ensure message exists in database")
		return err
	}

	// Update the message status to SENT when it begins processing
	if err := c.updateMessageStatus(message.UUID, models.StatusSent); err != nil {
		logger.WithError(err).Info("Failed to update message status to SENT - will continue processing")
	}
	// Fetch the template
	var template models.Template
	if err := c.db.Where("uuid = ? AND channel = ?", message.Message.Template, models.ChannelEmail).First(&template).Error; err != nil {
		logger.WithError(err).Error("Failed to find template")
		if err := c.updateMessageStatus(message.UUID, models.StatusRejected); err != nil {
			logger.WithError(err).Error("Failed to update message status to REJECTED")
		}
		return fmt.Errorf("template not found: %w", err)
	}

	// Fetch the provider
	var provider models.Provider
	if err := c.db.Where("uuid = ? AND channel = ?", message.Message.Provider, models.ChannelEmail).First(&provider).Error; err != nil {
		logger.WithError(err).Error("Failed to find provider")
		if err := c.updateMessageStatus(message.UUID, models.StatusRejected); err != nil {
			logger.WithError(err).Error("Failed to update message status to REJECTED")
		}
		return fmt.Errorf("provider not found: %w", err)
	}

	// Create the email service to actually send the email
	emailService, err := services.NewEmailService(c.db)
	if err != nil {
		logger.WithError(err).Error("Failed to create email service")
		if err := c.updateMessageStatus(message.UUID, models.StatusRejected); err != nil {
			logger.WithError(err).Error("Failed to update message status to FAILED")
		}
		return fmt.Errorf("failed to create email service: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"provider":    provider.Name,
		"recipients":  len(message.Message.To),
		"template":    template.Name,
		"messageData": message.Message,
	}).Info("Sending email with provider")

	// Send the actual email
	if err := emailService.SendEmail(&message.Message); err != nil {
		logger.WithError(err).Error("Failed to send email")
		if err := c.updateMessageStatus(message.UUID, models.StatusRejected); err != nil {
			logger.WithError(err).Error("Failed to update message status to FAILED")
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully")

	// Update status to DELIVERED
	if err := c.updateMessageStatus(message.UUID, models.StatusDelivered); err != nil {
		logger.WithError(err).Error("Failed to update message status to DELIVERED")
		return err
	}

	logger.Info("Successfully processed Email message")
	return nil
}

// ensureMessageExists checks if a message exists in the database and creates it if not
func (c *EmailConsumer) ensureMessageExists(message EmailMessage) error {
	// Check if the message exists in the database
	var dbMessage models.Message
	if err := c.db.Where("uuid = ?", message.UUID).First(&dbMessage).Error; err != nil {
		// If not found, create it with proper identifiers and categories

		// Create identifiers JSON for the database
		identifiersJSON := models.JSON{
			"tenant":     message.Message.Identifiers.Tenant,
			"eventUuid":  message.Message.Identifiers.EventUUID,
			"actionUuid": message.Message.Identifiers.ActionUUID,
			"actionCode": message.Message.Identifiers.ActionCode,
		}

		// Convert categories array to JSON
		categoriesJSON := models.JSON{}
		for i, category := range message.Message.Categories {
			categoriesJSON[fmt.Sprintf("%d", i)] = category
		}

		// Create a new message with ACCEPTED status
		newMessage := models.Message{
			UUID:        message.UUID,
			Channel:     models.ChannelEmail,
			Status:      models.StatusAccepted,
			RefNo:       message.Message.RefNo,
			Identifiers: identifiersJSON,
			Categories:  categoriesJSON,
		}

		if err := c.db.Create(&newMessage).Error; err != nil {
			helper.Log.WithError(err).WithField("message_uuid", message.UUID).Error("Failed to create new message")
			return fmt.Errorf("failed to create new message with UUID %s: %w", message.UUID, err)
		}

		// Create an ACCEPTED event in message_events table
		event := models.MessageEvent{
			MessageID: message.UUID,
			Status:    models.EventStatusAccepted,
			Timestamp: time.Now().UTC(),
			Reason:    "Message created during processing due to missing record",
		}

		if err := c.db.Create(&event).Error; err != nil {
			helper.Log.WithError(err).WithField("message_uuid", message.UUID).Error("Failed to create acceptance event")
		}

		helper.Log.WithField("message_uuid", message.UUID).Info("Created new message with ACCEPTED status")
	}

	return nil
}

// updateMessageStatus updates the status of a message in the database
func (c *EmailConsumer) updateMessageStatus(uuid string, status models.Status) error {
	// Find the message by UUID
	var message models.Message
	if err := c.db.Where("uuid = ?", uuid).First(&message).Error; err != nil {
		helper.Log.WithError(err).WithField("message_uuid", uuid).Info("Message not found when updating status")
		return err
	}

	// Update the status
	message.Status = status
	if err := c.db.Save(&message).Error; err != nil {
		return err
	}

	// Create an event for the status change
	event := models.MessageEvent{
		MessageID: uuid,
		Status:    models.MessageEventType(status),
		Timestamp: time.Now().UTC(),
	}

	return c.db.Create(&event).Error
}
