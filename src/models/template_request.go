package models

// TemplateRequest represents the request body for template management APIs
type TemplateRequest struct {
	Templates []TemplateRequestItem `json:"templates" binding:"required,min=1"`
}

// TemplateRequestItem represents a single template in the template request
type TemplateRequestItem struct {
	UUID        string `json:"uuid,omitempty"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Channel     string `json:"channel" binding:"required"`
	TemplateIds JSON   `json:"templateIds"`
	Tenant      string `json:"tenant" binding:"required"`
	Status      *int   `json:"status,omitempty"`
}

// TemplateResponse represents the response body for template management APIs
type TemplateResponse struct {
	Templates []TemplateResponseItem `json:"templates"`
}

// TemplateResponseItem represents a single template in the template response
type TemplateResponseItem struct {
	UUID        string `json:"uuid"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Content     string `json:"content"`
	Channel     string `json:"channel"`
	TemplateIds JSON   `json:"templateIds"`
	Tenant      string `json:"tenant"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// TemplateListParams represents parameters for listing templates
type TemplateListParams struct {
	Limit   int    `json:"limit" form:"limit"`
	Offset  int    `json:"offset" form:"offset"`
	Channel string `json:"channel" form:"channel"`
	Tenant  string `json:"tenant" form:"tenant"`
	Code    string `json:"code" form:"code"`
}
