package handler

import (
	"delivery/helper"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// SMSHandler handles SMS operations
type SMSHandler struct {
	DB *gorm.DB
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(db *gorm.DB) (*SMSHandler, error) {
	if db == nil {
		return nil, nil
	}
	return &SMSHandler{
		DB: db,
	}, nil
}

// RegisterSMSRoutes registers all SMS-related routes
func RegisterSMSRoutes(r *mux.Router, db *gorm.DB) {
	handler, err := NewSMSHandler(db)
	if err != nil {
		return
	}

	r.HandleFunc("/api/v1/sms", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleSMSRequest(w, r)
	}).Methods("POST")
}

// HandleSMSRequest handles SMS requests
func (h *SMSHandler) HandleSMSRequest(w http.ResponseWriter, r *http.Request) {
	// Placeholder for SMS handling logic
	helper.RespondWithError(w, http.StatusNotImplemented, helper.CodeError, "SMS sending not yet implemented")
}
