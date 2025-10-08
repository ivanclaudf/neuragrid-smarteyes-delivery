package handler

import (
	"delivery/api"
	"delivery/helper"
	"delivery/models"
	"delivery/services/queue"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// SMSHandler handles SMS endpoints
type SMSHandler struct {
	api          *api.SMSAPI
	pulsarClient *queue.PulsarClient
	db           *gorm.DB
	readerDB     *gorm.DB
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) (*SMSHandler, error) {
	return &SMSHandler{
		pulsarClient: pulsarClient,
		db:           db,
		readerDB:     readerDB,
	}, nil
}

// RegisterSMSRoutes registers all SMS-related routes
func RegisterSMSRoutes(r *mux.Router, db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) {
	handler, err := NewSMSHandler(db, readerDB, pulsarClient)
	if err != nil {
		return
	}

	// Initialize SMS API
	smsAPI, err := api.NewSMSAPI(db, readerDB, pulsarClient)
	if err != nil {
		helper.Log.Errorf("Failed to create SMS API: %v", err)
		return
	}

	handler.api = smsAPI

	// Combined SMS endpoint for messages
	r.HandleFunc("/api/v1/sms", handler.HandleSMSRequest).Methods("POST")
}

// HandleSMSRequest handles the SMS request
func (h *SMSHandler) HandleSMSRequest(w http.ResponseWriter, r *http.Request) {
	var request models.SMSRequest

	if err := helper.ValidateRequestBody(r, &request); err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, helper.MsgInvalidRequestBody)
		return
	}

	// Use the API layer to process the request
	responses, err := h.api.ProcessMessageBatch(request)
	if err != nil {
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, err.Error())
		return
	}

	// Wrap responses in a "messages" object
	responseWrapper := models.SMSResponse{
		Messages: responses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	helper.WriteJSON(w, responseWrapper)
}
