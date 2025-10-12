package models

// SMSMessage represents a single SMS message in the internal system
type SMSMessage struct {
	To          []SMSRecipient    `json:"to"`
	From        string            `json:"from"`
	Body        string            `json:"body"`
	Template    string            `json:"template"`
	Provider    string            `json:"provider"`
	RefNo       string            `json:"refno"`
	Categories  []string          `json:"categories"`
	Identifiers SMSIdentifiers    `json:"identifiers"`
	Params      map[string]string `json:"params"`
}

// SMSRecipient represents a recipient for an SMS message
type SMSRecipient struct {
	Telephone string `json:"telephone"`
}

// SMSIdentifiers represents identifiers for an SMS message
type SMSIdentifiers struct {
	Tenant     string `json:"tenant"`
	EventUUID  string `json:"eventUuid"`
	ActionUUID string `json:"actionUuid"`
	ActionCode string `json:"actionCode"`
}
