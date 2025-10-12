package services

import (
	"delivery/api/types"
	"delivery/helper"
	"delivery/models"
	"delivery/services/providers/email"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// EmailServiceImpl implements the EmailService interface
type EmailServiceImpl struct {
	db *gorm.DB
}

// NewEmailService creates a new email service
func NewEmailService(db *gorm.DB) (*EmailServiceImpl, error) {
	if db == nil {
		return nil, errors.New("database connection cannot be nil")
	}
	return &EmailServiceImpl{
		db: db,
	}, nil
}

// SendEmail sends an email message based on a template
func (s *EmailServiceImpl) SendEmail(message *models.EmailMessage) error {
	logger := helper.Log.WithFields(map[string]interface{}{
		"template": message.Template,
		"provider": message.Provider,
		"refNo":    message.RefNo,
	})

	// Fetch the template
	var template models.Template
	if err := s.db.Where("uuid = ? AND channel = ?", message.Template, models.ChannelEmail).First(&template).Error; err != nil {
		logger.WithError(err).Error("Failed to find template")
		return fmt.Errorf("template not found: %w", err)
	}

	// Fetch the provider
	var provider models.Provider
	if err := s.db.Where("uuid = ? AND channel = ?", message.Provider, models.ChannelEmail).First(&provider).Error; err != nil {
		logger.WithError(err).Error("Failed to find provider")
		return fmt.Errorf("provider not found: %w", err)
	}

	// Create the email provider service
	emailProvider, err := email.NewSendGridProviderFromDB(&provider)
	if err != nil {
		logger.WithError(err).Error("Failed to create email provider")
		return fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Log template content and params for debugging
	logger.WithFields(map[string]interface{}{
		"templateContent": template.Content,
		"params":          message.Params,
	}).Debug("Processing template with params")

	// Process the template content with params
	processedContent, err := helper.ProcessTemplate(template.Content, message.Params)
	if err != nil {
		// Use debug template to get more info about the error
		debugInfo := helper.DebugTemplate(template.Content, message.Params)
		logger.WithError(err).WithField("debugInfo", debugInfo).Error("Failed to process template")
		return fmt.Errorf("failed to process template: %w", err)
	}

	// Process the subject with params if provided
	subject := message.Subject
	if subject == "" && message.Params != nil {
		subject = message.Params["subject"]
	}

	// Use template subject if still empty
	if subject == "" && template.Subject != "" {
		// Process the template subject with params
		subject, err = helper.ProcessTemplate(template.Subject, message.Params)
		if err != nil {
			logger.WithError(err).Warn("Failed to process template subject, using raw template subject")
			subject = template.Subject
		}
	}

	// Hardcode a default subject if still empty to satisfy SendGrid requirements
	if subject == "" {
		subject = template.Name
		if subject == "" {
			subject = "Notification"
		}
		logger.WithField("defaultSubject", subject).Info("Using default subject for email")
	}

	// Extract recipient emails
	recipients := make([]string, len(message.To))
	for i, recipient := range message.To {
		recipients[i] = recipient.Email
	}

	// Log detailed message information
	logger.WithFields(map[string]interface{}{
		"recipients":       recipients,
		"subject":          subject,
		"hasAttachments":   len(message.Attachments) > 0,
		"templateContent":  template.Content,
		"processedContent": processedContent,
		"providerConfig":   provider.Config,
	}).Info("Preparing to send email")

	// If there are attachments, send with attachments
	if len(message.Attachments) > 0 {
		attachments := make([]types.EmailAttachment, len(message.Attachments))
		for i, att := range message.Attachments {
			decodedContent, err := helper.DecodeBase64(att.Content)
			if err != nil {
				logger.WithError(err).Error("Failed to decode attachment content")
				return fmt.Errorf("failed to decode attachment: %w", err)
			}
			attachments[i] = types.EmailAttachment{
				Filename:    att.Filename,
				ContentType: att.ContentType,
				Content:     decodedContent,
			}
		}
		return emailProvider.SendWithAttachments(recipients, subject, processedContent, true, attachments)
	}

	// Send the email without attachments
	return emailProvider.Send(recipients, subject, processedContent, true)
}

// Send implements the EmailService Send method
func (s *EmailServiceImpl) Send(to []string, subject string, body string, isHTML bool) error {
	// This is a placeholder implementation that would typically use the default provider
	// For now, we'll return an error suggesting to use SendEmail instead
	return errors.New("direct Send method not implemented, use SendEmail instead")
}

// SendWithAttachments implements the EmailService SendWithAttachments method
func (s *EmailServiceImpl) SendWithAttachments(to []string, subject string, body string, isHTML bool, attachments []types.EmailAttachment) error {
	// This is a placeholder implementation that would typically use the default provider
	// For now, we'll return an error suggesting to use SendEmail instead
	return errors.New("direct SendWithAttachments method not implemented, use SendEmail instead")
}

// GetStatus implements the EmailService GetStatus method
func (s *EmailServiceImpl) GetStatus(messageID string) (types.DeliveryStatus, error) {
	// Query the database for the message status
	var message models.Message
	if err := s.db.Where("uuid = ? AND channel = ?", messageID, models.ChannelEmail).First(&message).Error; err != nil {
		return types.DeliveryStatus{}, fmt.Errorf("message not found: %w", err)
	}

	// Get the latest event for this message
	var event models.MessageEvent
	if err := s.db.Where("message_id = ?", messageID).Order("created_at DESC").First(&event).Error; err != nil {
		// If no events, just return the message status
		return types.DeliveryStatus{
			MessageID: messageID,
			Status:    string(message.Status),
			Timestamp: message.UpdatedAt.Format(time.RFC3339),
		}, nil
	}

	// Return the status based on the event
	return types.DeliveryStatus{
		MessageID: messageID,
		Status:    string(event.Status),
		Details:   event.Reason,
		Timestamp: event.Timestamp.Format(time.RFC3339),
	}, nil
}
