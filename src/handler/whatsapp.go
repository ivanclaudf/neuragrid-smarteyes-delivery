package handler

import (
	"delivery/api"
	"delivery/helper"
	"delivery/services/queue"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// WhatsAppHandler handles WhatsApp endpoints
type WhatsAppHandler struct {
	api          *api.WhatsAppAPI
	pulsarClient *queue.PulsarClient
	db           *gorm.DB
	readerDB     *gorm.DB
}

// NewWhatsAppHandler creates a new WhatsApp handler
func NewWhatsAppHandler(db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) (*WhatsAppHandler, error) {
	return &WhatsAppHandler{
		pulsarClient: pulsarClient,
		db:           db,
		readerDB:     readerDB,
	}, nil
}

// RegisterWhatsAppRoutes registers all WhatsApp-related routes
func RegisterWhatsAppRoutes(r *mux.Router, db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) {
	handler, err := NewWhatsAppHandler(db, readerDB, pulsarClient)
	if err != nil {
		return
	}

	// Initialize WhatsApp API
	whatsAppAPI, err := api.NewWhatsAppAPI(db, readerDB, pulsarClient)
	if err != nil {
		helper.Log.Errorf("Failed to create WhatsApp API: %v", err)
		return
	}

	handler.api = whatsAppAPI

	// Combined WhatsApp endpoint for template messages
	r.HandleFunc("/api/v1/whatsapp", handler.HandleWhatsAppRequest).Methods("POST")
}

// HandleWhatsAppRequest handles the WhatsApp request
func (h *WhatsAppHandler) HandleWhatsAppRequest(w http.ResponseWriter, r *http.Request) {
	var request api.WhatsAppRequest

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
	responseWrapper := api.WhatsAppResponse{
		Messages: responses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	helper.WriteJSON(w, responseWrapper)
}
