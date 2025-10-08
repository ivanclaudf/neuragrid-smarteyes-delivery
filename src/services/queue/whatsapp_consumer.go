package queue

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services"
	"delivery/services/providers"
	"delivery/services/providers/whatsapp"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// WhatsAppConsumer handles consuming WhatsApp messages from the queue
type WhatsAppConsumer struct {
	pulsarClient *PulsarClient
	db           *gorm.DB
	readerDB     *gorm.DB
}

// NewWhatsAppConsumer creates a new WhatsApp consumer
func NewWhatsAppConsumer(pulsarClient *PulsarClient, db *gorm.DB, readerDB *gorm.DB) *WhatsAppConsumer {
	return &WhatsAppConsumer{
		pulsarClient: pulsarClient,
		db:           db,
		readerDB:     readerDB,
	}
}

// Start starts consuming messages
func (c *WhatsAppConsumer) Start(numConsumers ...int) error {
	workers := 3 // Default number of consumers
	if len(numConsumers) > 0 && numConsumers[0] > 0 {
		workers = numConsumers[0]
	}

	return c.pulsarClient.CreateConsumerGroup(
		WhatsAppTopic,
		"whatsapp-consumer",
		workers,
		c.handleMessage,
	)
}

// createRejectionEvent creates a message event with rejection status and reason
func (c *WhatsAppConsumer) createRejectionEvent(messageID string, reason string) error {
	uuid, err := helper.GenerateUUID()
	if err != nil {
		helper.Log.WithError(err).Error("Failed to generate UUID for rejection event")
		return fmt.Errorf("failed to generate event UUID: %w", err)
	}

	event := models.MessageEvent{
		UUID:      uuid,
		MessageID: messageID,
		Status:    models.EventStatusRejected,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	if err := c.db.Create(&event).Error; err != nil {
		helper.Log.WithError(err).Error("Failed to create rejection event in database")
		return fmt.Errorf("failed to create rejection event: %w", err)
	}

	return nil
}

// createSuccessEvent creates a message event with success status
func (c *WhatsAppConsumer) createSuccessEvent(messageID string) error {
	uuid, err := helper.GenerateUUID()
	if err != nil {
		helper.Log.WithError(err).Error("Failed to generate UUID for success event")
		return fmt.Errorf("failed to generate event UUID: %w", err)
	}

	event := models.MessageEvent{
		UUID:      uuid,
		MessageID: messageID,
		Status:    models.EventStatusSent,
		Reason:    "Message sent successfully",
		Timestamp: time.Now(),
	}

	if err := c.db.Create(&event).Error; err != nil {
		helper.Log.WithError(err).Error("Failed to create success event in database")
		return fmt.Errorf("failed to create success event: %w", err)
	}

	return nil
}

// rejectMessage updates message status to rejected and creates a rejection event
func (c *WhatsAppConsumer) rejectMessage(message *models.Message, reason string) error {
	helper.Log.WithFields(map[string]interface{}{
		"message_uuid": message.UUID,
		"reason":       reason,
	}).Error("Rejecting message")

	message.Status = models.StatusRejected
	if err := c.db.Save(message).Error; err != nil {
		helper.Log.WithError(err).Error("Failed to update message status to REJECTED")
		return fmt.Errorf("failed to update message status: %w", err)
	}

	if err := c.createRejectionEvent(message.UUID, reason); err != nil {
		return err
	}

	return errors.New(reason)
}

// fetchMessageFromDB gets a message from the database by UUID
func (c *WhatsAppConsumer) fetchMessageFromDB(messageUUID string) (*models.Message, error) {
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
func (c *WhatsAppConsumer) fetchTemplateFromDB(templateUUID string, tenant string) (*models.Template, error) {
	helper.Log.WithFields(map[string]interface{}{
		"template_uuid": templateUUID,
		"tenant":        tenant,
	}).Debug("Fetching template from database")

	var template models.Template
	if err := c.readerDB.Where("uuid = ?", templateUUID).
		Where("tenant = ?", tenant).
		Where("channel = ?", models.ChannelWhatsApp).
		Where("status = 1").
		First(&template).Error; err != nil {

		helper.Log.WithError(err).WithFields(map[string]interface{}{
			"template_uuid": templateUUID,
			"tenant":        tenant,
		}).Error("Template not found or inactive")

		return nil, fmt.Errorf("template not found or inactive: %w", err)
	}

	helper.Log.WithFields(map[string]interface{}{
		"template_uuid": templateUUID,
		"template_name": template.Name,
	}).Debug("Template fetched from database")

	return &template, nil
}

// fetchProviderFromDB gets a provider from the database by UUID
func (c *WhatsAppConsumer) fetchProviderFromDB(providerUUID string) (*models.Provider, error) {
	helper.Log.WithField("provider_uuid", providerUUID).Debug("Fetching provider from database")

	var provider models.Provider
	if err := c.readerDB.Where("uuid = ?", providerUUID).
		Where("channel = ?", models.ChannelWhatsApp).
		Where("status = 1").
		First(&provider).Error; err != nil {

		helper.Log.WithError(err).WithField("provider_uuid", providerUUID).Error("Provider not found or inactive")
		return nil, fmt.Errorf("provider not found or inactive: %w", err)
	}

	helper.Log.WithFields(map[string]interface{}{
		"provider_uuid": providerUUID,
		"provider_name": provider.Name,
		"provider_type": provider.Provider,
	}).Debug("Provider fetched from database")

	return &provider, nil
}

// getProviderTemplateID gets the provider-specific template ID from the template
func (c *WhatsAppConsumer) getProviderTemplateID(template *models.Template, providerName string) (string, error) {
	helper.Log.WithFields(map[string]interface{}{
		"template_uuid": template.UUID,
		"provider":      providerName,
	}).Debug("Getting provider-specific template ID")

	if template.TemplateIds == nil {
		helper.Log.WithField("template_uuid", template.UUID).Error("Template has no template_ids field")
		return "", errors.New("template_ids field is empty in template")
	}

	providerKey := strings.ToLower(providerName)
	providerID, exists := template.TemplateIds[providerKey]
	if !exists {
		helper.Log.WithFields(map[string]interface{}{
			"template_uuid": template.UUID,
			"provider":      providerName,
			"provider_key":  providerKey,
		}).Error("Template ID not found for provider")

		return "", fmt.Errorf("template ID not found for provider %s", providerName)
	}

	providerIDStr, ok := providerID.(string)
	if !ok {
		helper.Log.WithFields(map[string]interface{}{
			"template_uuid": template.UUID,
			"provider":      providerName,
			"provider_id":   providerID,
		}).Error("Invalid template ID format")

		return "", fmt.Errorf("invalid template ID format for provider %s", providerName)
	}

	helper.Log.WithFields(map[string]interface{}{
		"provider":    providerName,
		"template_id": providerIDStr,
	}).Debug("Found provider-specific template ID")

	return providerIDStr, nil
}

// sendToRecipients sends messages to all recipients
func (c *WhatsAppConsumer) sendToRecipients(
	whatsappProvider services.WhatsAppService,
	dbMessage *models.Message,
	recipients []models.WhatsAppRecipient,
	templateID string,
	params map[string]string,
) {
	helper.Log.WithField("recipient_count", len(recipients)).Info("Processing recipients")

	for _, recipient := range recipients {
		helper.Log.WithFields(map[string]interface{}{
			"telephone":   recipient.Telephone,
			"template_id": templateID,
		}).Info("Sending WhatsApp message")

		// Send the template message using the provider-specific template ID
		err := whatsappProvider.SendTemplate(recipient.Telephone, templateID, params)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to send WhatsApp message to %s: %v", recipient.Telephone, err)
			helper.Log.WithError(err).WithField("telephone", recipient.Telephone).Error("Send failed")

			// Create message event with error text but continue with next recipient
			eventErr := c.createRejectionEvent(dbMessage.UUID, errMsg)
			if eventErr != nil {
				helper.Log.WithError(eventErr).Error("Failed to create rejection event")
			}
			continue
		}

		// Create successful send event
		eventErr := c.createSuccessEvent(dbMessage.UUID)
		if eventErr != nil {
			helper.Log.WithError(eventErr).Error("Failed to create success event")
		} else {
			helper.Log.WithField("telephone", recipient.Telephone).Info("Message sent successfully")
		}
	}
}

// updateMessageTimestamp updates the message timestamp in the database
func (c *WhatsAppConsumer) updateMessageTimestamp(message *models.Message) error {
	message.UpdatedAt = time.Now()
	if err := c.db.Save(message).Error; err != nil {
		helper.Log.WithError(err).WithField("message_uuid", message.UUID).Error("Failed to update message timestamp")
		return fmt.Errorf("failed to update message timestamp: %w", err)
	}

	helper.Log.WithField("message_uuid", message.UUID).Debug("Updated message timestamp")
	return nil
}

// handleMessage handles a WhatsApp message from the queue
func (c *WhatsAppConsumer) handleMessage(data []byte) error {
	helper.Log.Debug("Processing WhatsApp message from queue")

	// Parse the queue message
	var queueMessage WhatsAppMessage
	if err := json.Unmarshal(data, &queueMessage); err != nil {
		helper.Log.WithError(err).Error("Failed to unmarshal queue message")
		return fmt.Errorf("failed to unmarshal queue message: %w", err)
	}

	message := queueMessage.Message
	messageUUID := queueMessage.UUID

	helper.Log.WithFields(map[string]interface{}{
		"message_uuid": messageUUID,
		"provider":     message.Provider,
		"template":     message.Template,
		"recipients":   len(message.To),
	}).Info("Processing WhatsApp message")

	// Get the message from the database
	dbMessage, err := c.fetchMessageFromDB(messageUUID)
	if err != nil {
		return err
	}

	// Check if template exists in our database
	template, err := c.fetchTemplateFromDB(message.Template, message.Identifiers.Tenant)
	if err != nil {
		return c.rejectMessage(dbMessage, fmt.Sprintf("template not found or inactive: %s", message.Template))
	}

	// Check if provider exists and is active
	provider, err := c.fetchProviderFromDB(message.Provider)
	if err != nil {
		return c.rejectMessage(dbMessage, fmt.Sprintf("provider not found or inactive: %s", message.Provider))
	}

	// Extract provider-specific template ID from template_ids JSON field
	templateID, err := c.getProviderTemplateID(template, provider.Provider)
	if err != nil {
		return c.rejectMessage(dbMessage, err.Error())
	}

	// Use the provider to send the template
	whatsappProvider, err := c.createProviderFromConfig(provider)
	if err != nil {
		return c.rejectMessage(dbMessage, fmt.Sprintf("failed to create WhatsApp provider: %v", err))
	}

	// Update database status to SENT
	dbMessage.Status = models.StatusSent
	if err := c.db.Save(dbMessage).Error; err != nil {
		helper.Log.WithError(err).Error("Failed to update message status to SENT")
		return fmt.Errorf("failed to update message status: %w", err)
	}

	// Send to all recipients
	c.sendToRecipients(whatsappProvider, dbMessage, message.To, templateID, message.Params)

	// Update message timestamp
	return c.updateMessageTimestamp(dbMessage)
}

// createProviderFromConfig creates a WhatsApp provider from the provider configuration
func (c *WhatsAppConsumer) createProviderFromConfig(provider *models.Provider) (services.WhatsAppService, error) {
	// Log which provider we're creating
	helper.Log.WithField("provider_uuid", provider.UUID).Info("Creating WhatsApp provider")

	// Fetch the provider details from the database
	var fullProvider models.Provider
	if err := c.readerDB.Where("uuid = ?", provider.UUID).First(&fullProvider).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch provider details: %w", err)
	}

	// Create the provider using the factory
	whatsAppProvider, err := providers.CreateWhatsAppProvider(&fullProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp provider: %w", err)
	}

	// Add debug logging to check if we have auth credentials
	if twilioProvider, ok := whatsAppProvider.(*whatsapp.TwilioProvider); ok {
		// Mask the auth token for security, only log the first 4 and last 4 chars
		authTokenLen := len(twilioProvider.AuthToken)
		maskedToken := "****"
		if authTokenLen > 8 {
			maskedToken = twilioProvider.AuthToken[0:4] + "****" + twilioProvider.AuthToken[authTokenLen-4:]
		} else if authTokenLen > 0 {
			minLen := 4
			if authTokenLen < minLen {
				minLen = authTokenLen
			}
			maskedToken = "****" + twilioProvider.AuthToken[authTokenLen-minLen:]
		}

		helper.Log.WithFields(map[string]interface{}{
			"accountSID": twilioProvider.AccountSID,
			"authToken":  maskedToken,
			"fromNumber": twilioProvider.FromNumber,
			"baseURL":    twilioProvider.BaseURL,
		}).Debug("Twilio provider created with credentials")
	}

	return whatsAppProvider, nil
}
