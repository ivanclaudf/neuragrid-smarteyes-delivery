package handler

import (
	"delivery/api"
	"delivery/helper"
	"delivery/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ProviderHandler handles provider management endpoints
type ProviderHandler struct {
	api *api.ProviderAPI
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(db *gorm.DB, readerDB *gorm.DB) *ProviderHandler {
	providerAPI, err := api.NewProviderAPI(db, readerDB)
	if err != nil {
		helper.Log.Errorf("Failed to create provider API: %v", err)
		return nil
	}

	return &ProviderHandler{
		api: providerAPI,
	}
}

// RegisterProviderRoutes registers all provider-related routes
func RegisterProviderRoutes(r *mux.Router, db *gorm.DB, readerDB *gorm.DB) {
	handler := NewProviderHandler(db, readerDB)
	if handler == nil {
		helper.Log.Error("Failed to create provider handler")
		return
	}

	// Provider management endpoints
	r.HandleFunc("/api/v1/providers", handler.CreateProviders).Methods("POST")
	r.HandleFunc("/api/v1/providers", handler.ListProviders).Methods("GET")
	r.HandleFunc("/api/v1/providers/{uuid}", handler.GetProvider).Methods("GET")
	r.HandleFunc("/api/v1/providers/{uuid}", handler.UpdateProvider).Methods("PUT")
}

// CreateProviders handles the creation of new providers
func (h *ProviderHandler) CreateProviders(w http.ResponseWriter, r *http.Request) {
	var request models.ProviderRequest
	if err := helper.ValidateRequestBody(r, &request); err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "CreateProviders",
			"error":   err.Error(),
		}).Warn("Bad request - invalid request body")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Invalid request body")
		return
	}

	// Call API to create providers
	response, err := h.api.CreateProviders(request)
	if err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "CreateProviders",
			"error":   err.Error(),
			"count":   len(request.Providers),
		}).Error("Failed to create providers")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler":       "CreateProviders",
		"providerCount": len(response.Providers),
	}).Info("Providers created successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusCreated, "Providers created successfully", response)
}

// UpdateProvider handles updating an existing provider
func (h *ProviderHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		helper.Log.WithFields(logrus.Fields{
			"handler": "UpdateProvider",
			"error":   "Missing UUID",
		}).Warn("Bad request - missing UUID")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Missing provider UUID")
		return
	}

	var request models.ProviderRequest
	if err := helper.ValidateRequestBody(r, &request); err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "UpdateProvider",
			"uuid":    uuid,
			"error":   err.Error(),
		}).Warn("Bad request - invalid request body")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Invalid request body")
		return
	}

	// Call API to update provider
	response, err := h.api.UpdateProvider(uuid, request)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "provider not found" {
			helper.Log.WithFields(logrus.Fields{
				"handler": "UpdateProvider",
				"uuid":    uuid,
				"error":   err.Error(),
			}).Warn("Provider not found")
			helper.RespondWithError(w, http.StatusNotFound, helper.CodeNotFound, "Provider not found")
			return
		}

		helper.Log.WithFields(logrus.Fields{
			"handler": "UpdateProvider",
			"uuid":    uuid,
			"error":   err.Error(),
		}).Error("Failed to update provider")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler": "UpdateProvider",
		"uuid":    uuid,
	}).Info("Provider updated successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusOK, "Provider updated successfully", response)
} // GetProvider retrieves a single provider by UUID
func (h *ProviderHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		helper.Log.WithFields(logrus.Fields{
			"handler": "GetProvider",
			"error":   "Missing UUID",
		}).Warn("Bad request - missing UUID")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Missing provider UUID")
		return
	}

	// Call API to get provider
	response, err := h.api.GetProvider(uuid)
	if err != nil {
		// Handle specific error cases
		if err.Error() == "provider not found" {
			helper.Log.WithFields(logrus.Fields{
				"handler": "GetProvider",
				"uuid":    uuid,
				"error":   err.Error(),
			}).Warn("Provider not found")
			helper.RespondWithError(w, http.StatusNotFound, helper.CodeNotFound, "Provider not found")
			return
		}

		helper.Log.WithFields(logrus.Fields{
			"handler": "GetProvider",
			"uuid":    uuid,
			"error":   err.Error(),
		}).Error("Failed to retrieve provider")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler": "GetProvider",
		"uuid":    uuid,
	}).Info("Provider retrieved successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusOK, "Provider retrieved successfully", response)
}

// ListProviders retrieves a list of providers with pagination
func (h *ProviderHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10 // Default limit
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0 // Default offset
	}

	channel := r.URL.Query().Get("channel")
	tenant := r.URL.Query().Get("tenant")

	helper.Log.WithFields(logrus.Fields{
		"handler": "ListProviders",
		"limit":   limit,
		"offset":  offset,
		"channel": channel,
		"tenant":  tenant,
	}).Debug("Listing providers with filters")

	// Call API to list providers
	response, total, err := h.api.ListProviders(limit, offset, channel, tenant)
	if err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "ListProviders",
			"limit":   limit,
			"offset":  offset,
			"channel": channel,
			"tenant":  tenant,
			"error":   err.Error(),
		}).Error("Failed to list providers")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Add pagination headers
	w.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
	w.Header().Set("X-Limit", strconv.Itoa(limit))
	w.Header().Set("X-Offset", strconv.Itoa(offset))

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler": "ListProviders",
		"count":   len(response.Providers),
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}).Info("Providers retrieved successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusOK, "Providers retrieved successfully", response)
}
