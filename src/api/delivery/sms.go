package api

import (
	"delivery/helper"
	"delivery/services/providers/sms"
	"net/http"
)

// SMSRequest represents a request to send an SMS
type SMSRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// SMSBulkRequest represents a request to send bulk SMS messages
type SMSBulkRequest struct {
	To      []string `json:"to"`
	Message string   `json:"message"`
}

// SMSHandler processes SMS related requests
type SMSHandler struct {
	Provider *sms.TwilioProvider
}

// NewSMSHandler creates a new SMS handler with the specified provider
func NewSMSHandler() (*SMSHandler, error) {
	provider, err := sms.NewTwilioProvider()
	if err != nil {
		return nil, err
	}

	return &SMSHandler{
		Provider: provider,
	}, nil
}

// HandleSMSRequest processes SMS send requests (both single and bulk)
func (h *SMSHandler) HandleSMSRequest(w http.ResponseWriter, r *http.Request) {
	// Try to parse as a bulk request first
	var bulkReq SMSBulkRequest
	if err := helper.ValidateRequestBody(r, &bulkReq); err == nil && len(bulkReq.To) > 0 {
		// This is a bulk request
		if err := h.Provider.SendBulk(bulkReq.To, bulkReq.Message); err != nil {
			helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeError, err.Error())
			return
		}

		helper.RespondWithJSON(w, http.StatusOK, helper.Response{
			Code:    helper.CodeSuccess,
			Message: "Bulk SMS sent successfully",
		})
		return
	}

	// If not a bulk request, try as a single SMS
	var req SMSRequest
	if err := helper.ValidateRequestBody(r, &req); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, helper.MsgInvalidRequestBody)
		return
	}

	if err := h.Provider.Send(req.To, req.Message); err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeError, err.Error())
		return
	}

	helper.RespondWithJSON(w, http.StatusOK, helper.Response{
		Code:    helper.CodeSuccess,
		Message: "SMS sent successfully",
	})
}

// HandleStatusRequest processes SMS status requests
func (h *SMSHandler) HandleStatusRequest(w http.ResponseWriter, r *http.Request) {
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
