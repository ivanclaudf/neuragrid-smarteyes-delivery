package models

// SMSRequest represents the request body for sending SMS messages
type SMSRequest struct {
	Messages []SMSMessage `json:"messages" validate:"required,min=1"`
}

// SMSMessage represents a single SMS message
type SMSMessage struct {
	To          []SMSRecipient    `json:"to" validate:"required,min=1"`
	From        string            `json:"from" validate:"required"`
	Body        string            `json:"body"` // For backward compatibility, not required when using template
	Template    string            `json:"template" validate:"required"`
	Provider    string            `json:"provider" validate:"required,uuid4"`
	RefNo       string            `json:"refno" validate:"required"`
	Categories  []string          `json:"categories" validate:"required,min=1"`
	Identifiers SMSIdentifiers    `json:"identifiers" validate:"required"`
	Params      map[string]string `json:"params"`
}

// SMSRecipient represents a recipient for an SMS message
type SMSRecipient struct {
	Telephone string `json:"telephone" validate:"required,e164"`
}

// SMSIdentifiers represents identifiers for an SMS message
type SMSIdentifiers struct {
	Tenant     string `json:"tenant" validate:"required"`
	EventUUID  string `json:"eventUuid" validate:"omitempty,uuid4"`
	ActionUUID string `json:"actionUuid" validate:"omitempty,uuid4"`
	ActionCode string `json:"actionCode"`
}

// SMSResponse represents the response body for sending SMS messages
type SMSResponse struct {
	Messages []SMSMessageResponse `json:"messages"`
}

// SMSMessageResponse represents a response for a single SMS message
type SMSMessageResponse struct {
	RefNo string `json:"refno"`
	UUID  string `json:"uuid"`
}
