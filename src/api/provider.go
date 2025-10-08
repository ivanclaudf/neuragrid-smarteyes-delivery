package api

import (
	"delivery/helper"
	"delivery/models"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ProviderAPI handles provider business logic
type ProviderAPI struct {
	DB       *gorm.DB
	ReaderDB *gorm.DB
}

// NewProviderAPI creates a new provider API
func NewProviderAPI(db *gorm.DB, readerDB *gorm.DB) (*ProviderAPI, error) {
	logger := helper.Log.WithField("component", "ProviderAPI")

	if db == nil {
		logger.Error("Writer database connection is nil")
		return nil, fmt.Errorf("writer database connection is nil")
	}
	if readerDB == nil {
		logger.Error("Reader database connection is nil")
		return nil, fmt.Errorf("reader database connection is nil")
	}

	logger.Info("Provider API initialized successfully")
	return &ProviderAPI{
		DB:       db,
		ReaderDB: readerDB,
	}, nil
}

// CreateProviders creates new providers
func (a *ProviderAPI) CreateProviders(request models.ProviderRequest) (*models.ProviderResponse, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "ProviderAPI",
		"method":    "CreateProviders",
		"count":     len(request.Providers),
	})

	logger.Info("Creating new providers")

	// The validation for required fields is now handled by binding tags
	// We'll just do provider-specific validation here

	response := &models.ProviderResponse{
		Providers: make([]models.ProviderResponseItem, 0, len(request.Providers)),
	}

	for idx, providerItem := range request.Providers {
		providerLogger := logger.WithFields(logrus.Fields{
			"index":    idx,
			"code":     providerItem.Code,
			"provider": providerItem.Provider,
			"name":     providerItem.Name,
			"channel":  providerItem.Channel,
			"tenant":   providerItem.Tenant,
		})

		providerLogger.Debug("Processing provider item")

		// Encrypt secureConfig
		encryptedConfig, err := a.encryptSecureConfig(providerItem.SecureConfig)
		if err != nil {
			providerLogger.WithError(err).Error("Failed to encrypt secure config")
			return nil, fmt.Errorf("failed to encrypt secure config: %v", err)
		}

		// Create provider object using DB defaults
		provider := models.Provider{
			Code:         providerItem.Code,
			Provider:     providerItem.Provider,
			Name:         providerItem.Name,
			Config:       providerItem.Config,
			SecureConfig: encryptedConfig,
			Channel:      models.Channel(providerItem.Channel),
			Tenant:       providerItem.Tenant,
		}

		// Set status if provided (otherwise DB default will be used)
		if providerItem.Status != nil {
			provider.Status = *providerItem.Status
		}

		// Generate UUID for the provider
		if err := provider.GenerateUUID(); err != nil {
			providerLogger.WithError(err).Error("Failed to generate UUID")
			return nil, fmt.Errorf("failed to generate UUID: %v", err)
		}

		// Save provider to database
		if err := a.DB.Create(&provider).Error; err != nil {
			providerLogger.WithError(err).Error("Failed to create provider in database")
			return nil, fmt.Errorf("failed to create provider: %v", err)
		}

		providerLogger.WithField("uuid", provider.UUID).Info("Provider created successfully")

		// Add to response
		responseItem := models.ProviderResponseItem{
			UUID:      provider.UUID,
			Code:      provider.Code,
			Provider:  provider.Provider,
			Name:      provider.Name,
			Config:    provider.Config,
			Channel:   string(provider.Channel),
			Tenant:    provider.Tenant,
			Status:    provider.Status,
			CreatedAt: provider.CreatedAt.Format(helper.TimeFormat),
			UpdatedAt: provider.UpdatedAt.Format(helper.TimeFormat),
		}

		response.Providers = append(response.Providers, responseItem)
	}

	logger.WithField("created_count", len(response.Providers)).Info("Providers created successfully")
	return response, nil
}

// UpdateProvider updates an existing provider
func (a *ProviderAPI) UpdateProvider(uuid string, request models.ProviderRequest) (*models.ProviderResponse, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "ProviderAPI",
		"method":    "UpdateProvider",
		"uuid":      uuid,
	})

	logger.Info("Updating provider")

	if uuid == "" {
		logger.Error("Missing provider UUID")
		return nil, fmt.Errorf("missing provider UUID")
	}

	if len(request.Providers) != 1 {
		logger.WithField("count", len(request.Providers)).Error("Update requires exactly one provider")
		return nil, fmt.Errorf("update requires exactly one provider")
	}

	providerItem := request.Providers[0]
	logger = logger.WithFields(logrus.Fields{
		"code":     providerItem.Code,
		"provider": providerItem.Provider,
		"name":     providerItem.Name,
		"channel":  providerItem.Channel,
		"tenant":   providerItem.Tenant,
	})

	// Get existing provider
	var provider models.Provider
	if err := a.DB.Where("uuid = ?", uuid).First(&provider).Error; err != nil {
		logger.WithError(err).Error("Provider not found")
		return nil, fmt.Errorf("provider not found: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"existing_code":     provider.Code,
		"existing_provider": provider.Provider,
	}).Debug("Found existing provider")

	// Update only the fields that are provided
	updates := make(map[string]interface{})

	// Code should not be editable
	if providerItem.Code != "" && providerItem.Code != provider.Code {
		logger.WithFields(logrus.Fields{
			"existing_code": provider.Code,
			"new_code":      providerItem.Code,
		}).Error("Provider code cannot be changed")
		return nil, fmt.Errorf("provider code cannot be changed")
	}

	// Provider should not be editable
	if providerItem.Provider != "" && providerItem.Provider != provider.Provider {
		logger.WithFields(logrus.Fields{
			"existing_provider": provider.Provider,
			"new_provider":      providerItem.Provider,
		}).Error("Provider implementation cannot be changed")
		return nil, fmt.Errorf("provider implementation cannot be changed")
	}

	if providerItem.Name != "" {
		updates["name"] = providerItem.Name
	}

	if providerItem.Config != nil {
		updates["config"] = providerItem.Config
	}

	if providerItem.SecureConfig != nil {
		logger.Debug("Encrypting secure config")
		encryptedConfig, err := a.encryptSecureConfig(providerItem.SecureConfig)
		if err != nil {
			logger.WithError(err).Error("Failed to encrypt secure config")
			return nil, fmt.Errorf("failed to encrypt secure config: %v", err)
		}
		updates["secure_config"] = encryptedConfig
		logger.Debug("Secure config encrypted successfully")
	}

	if providerItem.Channel != "" {
		updates["channel"] = models.Channel(providerItem.Channel)
	}

	if providerItem.Tenant != "" {
		updates["tenant"] = providerItem.Tenant
	}

	if providerItem.Status != nil {
		updates["status"] = *providerItem.Status
	}

	// Update provider in database (only if there are fields to update)
	if len(updates) > 0 {
		// Get the keys that are being updated for logging
		updateFields := make([]string, 0, len(updates))
		for k := range updates {
			updateFields = append(updateFields, k)
		}

		logger.WithField("update_fields", updateFields).Debug("Updating provider fields")
		if err := a.DB.Model(&provider).Updates(updates).Error; err != nil {
			logger.WithError(err).Error("Failed to update provider")
			return nil, fmt.Errorf("failed to update provider: %v", err)
		}
		logger.Debug("Provider updated in database")
	} else {
		logger.Debug("No fields to update")
	}

	// Get updated provider
	if err := a.DB.Where("uuid = ?", uuid).First(&provider).Error; err != nil {
		logger.WithError(err).Error("Failed to retrieve updated provider")
		return nil, fmt.Errorf("failed to retrieve updated provider: %v", err)
	}
	logger.Debug("Retrieved updated provider details")

	// Create response
	response := &models.ProviderResponse{
		Providers: []models.ProviderResponseItem{
			{
				UUID:      provider.UUID,
				Code:      provider.Code,
				Provider:  provider.Provider,
				Name:      provider.Name,
				Config:    provider.Config,
				Channel:   string(provider.Channel),
				Tenant:    provider.Tenant,
				Status:    provider.Status,
				CreatedAt: provider.CreatedAt.Format(helper.TimeFormat),
				UpdatedAt: provider.UpdatedAt.Format(helper.TimeFormat),
			},
		},
	}

	logger.Info("Provider updated successfully")
	return response, nil
}

// GetProvider retrieves a single provider by UUID
func (a *ProviderAPI) GetProvider(uuid string) (*models.ProviderResponse, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "ProviderAPI",
		"method":    "GetProvider",
		"uuid":      uuid,
	})

	logger.Info("Retrieving provider by UUID")

	if uuid == "" {
		logger.Error("Missing provider UUID")
		return nil, fmt.Errorf("missing provider UUID")
	}

	// Get provider from database
	var provider models.Provider
	if err := a.ReaderDB.Where("uuid = ?", uuid).First(&provider).Error; err != nil {
		logger.WithError(err).Error("Provider not found")
		return nil, fmt.Errorf("provider not found: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"code":     provider.Code,
		"provider": provider.Provider,
		"name":     provider.Name,
		"channel":  provider.Channel,
		"tenant":   provider.Tenant,
	}).Debug("Provider found")

	// Create response
	response := &models.ProviderResponse{
		Providers: []models.ProviderResponseItem{
			{
				UUID:      provider.UUID,
				Code:      provider.Code,
				Provider:  provider.Provider,
				Name:      provider.Name,
				Config:    provider.Config,
				Channel:   string(provider.Channel),
				Tenant:    provider.Tenant,
				Status:    provider.Status,
				CreatedAt: provider.CreatedAt.Format(helper.TimeFormat),
				UpdatedAt: provider.UpdatedAt.Format(helper.TimeFormat),
			},
		},
	}

	logger.Info("Provider retrieved successfully")
	return response, nil
}

// ListProviders retrieves a list of providers with pagination
func (a *ProviderAPI) ListProviders(limit int, offset int, channel string, tenant string) (*models.ProviderResponse, int64, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "ProviderAPI",
		"method":    "ListProviders",
		"limit":     limit,
		"offset":    offset,
		"channel":   channel,
		"tenant":    tenant,
	})

	logger.Info("Listing providers")

	// Apply defaults
	if limit <= 0 {
		limit = 10 // Default limit
		logger.WithField("default_limit", limit).Debug("Using default limit")
	}

	if offset < 0 {
		offset = 0 // Default offset
		logger.WithField("default_offset", offset).Debug("Using default offset")
	}

	// Build query
	query := a.ReaderDB.Model(&models.Provider{})

	// Apply filters
	if channel != "" {
		query = query.Where("channel = ?", channel)
		logger.WithField("channel_filter", channel).Debug("Applied channel filter")
	}

	if tenant != "" {
		query = query.Where("tenant = ?", tenant)
		logger.WithField("tenant_filter", tenant).Debug("Applied tenant filter")
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count providers")
		return nil, 0, fmt.Errorf("failed to count providers: %v", err)
	}

	logger.WithField("total_count", total).Debug("Counted total number of providers")

	// Get providers with pagination
	var providers []models.Provider
	if err := query.Limit(limit).Offset(offset).Find(&providers).Error; err != nil {
		logger.WithError(err).Error("Failed to retrieve providers")
		return nil, 0, fmt.Errorf("failed to retrieve providers: %v", err)
	}

	logger.WithField("retrieved_count", len(providers)).Debug("Retrieved providers")

	// Create response
	responseItems := make([]models.ProviderResponseItem, 0, len(providers))
	for _, provider := range providers {
		responseItem := models.ProviderResponseItem{
			UUID:      provider.UUID,
			Code:      provider.Code,
			Provider:  provider.Provider,
			Name:      provider.Name,
			Config:    provider.Config,
			Channel:   string(provider.Channel),
			Tenant:    provider.Tenant,
			Status:    provider.Status,
			CreatedAt: provider.CreatedAt.Format(helper.TimeFormat),
			UpdatedAt: provider.UpdatedAt.Format(helper.TimeFormat),
		}
		responseItems = append(responseItems, responseItem)
	}

	response := &models.ProviderResponse{
		Providers: responseItems,
	}

	logger.WithFields(logrus.Fields{
		"total":    total,
		"returned": len(responseItems),
	}).Info("Providers listed successfully")
	return response, total, nil
}

// encryptSecureConfig encrypts sensitive provider configuration
func (a *ProviderAPI) encryptSecureConfig(secureConfig models.JSON) (models.JSON, error) {
	logger := helper.Log.WithFields(logrus.Fields{
		"component": "ProviderAPI",
		"method":    "encryptSecureConfig",
	})

	logger.Debug("Encrypting secure configuration")

	// Convert JSON to string
	secureConfigBytes, err := json.Marshal(secureConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal secure config")
		return nil, fmt.Errorf("failed to marshal secure config: %w", err)
	}

	// Get encryption key from environment
	encryptionKey := []byte(helper.GetEnv("ENCRYPTION_KEY", ""))
	if len(encryptionKey) != 32 {
		// For development, use a fixed key if not provided
		encryptionKey = []byte("12345678901234567890123456789012")
		logger.Debug("Using default development encryption key")
	} else {
		logger.Debug("Using environment encryption key")
	}

	// Encrypt and encode to base64
	encryptedBase64, err := helper.EncryptAndEncodeBase64(string(secureConfigBytes), encryptionKey)
	if err != nil {
		logger.WithError(err).Error("Failed to encrypt secure config")
		return nil, fmt.Errorf("failed to encrypt secure config: %w", err)
	}

	logger.WithField("encrypted_length", len(encryptedBase64)).Debug("Secure config encrypted successfully")

	// Create a new JSON object with the encrypted data
	return models.JSON{"encrypted": encryptedBase64}, nil
}
