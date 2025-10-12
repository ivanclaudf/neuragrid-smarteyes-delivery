package handler

import (
	"delivery/api"
	"delivery/helper"
	"delivery/services/queue"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// EmailHandler handles email operations
type EmailHandler struct {
	api          *api.EmailAPI
	pulsarClient *queue.PulsarClient
	db           *gorm.DB
	readerDB     *gorm.DB
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) (*EmailHandler, error) {
	return &EmailHandler{
		pulsarClient: pulsarClient,
		db:           db,
		readerDB:     readerDB,
	}, nil
}

// RegisterEmailRoutes registers all email-related routes
func RegisterEmailRoutes(r *mux.Router, db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) {
	handler, err := NewEmailHandler(db, readerDB, pulsarClient)
	if err != nil {
		return
	}

	// Initialize Email API
	emailAPI, err := api.NewEmailAPI(db, readerDB, pulsarClient)
	if err != nil {
		helper.Log.Errorf("Failed to create Email API: %v", err)
		return
	}

	handler.api = emailAPI

	// Combined Email endpoint for template messages
	r.HandleFunc("/api/v1/email", handler.HandleEmailRequest).Methods("POST")
}

// HandleEmailRequest handles the Email request
func (h *EmailHandler) HandleEmailRequest(w http.ResponseWriter, r *http.Request) {
	var request api.EmailRequest

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
	responseWrapper := api.EmailResponse{
		Messages: responses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	helper.WriteJSON(w, responseWrapper)
}
