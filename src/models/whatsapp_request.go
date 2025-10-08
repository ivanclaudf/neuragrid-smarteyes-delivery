package models

// WhatsAppRequest represents the request body for sending WhatsApp messages
type WhatsAppRequest struct {
	Messages []WhatsAppMessage `json:"messages" validate:"required,min=1"`
}

// WhatsAppMessage represents a single WhatsApp message
type WhatsAppMessage struct {
	Template    string               `json:"template" validate:"required"`
	To          []WhatsAppRecipient  `json:"to" validate:"required,min=1"`
	Provider    string               `json:"provider" validate:"required,uuid4"`
	RefNo       string               `json:"refno" validate:"required"`
	Categories  []string             `json:"categories" validate:"required,min=1"`
	Identifiers WhatsAppIdentifiers  `json:"identifiers" validate:"required"`
	Params      map[string]string    `json:"params"`
	Attachments *WhatsAppAttachments `json:"attachments"`
}

// WhatsAppRecipient represents a recipient for a WhatsApp message
type WhatsAppRecipient struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone" validate:"required,e164"`
}

// WhatsAppIdentifiers represents identifiers for a WhatsApp message
type WhatsAppIdentifiers struct {
	Tenant     string `json:"tenant" validate:"required"`
	EventUUID  string `json:"eventUuid" validate:"omitempty,uuid4"`
	ActionUUID string `json:"actionUuid" validate:"omitempty,uuid4"`
	ActionCode string `json:"actionCode"`
}

// WhatsAppAttachments represents attachments for a WhatsApp message
type WhatsAppAttachments struct {
	Inline []WhatsAppInlineAttachment `json:"inline"`
}

// WhatsAppInlineAttachment represents an inline attachment for a WhatsApp message
type WhatsAppInlineAttachment struct {
	Filename  string `json:"filename" validate:"required"`
	Type      string `json:"type" validate:"required"`
	Content   string `json:"content" validate:"required"`
	ContentID string `json:"contentId" validate:"required"`
}

// WhatsAppResponse represents the response body for sending WhatsApp messages
type WhatsAppResponse struct {
	Messages []WhatsAppMessageResponse `json:"messages"`
}

// WhatsAppMessageResponse represents a single WhatsApp message response
type WhatsAppMessageResponse struct {
	RefNo string `json:"refno"`
	UUID  string `json:"uuid"`
}
