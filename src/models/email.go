package models

// EmailMessage represents a single email message in the internal system
type EmailMessage struct {
	Template    string                 `json:"template"`
	To          []EmailRecipient       `json:"to"`
	Provider    string                 `json:"provider"`
	RefNo       string                 `json:"refno"`
	Categories  []string               `json:"categories"`
	Identifiers map[string]interface{} `json:"identifiers"`
	Params      map[string]string      `json:"params"`
	Subject     string                 `json:"subject,omitempty"`
	Attachments []AttachmentMetadata   `json:"attachments,omitempty"`
	TenantID    string                 `json:"tenantId"`
}

// EmailRecipient represents an email recipient with name and email
type EmailRecipient struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// EmailIdentifiers represents identifiers for an email message

// AttachmentMetadata represents metadata for an email attachment
type AttachmentMetadata struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"` // base64 encoded
}
