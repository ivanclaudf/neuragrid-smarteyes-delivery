package queue

import (
	"context"
	"delivery/helper"
	"delivery/models"
	"delivery/services/providers"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"gorm.io/gorm"
)

const (
	// SMSConsumerSubscription is the subscription name for SMS consumers
	SMSConsumerSubscription = "sms-consumer"
)

// SMSConsumer consumes SMS messages from the queue
type SMSConsumer struct {
	pulsarClient *PulsarClient
	db           *gorm.DB
	readerDB     *gorm.DB
	consumer     pulsar.Consumer
	running      bool
}

// NewSMSConsumer creates a new SMS consumer
func NewSMSConsumer(pulsarClient *PulsarClient, db *gorm.DB, readerDB *gorm.DB) (*SMSConsumer, error) {
	return &SMSConsumer{
		pulsarClient: pulsarClient,
		db:           db,
		readerDB:     readerDB,
		running:      false,
	}, nil
}

// Start starts the SMS consumer
func (c *SMSConsumer) Start() error {
	if c.running {
		return nil
	}

	if c.pulsarClient == nil {
		return errors.New("pulsar client is nil")
	}

	var err error
	c.consumer, err = c.pulsarClient.client.Subscribe(pulsar.ConsumerOptions{
		Topic:            SMSTopic,
		SubscriptionName: SMSConsumerSubscription,
		Type:             pulsar.Shared,
	})

	if err != nil {
		return err
	}

	c.running = true
	go c.consume()
	return nil
}

// Stop stops the SMS consumer
func (c *SMSConsumer) Stop() error {
	if !c.running {
		return nil
	}

	c.running = false
	if c.consumer != nil {
		c.consumer.Close()
	}

	return nil
}

// consume consumes SMS messages from the queue
func (c *SMSConsumer) consume() {
	for c.running {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msg, err := c.consumer.Receive(ctx)
		cancel()

		if err != nil {
			if err != context.DeadlineExceeded {
				helper.Log.WithError(err).Error("Failed to receive SMS message from queue")
			}
			continue
		}

		// Process the message
		err = c.processSMSMessage(msg)
		if err != nil {
			helper.Log.WithError(err).Error("Failed to process SMS message")
			c.consumer.Nack(msg)
		} else {
			c.consumer.Ack(msg)
		}
	}
}

// processSMSMessage processes an SMS message from the queue
func (c *SMSConsumer) processSMSMessage(msg pulsar.Message) error {
	helper.Log.Debug("Processing SMS message from queue")

	// Parse the queue message
	var smsMessage SMSMessage
	if err := json.Unmarshal(msg.Payload(), &smsMessage); err != nil {
		helper.Log.WithError(err).Error("Failed to unmarshal queue message")
		return fmt.Errorf("failed to unmarshal queue message: %w", err)
	}

	message := smsMessage.Message
	messageUUID := smsMessage.UUID

	messageLogger := helper.Log.WithFields(map[string]interface{}{
		"message_uuid": messageUUID,
		"refNo":        message.RefNo,
		"provider":     message.Provider,
		"template":     message.Template,
		"tenantId":     message.TenantID,
		"recipients":   len(message.To),
	})

	messageLogger.Info("Processing SMS message")

	// Get the message from the database or create if not found
	dbMessage, err := c.fetchOrCreateMessageFromDB(messageUUID, message)
	if err != nil {
		return err
	}

	// Check if template exists in our database
	template, err := c.fetchTemplateFromDB(message.Template, message.TenantID)
	if err != nil {
		return c.rejectMessage(dbMessage, fmt.Sprintf("template not found or inactive: %s", message.Template))
	}

	// Check if provider exists and is active
	var provider models.Provider
	if err := c.readerDB.Where("uuid = ? AND channel = ?", message.Provider, models.ChannelSMS).
		Where("status = 1").
		First(&provider).Error; err != nil {

		messageLogger.WithError(err).Error("Failed to find SMS provider in database")
		return c.rejectMessage(dbMessage, fmt.Sprintf("provider not found or inactive: %s", message.Provider))
	}

	messageLogger.WithField("provider", provider.Provider).Info("Found SMS provider in database")

	// Create SMS service based on provider using the factory
	smsService, provErr := providers.CreateSMSProvider(&provider)
	if provErr != nil {
		messageLogger.WithError(provErr).Error("Failed to create SMS provider")
		return c.rejectMessage(dbMessage, fmt.Sprintf("failed to create SMS provider: %v", provErr))
	}

	// Set initial message status to ACCEPTED (processing state)
	dbMessage.Status = models.StatusAccepted
	if err := c.db.Save(dbMessage).Error; err != nil {
		helper.Log.WithError(err).Error("Failed to update message status to ACCEPTED")
		return fmt.Errorf("failed to update message status: %w", err)
	}

	// Log the template content that will be used
	messageLogger.WithFields(map[string]interface{}{
		"template_uuid": template.UUID,
		"template_name": template.Name,
		"content":       template.Content,
	}).Debug("Using template content for rendering")

	// Extract the telephone number from the first recipient (assuming single recipient for now)
	if len(message.To) == 0 {
		messageLogger.Error("No recipients found in SMS message")
		return c.rejectMessage(dbMessage, "no recipients found in SMS message")
	}

	toNumber := message.To[0].Telephone

	// Render the template with variables using Go's text/template
	messageLogger.WithFields(map[string]interface{}{
		"telephone": toNumber,
		"params":    message.Params,
		"template":  template.Content,
	}).Debug("Rendering template with parameters")

	renderedContent, err := helper.RenderTemplate(template.Content, message.Params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to render template for %s: %v", toNumber, err)
		messageLogger.WithError(err).WithFields(map[string]interface{}{
			"telephone": toNumber,
			"params":    message.Params,
			"template":  template.Content,
		}).Error("Template rendering failed")

		return c.rejectMessage(dbMessage, errMsg)
	}

	messageLogger.WithFields(map[string]interface{}{
		"telephone":        toNumber,
		"original_content": template.Content,
		"rendered_content": renderedContent,
	}).Debug("Template rendered successfully")

	messageLogger.WithFields(map[string]interface{}{
		"to":            toNumber,
		"template":      template.Name,
		"content":       renderedContent,
		"provider":      provider.Provider,
		"provider_uuid": provider.UUID,
	}).Info("Sending SMS message from template")

	// For SendTemplate, add a special parameter with the rendered content
	paramsWithRenderedContent := make(map[string]string)
	for k, v := range message.Params {
		paramsWithRenderedContent[k] = v
	}
	// Add a special parameter for the rendered content
	paramsWithRenderedContent["rendered_content"] = renderedContent

	// Send the SMS with the rendered template content using template API
	if err := smsService.SendTemplate(toNumber, template.Name, paramsWithRenderedContent); err != nil {
		errMsg := fmt.Sprintf("Failed to send SMS message to %s: %v", toNumber, err)
		messageLogger.WithError(err).Error("Failed to send SMS message")
		return c.rejectMessage(dbMessage, errMsg)
	}

	// Create successful send event
	eventErr := c.createSuccessEvent(dbMessage.ID)
	if eventErr != nil {
		messageLogger.WithError(eventErr).Error("Failed to create success event")
	}

	// Update message status to SENT
	dbMessage.Status = models.StatusSent
	if err := c.db.Save(dbMessage).Error; err != nil {
		messageLogger.WithError(err).Info("Failed to update message status to SENT - will continue processing")
	}

	// Update message timestamp
	if err := c.updateMessageTimestamp(dbMessage); err != nil {
		return err
	}

	messageLogger.Info("Successfully processed SMS message")
	return nil
}

// fetchOrCreateMessageFromDB gets a message from the database by UUID or creates one if not found
func (c *SMSConsumer) fetchOrCreateMessageFromDB(messageUUID string, message models.SMSMessage) (*models.Message, error) {
	helper.Log.WithField("message_uuid", messageUUID).Debug("Fetching message from database")

	var dbMessage models.Message
	if err := c.db.Where("uuid = ?", messageUUID).First(&dbMessage).Error; err != nil {
		helper.Log.WithField("message_uuid", messageUUID).Info("Message not found in database, will create new one")

		// Save the Identifiers object as-is
		identifiersJSON := message.Identifiers

		// Convert categories array to JSON
		categoriesJSON := models.JSON{}
		for i, category := range message.Categories {
			categoriesJSON[fmt.Sprintf("%d", i)] = category
		}

		// Generate UUID for new message if not provided
		uuid := messageUUID
		if uuid == "" {
			var err error
			uuid, err = helper.GenerateUUID()
			if err != nil {
				helper.Log.WithError(err).Error("Failed to generate UUID for new message")
				return nil, fmt.Errorf("failed to generate UUID for new message: %w", err)
			}
		}
		newMessage := models.Message{
			UUID:        uuid,
			Channel:     models.ChannelSMS,
			Status:      models.StatusAccepted,
			RefNo:       message.RefNo,
			Identifiers: identifiersJSON,
			Categories:  categoriesJSON,
			TenantID:    message.TenantID,
		}

		if err := c.db.Create(&newMessage).Error; err != nil {
			helper.Log.WithError(err).WithField("message_uuid", uuid).Error("Failed to create new message")
			return nil, fmt.Errorf("failed to create new message with UUID %s: %w", uuid, err)
		}

		// Create an ACCEPTED event in message_events table
		event := models.MessageEvent{
			MessageID: newMessage.ID,
			Status:    models.EventStatusAccepted,
			Timestamp: time.Now().UTC(),
			Reason:    "Message created during processing due to missing record",
		}
		if err := helper.InsertMessageEvent(c.db, event); err != nil {
			helper.Log.WithError(err).WithField("message_id", newMessage.ID).Error("Failed to create acceptance event")
		}

		helper.Log.WithField("message_uuid", messageUUID).Info("Created new message with ACCEPTED status")
		return &newMessage, nil
	}

	helper.Log.WithFields(map[string]interface{}{
		"message_uuid": messageUUID,
		"status":       dbMessage.Status,
	}).Debug("Message fetched from database")

	return &dbMessage, nil
}

// fetchMessageFromDB gets a message from the database by UUID
func (c *SMSConsumer) fetchMessageFromDB(messageUUID string) (*models.Message, error) {
	helper.Log.WithField("message_uuid", messageUUID).Debug("Fetching message from database")

	var dbMessage models.Message
	if err := c.db.Where("uuid = ?", messageUUID).First(&dbMessage).Error; err != nil {
		helper.Log.WithError(err).WithField("message_uuid", messageUUID).Error("Failed to fetch message from database")
		return nil, fmt.Errorf("failed to fetch message with UUID %s: %w", messageUUID, err)
	}

	helper.Log.WithFields(map[string]interface{}{
		"message_uuid": messageUUID,
		"status":       dbMessage.Status,
	}).Debug("Message fetched from database")

	return &dbMessage, nil
}

// fetchTemplateFromDB gets a template from the database by UUID and tenant
func (c *SMSConsumer) fetchTemplateFromDB(templateUUID string, tenantID string) (*models.Template, error) {
	helper.Log.WithFields(map[string]interface{}{
		"template_uuid": templateUUID,
		"tenantId":      tenantID,
	}).Debug("Fetching template from database")

	var template models.Template
	if err := c.readerDB.Where("uuid = ?", templateUUID).
		Where("tenant_id = ?", tenantID).
		Where("channel = ?", models.ChannelSMS).
		Where("status = ?", 1).
		First(&template).Error; err != nil {

		helper.Log.WithError(err).WithFields(map[string]interface{}{
			"template_uuid": templateUUID,
			"tenantId":      tenantID,
		}).Error("Template not found or inactive")

		return nil, fmt.Errorf("template not found or inactive: %w", err)
	}

	helper.Log.WithFields(map[string]interface{}{
		"template_uuid": templateUUID,
		"template_name": template.Name,
		"content":       template.Content,
	}).Debug("Template fetched from database")

	return &template, nil
}

// createRejectionEvent creates a message event with rejection status and reason
func (c *SMSConsumer) createRejectionEvent(messageID uint, reason string) error {
	event := models.MessageEvent{
		MessageID: messageID,
		Status:    models.EventStatusRejected,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	if err := helper.InsertMessageEvent(c.db, event); err != nil {
		helper.Log.WithError(err).Error("Failed to create rejection event in database")
		return fmt.Errorf("failed to create rejection event: %w", err)
	}
	return nil
}

// createSuccessEvent creates a message event with success status
func (c *SMSConsumer) createSuccessEvent(messageID uint) error {
	event := models.MessageEvent{
		MessageID: messageID,
		Status:    models.EventStatusSent,
		Reason:    "Message sent successfully",
		Timestamp: time.Now(),
	}
	if err := helper.InsertMessageEvent(c.db, event); err != nil {
		helper.Log.WithError(err).Error("Failed to create success event in database")
		return fmt.Errorf("failed to create success event: %w", err)
	}
	return nil
}

// rejectMessage updates message status to rejected and creates a rejection event
func (c *SMSConsumer) rejectMessage(message *models.Message, reason string) error {
	helper.Log.WithFields(map[string]interface{}{
		"message_uuid": message.UUID,
		"reason":       reason,
	}).Info("Rejecting SMS message")

	// Update message status to rejected
	message.Status = models.StatusRejected
	if err := c.db.Save(message).Error; err != nil {
		helper.Log.WithError(err).Error("Failed to update message status to REJECTED")
		return err
	}

	// Create rejection event
	return c.createRejectionEvent(message.ID, reason)
}

// updateMessageTimestamp updates the message timestamp in the database
func (c *SMSConsumer) updateMessageTimestamp(message *models.Message) error {
	message.UpdatedAt = time.Now()
	if err := c.db.Save(message).Error; err != nil {
		helper.Log.WithError(err).WithField("message_uuid", message.UUID).Error("Failed to update message timestamp")
		return fmt.Errorf("failed to update message timestamp: %w", err)
	}

	helper.Log.WithField("message_uuid", message.UUID).Debug("Updated message timestamp")
	return nil
}

// updateMessageStatus updates the status of a message
func (c *SMSConsumer) updateMessageStatus(uuid string, status models.Status) error {
	return c.db.Model(&models.Message{}).Where("uuid = ?", uuid).Update("status", status).Error
}
