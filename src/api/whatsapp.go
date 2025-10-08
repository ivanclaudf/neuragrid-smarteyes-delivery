package api

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services/queue"
	"errors"

	"gorm.io/gorm"
)

// WhatsAppAPI handles WhatsApp business logic
type WhatsAppAPI struct {
	DB              *gorm.DB
	ReaderDB        *gorm.DB
	MessageProducer *queue.WhatsAppProducer
}

// NewWhatsAppAPI creates a new WhatsApp API
func NewWhatsAppAPI(db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) (*WhatsAppAPI, error) {
	helper.Log.Debug("Initializing WhatsApp API")

	if db == nil {
		helper.Log.Error("Failed to initialize WhatsApp API: writer database connection is nil")
		return nil, errors.New("writer database connection is nil")
	}
	if readerDB == nil {
		helper.Log.Error("Failed to initialize WhatsApp API: reader database connection is nil")
		return nil, errors.New("reader database connection is nil")
	}
	if pulsarClient == nil {
		helper.Log.Error("Failed to initialize WhatsApp API: pulsar client is nil")
		return nil, errors.New("pulsar client is nil")
	}

	producer := queue.NewWhatsAppProducer(pulsarClient, db)
	helper.Log.Info("WhatsApp API initialized successfully")

	return &WhatsAppAPI{
		DB:              db,
		ReaderDB:        readerDB,
		MessageProducer: producer,
	}, nil
}

// ProcessMessageBatch processes a batch of WhatsApp messages
func (a *WhatsAppAPI) ProcessMessageBatch(request models.WhatsAppRequest) ([]models.WhatsAppMessageResponse, error) {
	batchLogger := helper.Log.WithFields(map[string]interface{}{
		"batchSize": len(request.Messages),
	})

	batchLogger.Info("Starting to process WhatsApp message batch")
	responses := []models.WhatsAppMessageResponse{}

	for idx, message := range request.Messages {
		messageLogger := batchLogger.WithFields(map[string]interface{}{
			"messageIndex": idx,
			"template":     message.Template,
			"provider":     message.Provider,
			"refNo":        message.RefNo,
			"tenant":       message.Identifiers.Tenant,
			"to":           message.To,
		})

		messageLogger.Debug("Processing individual WhatsApp message")

		//TODO need to handle de duplication based on RefNo
		messageLogger.Debug("Generating UUID for WhatsApp message")
		// Generate a random UUID
		messageUUID, err := helper.GenerateUUID()
		if err != nil {
			messageLogger.WithError(err).Error("Failed to generate UUID for WhatsApp message")
			return nil, errors.New("failed to generate message ID: " + err.Error())
		}

		// Log message information with UUID
		messageLogger = messageLogger.WithField("uuid", messageUUID)
		messageLogger.Info("Processing WhatsApp message")

		// Create a message response
		messageResponse := models.WhatsAppMessageResponse{
			RefNo: message.RefNo,
			UUID:  messageUUID,
		}

		// Add to responses
		responses = append(responses, messageResponse)

		// Send to queue
		messageLogger.Debug("Sending WhatsApp message to queue")
		if err := a.MessageProducer.ProduceWhatsAppMessage(&message, messageUUID); err != nil {
			messageLogger.WithError(err).Error("Failed to produce WhatsApp message to queue")
			return nil, errors.New("failed to process message: " + err.Error())
		}
		messageLogger.Debug("Successfully sent WhatsApp message to queue")
	}

	batchLogger.WithField("responseCount", len(responses)).Info("Successfully processed WhatsApp message batch")
	return responses, nil

	return responses, nil
}
