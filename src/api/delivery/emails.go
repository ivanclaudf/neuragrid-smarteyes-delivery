package api

import (
	"delivery/helper"
	"delivery/services"
	"delivery/services/providers/email"
	"net/http"
)

// EmailRequest represents a request to send an email
type EmailRequest struct {
	To          []string              `json:"to"`
	Subject     string                `json:"subject"`
	Body        string                `json:"body"`
	IsHTML      bool                  `json:"isHtml"`
	Attachments []services.Attachment `json:"attachments,omitempty"`
}

// StatusRequest represents a request to get the status of a message
type StatusRequest struct {
	MessageID string `json:"messageId"`
}

// EmailHandler processes email related requests
type EmailHandler struct {
	Provider *email.SendGridProvider
}

// NewEmailHandler creates a new email handler with the specified provider
func NewEmailHandler() (*EmailHandler, error) {
	provider, err := email.NewSendGridProvider()
	if err != nil {
		return nil, err
	}

	return &EmailHandler{
		Provider: provider,
	}, nil
}

// HandleSendRequest processes email send requests
func (h *EmailHandler) HandleSendRequest(w http.ResponseWriter, r *http.Request) {
	var req EmailRequest
	if err := helper.ValidateRequestBody(r, &req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, helper.MsgInvalidRequestBody)
		return
	}

	var err error
	if len(req.Attachments) > 0 {
		err = h.Provider.SendWithAttachments(req.To, req.Subject, req.Body, req.IsHTML, req.Attachments)
	} else {
		err = h.Provider.Send(req.To, req.Subject, req.Body, req.IsHTML)
	}

	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeError, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, helper.Response{
		Code:    helper.CodeSuccess,
		Message: "Email sent successfully",
	})
}

// HandleStatusRequest processes email status requests
func (h *EmailHandler) HandleStatusRequest(w http.ResponseWriter, r *http.Request) {
	var req StatusRequest
	if err := helper.ValidateRequestBody(r, &req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, helper.MsgInvalidRequestBody)
		return
	}

	status, err := h.Provider.GetStatus(req.MessageID)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeError, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, helper.Response{
		Code:    helper.CodeSuccess,
		Message: "Status retrieved successfully",
		Data:    status,
	})
}
