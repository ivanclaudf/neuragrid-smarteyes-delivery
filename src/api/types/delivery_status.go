package types

// DeliveryStatus represents the status of a message delivery
type DeliveryStatus struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`    // delivered, failed, pending, etc.
	Details   string `json:"details"`   // Error details or delivery confirmation
	Timestamp string `json:"timestamp"` // When the status was last updated
}
