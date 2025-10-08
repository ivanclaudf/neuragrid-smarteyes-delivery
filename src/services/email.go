package services

// EmailService defines operations for sending emails
type EmailService interface {
	// Send an email to a list of recipients
	Send(to []string, subject string, body string, isHTML bool) error

	// Send an email with attachments to a list of recipients
	SendWithAttachments(to []string, subject string, body string, isHTML bool, attachments []Attachment) error

	// Get email delivery status by message ID
	GetStatus(messageID string) (DeliveryStatus, error)
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
}

// DeliveryStatus represents the status of a message delivery
type DeliveryStatus struct {
	MessageID string
	Status    string // delivered, failed, pending, etc.
	Details   string // Error details or delivery confirmation
	Timestamp string // When the status was last updated
}
