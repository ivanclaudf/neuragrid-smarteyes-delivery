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

// TemplateHandler handles template management endpoints
type TemplateHandler struct {
	api *api.TemplateAPI
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(db *gorm.DB, readerDB *gorm.DB) *TemplateHandler {
	templateAPI, err := api.NewTemplateAPI(db, readerDB)
	if err != nil {
		helper.Log.Errorf("Failed to create template API: %v", err)
		return nil
	}

	return &TemplateHandler{
		api: templateAPI,
	}
}

// RegisterTemplateRoutes registers all template-related routes
func RegisterTemplateRoutes(r *mux.Router, db *gorm.DB, readerDB *gorm.DB) {
	handler := NewTemplateHandler(db, readerDB)
	if handler == nil {
		helper.Log.Error("Failed to create template handler")
		return
	}

	// Template management endpoints
	r.HandleFunc("/api/v1/templates", handler.CreateTemplates).Methods("POST")
	r.HandleFunc("/api/v1/templates", handler.ListTemplates).Methods("GET")
	r.HandleFunc("/api/v1/templates/{uuid}", handler.GetTemplate).Methods("GET")
	r.HandleFunc("/api/v1/templates/{uuid}", handler.UpdateTemplate).Methods("PUT")
}

// CreateTemplates handles the creation of new templates
func (h *TemplateHandler) CreateTemplates(w http.ResponseWriter, r *http.Request) {
	var request models.TemplateRequest
	if err := helper.ValidateRequestBody(r, &request); err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "CreateTemplates",
			"error":   err.Error(),
		}).Warn("Bad request - invalid request body")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Invalid request body")
		return
	}

	// Call API to create templates
	response, err := h.api.CreateTemplates(request)
	if err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "CreateTemplates",
			"error":   err.Error(),
			"count":   len(request.Templates),
		}).Error("Failed to create templates")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler":       "CreateTemplates",
		"templateCount": len(response.Templates),
	}).Info("Templates created successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusCreated, "Templates created successfully", response)
}

// UpdateTemplate handles updating an existing template
func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		helper.Log.WithFields(logrus.Fields{
			"handler": "UpdateTemplate",
			"error":   "Missing UUID",
		}).Warn("Bad request - missing UUID")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Missing template UUID")
		return
	}

	var request models.TemplateRequest
	if err := helper.ValidateRequestBody(r, &request); err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "UpdateTemplate",
			"uuid":    uuid,
			"error":   err.Error(),
		}).Warn("Bad request - invalid request body")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Invalid request body")
		return
	}

	// Call API to update template
	response, err := h.api.UpdateTemplate(uuid, request)
	if err != nil {
		if err.Error() == "template not found" {
			helper.Log.WithFields(logrus.Fields{
				"handler": "UpdateTemplate",
				"uuid":    uuid,
				"error":   err.Error(),
			}).Warn("Template not found")
			helper.RespondWithError(w, http.StatusNotFound, helper.CodeNotFound, "Template not found")
			return
		}

		helper.Log.WithFields(logrus.Fields{
			"handler": "UpdateTemplate",
			"uuid":    uuid,
			"error":   err.Error(),
		}).Error("Failed to update template")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler": "UpdateTemplate",
		"uuid":    uuid,
	}).Info("Template updated successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusOK, "Template updated successfully", response)
}

// GetTemplate handles retrieving a template by UUID
func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		helper.Log.WithFields(logrus.Fields{
			"handler": "GetTemplate",
			"error":   "Missing UUID",
		}).Warn("Bad request - missing UUID")
		helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Missing template UUID")
		return
	}

	// Call API to get template
	response, err := h.api.GetTemplate(uuid)
	if err != nil {
		if err.Error() == "template not found" {
			helper.Log.WithFields(logrus.Fields{
				"handler": "GetTemplate",
				"uuid":    uuid,
				"error":   err.Error(),
			}).Warn("Template not found")
			helper.RespondWithError(w, http.StatusNotFound, helper.CodeNotFound, "Template not found")
			return
		}

		helper.Log.WithFields(logrus.Fields{
			"handler": "GetTemplate",
			"uuid":    uuid,
			"error":   err.Error(),
		}).Error("Failed to get template")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler": "GetTemplate",
		"uuid":    uuid,
	}).Info("Template retrieved successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusOK, "Template retrieved successfully", response)
}

// ListTemplates handles listing templates with optional filtering
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	var params models.TemplateListParams

	// Parse limit
	limitStr := query.Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			helper.Log.WithFields(logrus.Fields{
				"handler": "ListTemplates",
				"error":   err.Error(),
				"limit":   limitStr,
			}).Warn("Bad request - invalid limit")
			helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Invalid limit parameter")
			return
		}
		params.Limit = limit
	}

	// Parse offset
	offsetStr := query.Get("offset")
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			helper.Log.WithFields(logrus.Fields{
				"handler": "ListTemplates",
				"error":   err.Error(),
				"offset":  offsetStr,
			}).Warn("Bad request - invalid offset")
			helper.RespondWithError(w, http.StatusBadRequest, helper.CodeBadRequest, "Invalid offset parameter")
			return
		}
		params.Offset = offset
	}

	// Get other filters
	params.Channel = query.Get("channel")
	params.Tenant = query.Get("tenant")
	params.Code = query.Get("code")

	// Call API to list templates
	response, err := h.api.ListTemplates(params)
	if err != nil {
		helper.Log.WithFields(logrus.Fields{
			"handler": "ListTemplates",
			"error":   err.Error(),
			"params":  params,
		}).Error("Failed to list templates")
		helper.RespondWithError(w, http.StatusInternalServerError, helper.CodeServerError, helper.MsgServerError)
		return
	}

	// Log success
	helper.Log.WithFields(logrus.Fields{
		"handler":       "ListTemplates",
		"templateCount": len(response.Templates),
	}).Info("Templates listed successfully")

	// Return success response without data wrapper
	helper.RespondWithSuccessNoDataWrapper(w, http.StatusOK, "Templates retrieved successfully", response)
}
