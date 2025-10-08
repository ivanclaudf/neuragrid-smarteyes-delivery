package services

// SMSService defines operations for sending SMS messages
type SMSService interface {
	// Send an SMS to a recipient
	Send(to string, message string) error

	// Send an SMS to multiple recipients
	SendBulk(to []string, message string) error

	// SendTemplate sends a template message to a recipient
	// Currently implemented by rendering template on server side and using Send
	// Future implementations may use provider-specific template APIs
	SendTemplate(to string, templateName string, params map[string]string) error

	// Get SMS delivery status by message ID
	GetStatus(messageID string) (DeliveryStatus, error)
}
