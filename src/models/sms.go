package models

// SMSMessage represents a single SMS message in the internal system
type SMSMessage struct {
	To          []SMSRecipient         `json:"to"`
	From        string                 `json:"from"`
	Body        string                 `json:"body"`
	Template    string                 `json:"template"`
	Provider    string                 `json:"provider"`
	RefNo       string                 `json:"refno"`
	Categories  []string               `json:"categories"`
	Identifiers map[string]interface{} `json:"identifiers"`
	Params      map[string]string      `json:"params"`
	TenantID    string                 `json:"tenantId"`
}

// SMSRecipient represents a recipient for an SMS message
type SMSRecipient struct {
	Telephone string `json:"telephone"`
}

// SMSIdentifiers represents identifiers for an SMS message
