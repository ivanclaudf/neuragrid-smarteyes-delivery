package types

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Content     []byte `json:"content"`
}
