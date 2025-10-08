package services

// SMSService defines operations for sending SMS messages
type SMSService interface {
	// Send an SMS to a recipient
	Send(to string, message string) error

	// Send an SMS to multiple recipients
	SendBulk(to []string, message string) error

	// Get SMS delivery status by message ID
	GetStatus(messageID string) (DeliveryStatus, error)
}
