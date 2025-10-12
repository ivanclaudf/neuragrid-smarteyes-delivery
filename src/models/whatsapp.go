package models

// WhatsAppMessage represents a single WhatsApp message in the internal system
type WhatsAppMessage struct {
	Template    string               `json:"template"`
	To          []WhatsAppRecipient  `json:"to"`
	Provider    string               `json:"provider"`
	RefNo       string               `json:"refno"`
	Categories  []string             `json:"categories"`
	Identifiers WhatsAppIdentifiers  `json:"identifiers"`
	Params      map[string]string    `json:"params"`
	Attachments *WhatsAppAttachments `json:"attachments"`
}

// WhatsAppRecipient represents a recipient for a WhatsApp message
type WhatsAppRecipient struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
}

// WhatsAppIdentifiers represents identifiers for a WhatsApp message
type WhatsAppIdentifiers struct {
	Tenant     string `json:"tenant"`
	EventUUID  string `json:"eventUuid"`
	ActionUUID string `json:"actionUuid"`
	ActionCode string `json:"actionCode"`
}

// WhatsAppAttachments represents attachments for a WhatsApp message
type WhatsAppAttachments struct {
	Inline []WhatsAppInlineAttachment `json:"inline"`
}

// WhatsAppInlineAttachment represents an inline attachment for a WhatsApp message
type WhatsAppInlineAttachment struct {
	Filename  string `json:"filename"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	ContentID string `json:"contentId"`
}
