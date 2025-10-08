package providers

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services"
	"delivery/services/providers/sms"
	"errors"
	"strings"
)

// CreateSMSProvider creates an SMS provider based on the provider code
func CreateSMSProvider(provider *models.Provider) (services.SMSService, error) {
	logger := helper.Log.WithField("component", "SMSProviderFactory")

	if provider == nil {
		logger.Error("Provider cannot be nil")
		return nil, errors.New("provider cannot be nil")
	}

	logger = logger.WithFields(map[string]interface{}{
		"providerUUID": provider.UUID,
		"providerCode": provider.Code,
		"providerImpl": provider.Provider,
	})

	logger.Debug("Creating SMS provider instance")

	// Create provider based on the implementation class (provider column) with case insensitive comparison
	providerType := strings.ToUpper(provider.Provider)

	switch providerType {
	case "TWILIO":
		logger.Info("Creating Twilio SMS provider")
		twilioProvider, err := sms.NewTwilioProviderFromDB(provider)
		if err != nil {
			logger.WithError(err).Error("Failed to create Twilio SMS provider")
			return nil, err
		}
		logger.Debug("Twilio SMS provider created successfully")
		return twilioProvider, nil
	// Add additional provider implementations here as they become available
	default:
		unsupportedErr := "unsupported SMS provider implementation: " + provider.Provider
		logger.Error(unsupportedErr)
		return nil, errors.New(unsupportedErr)
	}
}
