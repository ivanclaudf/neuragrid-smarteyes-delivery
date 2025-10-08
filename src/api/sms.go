package api

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services/queue"
	"errors"

	"gorm.io/gorm"
)

// SMSAPI handles SMS business logic
type SMSAPI struct {
	DB          *gorm.DB
	ReaderDB    *gorm.DB
	SMSProducer *queue.SMSProducer // This will need to be implemented
}

// NewSMSAPI creates a new SMS API
func NewSMSAPI(db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) (*SMSAPI, error) {
	helper.Log.Debug("Initializing SMS API")

	if db == nil {
		helper.Log.Error("Failed to initialize SMS API: writer database connection is nil")
		return nil, errors.New("writer database connection is nil")
	}
	if readerDB == nil {
		helper.Log.Error("Failed to initialize SMS API: reader database connection is nil")
		return nil, errors.New("reader database connection is nil")
	}
	if pulsarClient == nil {
		helper.Log.Error("Failed to initialize SMS API: pulsar client is nil")
		return nil, errors.New("pulsar client is nil")
	}

	producer := queue.NewSMSProducer(pulsarClient, db) // This will need to be implemented
	helper.Log.Info("SMS API initialized successfully")

	return &SMSAPI{
		DB:          db,
		ReaderDB:    readerDB,
		SMSProducer: producer,
	}, nil
}

// ProcessMessageBatch processes a batch of SMS messages
func (a *SMSAPI) ProcessMessageBatch(request models.SMSRequest) ([]models.SMSMessageResponse, error) {
	batchLogger := helper.Log.WithFields(map[string]interface{}{
		"batchSize": len(request.Messages),
	})

	batchLogger.Info("Starting to process SMS message batch")
	responses := []models.SMSMessageResponse{}

	for idx, message := range request.Messages {
		messageLogger := batchLogger.WithFields(map[string]interface{}{
			"messageIndex": idx,
			"provider":     message.Provider,
			"refNo":        message.RefNo,
			"tenant":       message.Identifiers.Tenant,
			"from":         message.From,
		})

		messageLogger.Debug("Processing individual SMS message")

		//TODO need to handle de duplication based on RefNo
		messageLogger.Debug("Generating UUID for SMS message")
		// Generate a random UUID
		messageUUID, err := helper.GenerateUUID()
		if err != nil {
			messageLogger.WithError(err).Error("Failed to generate UUID for SMS message")
			return nil, errors.New("failed to generate message ID: " + err.Error())
		}

		// Log message information with UUID
		messageLogger = messageLogger.WithField("uuid", messageUUID)
		messageLogger.Info("Processing SMS message")

		// Create a message response
		messageResponse := models.SMSMessageResponse{
			RefNo: message.RefNo,
			UUID:  messageUUID,
		}

		// Add to responses
		responses = append(responses, messageResponse)

		// Send to queue
		messageLogger.Debug("Sending SMS message to queue")
		if err := a.SMSProducer.ProduceSMSMessage(&message, messageUUID); err != nil {
			messageLogger.WithError(err).Error("Failed to produce SMS message to queue")
			return nil, errors.New("failed to process message: " + err.Error())
		}
		messageLogger.Debug("Successfully sent SMS message to queue")
	}

	batchLogger.WithField("responseCount", len(responses)).Info("Successfully processed SMS message batch")
	return responses, nil
}
