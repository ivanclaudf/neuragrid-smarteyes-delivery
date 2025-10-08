package handler

import (
	"delivery/helper"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// EmailHandler handles email operations
type EmailHandler struct {
	DB *gorm.DB
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(db *gorm.DB) (*EmailHandler, error) {
	if db == nil {
		return nil, nil
	}
	return &EmailHandler{
		DB: db,
	}, nil
}

// RegisterEmailRoutes registers all email-related routes
func RegisterEmailRoutes(r *mux.Router, db *gorm.DB) {
	handler, err := NewEmailHandler(db)
	if err != nil {
		return
	}

	r.HandleFunc("/api/v1/email", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleEmailRequest(w, r)
	}).Methods("POST")
}

// HandleEmailRequest handles email requests
func (h *EmailHandler) HandleEmailRequest(w http.ResponseWriter, r *http.Request) {
	// Placeholder for email handling logic
	helper.RespondWithError(w, http.StatusNotImplemented, helper.CodeError, "Email sending not yet implemented")
}
