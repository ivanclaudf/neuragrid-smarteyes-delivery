package services

import (
	"delivery/api/types"
)

// EmailService defines operations for sending emails
type EmailService interface {
	// Send an email to a list of recipients
	Send(to []string, subject string, body string, isHTML bool) error

	// Send an email with attachments to a list of recipients
	SendWithAttachments(to []string, subject string, body string, isHTML bool, attachments []types.EmailAttachment) error

	// Get email delivery status by message ID
	GetStatus(messageID string) (types.DeliveryStatus, error)
}

// Note: Attachment and DeliveryStatus types are now defined in delivery/services/providers/email
