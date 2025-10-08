package models

// ProviderRequest represents the request body for provider management APIs
type ProviderRequest struct {
	Providers []ProviderRequestItem `json:"providers" binding:"required,min=1"`
}

// ProviderRequestItem represents a single provider in the provider request
type ProviderRequestItem struct {
	UUID         string `json:"uuid,omitempty"`
	Code         string `json:"code" binding:"required"`
	Provider     string `json:"provider" binding:"required"` // Implementation class name (e.g. twilio)
	Name         string `json:"name" binding:"required"`
	Config       JSON   `json:"config" binding:"required"`
	SecureConfig JSON   `json:"secureConfig" binding:"required"`
	Channel      string `json:"channel" binding:"required"`
	Tenant       string `json:"tenant" binding:"required"`
	Status       *int   `json:"status,omitempty"`
}

// ProviderResponse represents the response body for provider management APIs
type ProviderResponse struct {
	Providers []ProviderResponseItem `json:"providers"`
}

// ProviderResponseItem represents a single provider in the provider response
type ProviderResponseItem struct {
	UUID      string `json:"uuid"`
	Code      string `json:"code"`
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	Config    JSON   `json:"config"`
	Channel   string `json:"channel"`
	Tenant    string `json:"tenant"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ProviderListParams represents parameters for listing providers
type ProviderListParams struct {
	Limit   int    `json:"limit" form:"limit"`
	Offset  int    `json:"offset" form:"offset"`
	Channel string `json:"channel" form:"channel"`
	Tenant  string `json:"tenant" form:"tenant"`
}
