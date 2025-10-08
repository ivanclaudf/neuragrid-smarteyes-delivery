package api

import (
	"delivery/helper"
	"delivery/models"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TemplateAPI handles template business logic
type TemplateAPI struct {
	DB       *gorm.DB
	ReaderDB *gorm.DB
}

// NewTemplateAPI creates a new template API
func NewTemplateAPI(db *gorm.DB, readerDB *gorm.DB) (*TemplateAPI, error) {
	logger := helper.Log.WithField("component", "TemplateAPI")

	if db == nil {
		logger.Error("Writer database connection is nil")
		return nil, fmt.Errorf("writer database connection is nil")
	}
	if readerDB == nil {
		logger.Error("Reader database connection is nil")
		return nil, fmt.Errorf("reader database connection is nil")
	}

	logger.Info("Template API initialized successfully")
	return &TemplateAPI{
		DB:       db,
		ReaderDB: readerDB,
	}, nil
}

// CreateTemplates creates new templates
func (a *TemplateAPI) CreateTemplates(request models.TemplateRequest) (*models.TemplateResponse, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "TemplateAPI",
		"method":    "CreateTemplates",
		"count":     len(request.Templates),
	})

	logger.Info("Creating new templates")

	response := &models.TemplateResponse{
		Templates: make([]models.TemplateResponseItem, 0, len(request.Templates)),
	}

	for idx, templateItem := range request.Templates {
		templateLogger := logger.WithFields(logrus.Fields{
			"index":   idx,
			"name":    templateItem.Name,
			"channel": templateItem.Channel,
			"tenant":  templateItem.Tenant,
		})

		templateLogger.Debug("Processing template item")

		// Generate UUID for the template
		uuid, err := helper.GenerateUUID()
		if err != nil {
			templateLogger.WithError(err).Error("Failed to generate UUID for template")
			return nil, fmt.Errorf("failed to generate UUID: %v", err)
		}

		// Create template object using DB defaults
		template := models.Template{
			UUID:        uuid,
			Name:        templateItem.Name,
			Content:     templateItem.Content,
			Channel:     models.Channel(templateItem.Channel),
			TemplateIds: templateItem.TemplateIds,
			Tenant:      templateItem.Tenant,
		}

		// Set status if provided (otherwise DB default will be used)
		if templateItem.Status != nil {
			template.Status = *templateItem.Status
		}

		// Save template to database
		if err := a.DB.Create(&template).Error; err != nil {
			templateLogger.WithError(err).Error("Failed to create template in database")
			return nil, fmt.Errorf("failed to create template: %v", err)
		}

		templateLogger.WithField("uuid", template.UUID).Info("Template created successfully")

		// Add to response
		responseItem := models.TemplateResponseItem{
			UUID:        template.UUID,
			Name:        template.Name,
			Content:     template.Content,
			Channel:     string(template.Channel),
			TemplateIds: template.TemplateIds,
			Tenant:      template.Tenant,
			Status:      template.Status,
			CreatedAt:   template.CreatedAt.Format(helper.TimeFormat),
			UpdatedAt:   template.UpdatedAt.Format(helper.TimeFormat),
		}

		response.Templates = append(response.Templates, responseItem)
	}

	logger.WithField("created_count", len(response.Templates)).Info("Templates created successfully")
	return response, nil
}

// UpdateTemplate updates an existing template
func (a *TemplateAPI) UpdateTemplate(uuid string, request models.TemplateRequest) (*models.TemplateResponse, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "TemplateAPI",
		"method":    "UpdateTemplate",
		"uuid":      uuid,
	})

	logger.Info("Updating template")

	if uuid == "" {
		logger.Error("Missing template UUID")
		return nil, fmt.Errorf("missing template UUID")
	}

	if len(request.Templates) != 1 {
		logger.WithField("count", len(request.Templates)).Error("Update requires exactly one template")
		return nil, fmt.Errorf("update requires exactly one template")
	}

	templateItem := request.Templates[0]
	logger = logger.WithFields(logrus.Fields{
		"name":    templateItem.Name,
		"channel": templateItem.Channel,
		"tenant":  templateItem.Tenant,
	})

	// Get existing template
	var template models.Template
	if err := a.DB.Where("uuid = ?", uuid).First(&template).Error; err != nil {
		logger.WithError(err).Error("Template not found")
		return nil, fmt.Errorf("template not found: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"existing_name": template.Name,
	}).Debug("Found existing template")

	// Update only the fields that are provided
	updates := make(map[string]interface{})

	if templateItem.Name != "" {
		updates["name"] = templateItem.Name
	}

	if templateItem.Content != "" {
		updates["content"] = templateItem.Content
	}

	if templateItem.TemplateIds != nil {
		updates["template_ids"] = templateItem.TemplateIds
	}

	if templateItem.Status != nil {
		updates["status"] = *templateItem.Status
	}

	// Apply updates if there are any
	if len(updates) > 0 {
		if err := a.DB.Model(&template).Updates(updates).Error; err != nil {
			logger.WithError(err).Error("Failed to update template")
			return nil, fmt.Errorf("failed to update template: %v", err)
		}
		// Refresh the template
		if err := a.DB.Where("uuid = ?", uuid).First(&template).Error; err != nil {
			logger.WithError(err).Error("Failed to retrieve updated template")
			return nil, fmt.Errorf("failed to retrieve updated template: %v", err)
		}
	} else {
		logger.Debug("No updates to apply")
	}

	// Create response
	response := &models.TemplateResponse{
		Templates: []models.TemplateResponseItem{
			{
				UUID:        template.UUID,
				Name:        template.Name,
				Content:     template.Content,
				Channel:     string(template.Channel),
				TemplateIds: template.TemplateIds,
				Tenant:      template.Tenant,
				Status:      template.Status,
				CreatedAt:   template.CreatedAt.Format(helper.TimeFormat),
				UpdatedAt:   template.UpdatedAt.Format(helper.TimeFormat),
			},
		},
	}

	logger.Info("Template updated successfully")
	return response, nil
}

// GetTemplate retrieves a template by UUID
func (a *TemplateAPI) GetTemplate(uuid string) (*models.TemplateResponse, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "TemplateAPI",
		"method":    "GetTemplate",
		"uuid":      uuid,
	})

	logger.Info("Retrieving template")

	if uuid == "" {
		logger.Error("Missing template UUID")
		return nil, fmt.Errorf("missing template UUID")
	}

	// Get template
	var template models.Template
	if err := a.ReaderDB.Where("uuid = ?", uuid).First(&template).Error; err != nil {
		logger.WithError(err).Error("Template not found")
		return nil, fmt.Errorf("template not found: %v", err)
	}

	// Create response
	response := &models.TemplateResponse{
		Templates: []models.TemplateResponseItem{
			{
				UUID:        template.UUID,
				Name:        template.Name,
				Content:     template.Content,
				Channel:     string(template.Channel),
				TemplateIds: template.TemplateIds,
				Tenant:      template.Tenant,
				Status:      template.Status,
				CreatedAt:   template.CreatedAt.Format(helper.TimeFormat),
				UpdatedAt:   template.UpdatedAt.Format(helper.TimeFormat),
			},
		},
	}

	logger.Info("Template retrieved successfully")
	return response, nil
}

// ListTemplates lists templates with optional filtering
func (a *TemplateAPI) ListTemplates(params models.TemplateListParams) (*models.TemplateResponse, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "TemplateAPI",
		"method":    "ListTemplates",
		"limit":     params.Limit,
		"offset":    params.Offset,
		"channel":   params.Channel,
		"tenant":    params.Tenant,
	})

	logger.Info("Listing templates")

	// Set default limit
	if params.Limit <= 0 {
		params.Limit = 50
	}

	// Build query
	query := a.ReaderDB.Model(&models.Template{})

	// Apply filters
	if params.Channel != "" {
		query = query.Where("channel = ?", params.Channel)
	}
	if params.Tenant != "" {
		query = query.Where("tenant = ?", params.Tenant)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count templates")
		return nil, fmt.Errorf("failed to count templates: %v", err)
	}

	// Get templates with pagination
	var templates []models.Template
	if err := query.Limit(params.Limit).Offset(params.Offset).Find(&templates).Error; err != nil {
		logger.WithError(err).Error("Failed to retrieve templates")
		return nil, fmt.Errorf("failed to retrieve templates: %v", err)
	}

	// Build response
	response := &models.TemplateResponse{
		Templates: make([]models.TemplateResponseItem, 0, len(templates)),
	}

	for _, template := range templates {
		responseItem := models.TemplateResponseItem{
			UUID:        template.UUID,
			Name:        template.Name,
			Content:     template.Content,
			Channel:     string(template.Channel),
			TemplateIds: template.TemplateIds,
			Tenant:      template.Tenant,
			Status:      template.Status,
			CreatedAt:   template.CreatedAt.Format(helper.TimeFormat),
			UpdatedAt:   template.UpdatedAt.Format(helper.TimeFormat),
		}
		response.Templates = append(response.Templates, responseItem)
	}

	logger.WithFields(logrus.Fields{
		"total_count":  total,
		"result_count": len(response.Templates),
	}).Info("Templates listed successfully")

	return response, nil
}
