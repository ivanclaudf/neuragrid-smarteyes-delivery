package providers

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services"
	"delivery/services/providers/email"
	"errors"
	"strings"
)

// CreateEmailProvider creates an Email provider based on the provider code
func CreateEmailProvider(provider *models.Provider) (services.EmailService, error) {
	logger := helper.Log.WithField("component", "EmailProviderFactory")

	if provider == nil {
		logger.Error("Provider cannot be nil")
		return nil, errors.New("provider cannot be nil")
	}

	logger = logger.WithFields(map[string]interface{}{
		"providerUUID": provider.UUID,
		"providerCode": provider.Code,
		"providerImpl": provider.Provider,
	})

	logger.Debug("Creating Email provider instance")

	// Create provider based on the implementation class (provider column) with case insensitive comparison
	providerType := strings.ToUpper(provider.Provider)

	switch providerType {
	case "SENDGRID", "TWILIO": // Supporting both "SENDGRID" and "TWILIO" as SendGrid is now part of Twilio
		logger.Info("Creating SendGrid Email provider")
		sendGridProvider, err := email.NewSendGridProviderFromDB(provider)
		if err != nil {
			logger.WithError(err).Error("Failed to create SendGrid Email provider")
			return nil, err
		}
		logger.Debug("SendGrid Email provider created successfully")
		return sendGridProvider, nil
	// Add additional provider implementations here as they become available
	default:
		unsupportedErr := "unsupported Email provider implementation: " + provider.Provider
		logger.Error(unsupportedErr)
		return nil, errors.New(unsupportedErr)
	}
}
