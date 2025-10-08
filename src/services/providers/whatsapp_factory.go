package providers

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services"
	"delivery/services/providers/whatsapp"
	"errors"
)

// CreateWhatsAppProvider creates a WhatsApp provider based on the provider code
func CreateWhatsAppProvider(provider *models.Provider) (services.WhatsAppService, error) {
	logger := helper.Log.WithField("component", "WhatsAppProviderFactory")

	if provider == nil {
		logger.Error("Provider cannot be nil")
		return nil, errors.New("provider cannot be nil")
	}

	logger = logger.WithFields(map[string]interface{}{
		"providerUUID": provider.UUID,
		"providerCode": provider.Code,
		"providerImpl": provider.Provider,
	})

	logger.Debug("Creating WhatsApp provider instance")

	// Create provider based on the implementation class (provider column)
	switch provider.Provider {
	case "TWILIO":
		logger.Info("Creating Twilio WhatsApp provider")
		whatsappProvider, err := whatsapp.NewTwilioProviderFromDB(provider)
		if err != nil {
			logger.WithError(err).Error("Failed to create Twilio WhatsApp provider")
			return nil, err
		}
		logger.Debug("Twilio WhatsApp provider created successfully")
		return whatsappProvider, nil
	// Add additional provider implementations here as they become available
	default:
		unsupportedErr := "unsupported WhatsApp provider implementation: " + provider.Provider
		logger.Error(unsupportedErr)
		return nil, errors.New(unsupportedErr)
	}
}
