package email

import (
	"delivery/services"
	"errors"
	"os"
)

// SendGridProvider implements the EmailService interface using SendGrid
type SendGridProvider struct {
	apiKey string
	from   string
}

// NewSendGridProvider creates a new SendGrid email provider
func NewSendGridProvider() (*SendGridProvider, error) {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("SENDGRID_FROM_EMAIL")

	if apiKey == "" {
		return nil, errors.New("SENDGRID_API_KEY environment variable not set")
	}

	if from == "" {
		return nil, errors.New("SENDGRID_FROM_EMAIL environment variable not set")
	}

	return &SendGridProvider{
		apiKey: apiKey,
		from:   from,
	}, nil
}

// Send implements the EmailService.Send method
func (p *SendGridProvider) Send(to []string, subject string, body string, isHTML bool) error {
	// TODO: Implement SendGrid API call
	return nil
}

// SendWithAttachments implements the EmailService.SendWithAttachments method
func (p *SendGridProvider) SendWithAttachments(to []string, subject string, body string, isHTML bool, attachments []services.Attachment) error {
	// TODO: Implement SendGrid API call with attachments
	return nil
}

// GetStatus implements the EmailService.GetStatus method
func (p *SendGridProvider) GetStatus(messageID string) (services.DeliveryStatus, error) {
	// TODO: Implement SendGrid status check
	return services.DeliveryStatus{
		MessageID: messageID,
		Status:    "unknown",
		Details:   "Status check not implemented yet",
		Timestamp: "",
	}, nil
}
