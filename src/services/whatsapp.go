package services

import (
	"delivery/api/types"
)

// WhatsAppService defines operations for sending WhatsApp messages
type WhatsAppService interface {
	// Send a text message to a recipient
	SendText(to string, message string) error

	// Send a media message to a recipient
	SendMedia(to string, caption string, mediaType string, mediaURL string) error

	// Send a template message to a recipient
	// templateName is the provider's template ID, content is the rendered template content,
	// and params are the variables to replace in the template
	SendTemplate(to string, templateName string, params map[string]string) error

	// Get WhatsApp message delivery status by message ID
	GetStatus(messageID string) (types.DeliveryStatus, error)
}
