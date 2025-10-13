package api

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services/queue"
	"errors"

	"gorm.io/gorm"
)

// SMSRequest represents the request body for sending SMS messages
type SMSRequest struct {
	Messages []SMSMessage `json:"messages" validate:"required,min=1"`
}

// SMSMessage represents a single SMS message
type SMSMessage struct {
	To          []SMSRecipient    `json:"to" validate:"required,min=1"`
	From        string            `json:"from" validate:"required"`
	Body        string            `json:"body"` // For backward compatibility, not required when using template
	Template    string            `json:"template" validate:"required"`
	Provider    string            `json:"provider" validate:"required,uuid4"`
	RefNo       string            `json:"refno" validate:"required"`
	Categories  []string          `json:"categories" validate:"required,min=1"`
	Identifiers SMSIdentifiers    `json:"identifiers" validate:"required"`
	Params      map[string]string `json:"params"`
}

// ToModelSMSMessage converts API SMSMessage to models.SMSMessage
func (s *SMSMessage) ToModelSMSMessage() *models.SMSMessage {
	modelMessage := &models.SMSMessage{
		From:       s.From,
		Body:       s.Body,
		Template:   s.Template,
		Provider:   s.Provider,
		RefNo:      s.RefNo,
		Categories: s.Categories,
		Identifiers: models.SMSIdentifiers{
			Tenant:     s.Identifiers.Tenant,
			EventUUID:  s.Identifiers.EventUUID,
			ActionUUID: s.Identifiers.ActionUUID,
			ActionCode: s.Identifiers.ActionCode,
		},
		Params: s.Params,
	}

	// Convert recipients
	modelMessage.To = make([]models.SMSRecipient, len(s.To))
	for i, recipient := range s.To {
		modelMessage.To[i] = models.SMSRecipient{
			Telephone: recipient.Telephone,
		}
	}

	return modelMessage
}

// SMSRecipient represents a recipient for an SMS message
type SMSRecipient struct {
	Telephone string `json:"telephone" validate:"required,e164"`
}

// SMSIdentifiers represents identifiers for an SMS message
type SMSIdentifiers struct {
	Tenant     string `json:"tenant" validate:"required"`
	EventUUID  string `json:"eventUuid" validate:"omitempty,uuid4"`
	ActionUUID string `json:"actionUuid" validate:"omitempty,uuid4"`
	ActionCode string `json:"actionCode"`
}

// SMSResponse represents the response body for sending SMS messages
type SMSResponse struct {
	Messages []SMSMessageResponse `json:"messages"`
}

// SMSMessageResponse represents a response for a single SMS message
type SMSMessageResponse struct {
	RefNo string `json:"refno"`
	UUID  string `json:"uuid"`
}

// CreateSMSMessageResponse creates a new SMSMessageResponse
func CreateSMSMessageResponse(refNo string, uuid string) SMSMessageResponse {
	return SMSMessageResponse{
		RefNo: refNo,
		UUID:  uuid,
	}
}

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
func (a *SMSAPI) ProcessMessageBatch(request SMSRequest) ([]SMSMessageResponse, error) {
	batchLogger := helper.Log.WithFields(map[string]interface{}{
		"batchSize": len(request.Messages),
	})

	batchLogger.Info("Starting to process SMS message batch")
	responses := []SMSMessageResponse{}

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
		messageResponse := CreateSMSMessageResponse(message.RefNo, messageUUID)

		// Add to responses
		responses = append(responses, messageResponse)

		// Convert to model message
		modelMessage := message.ToModelSMSMessage()

		// Send to queue
		messageLogger.Debug("Sending SMS message to queue")
		if err := a.SMSProducer.ProduceSMSMessage(modelMessage, messageUUID); err != nil {
			messageLogger.WithError(err).Error("Failed to produce SMS message to queue")
			return nil, errors.New("failed to process message: " + err.Error())
		}
		messageLogger.Debug("Successfully sent SMS message to queue")
	}

	batchLogger.WithField("responseCount", len(responses)).Info("Successfully processed SMS message batch")
	return responses, nil
}

// DirectPushSMSMessage pushes an SMS message directly to Pulsar queue
func (a *SMSAPI) DirectPushSMSMessage(modelMessage *models.SMSMessage) (string, error) {
	logger := helper.Log.WithFields(map[string]interface{}{
		"provider": modelMessage.Provider,
		"refNo":    modelMessage.RefNo,
		"tenant":   modelMessage.Identifiers.Tenant,
		"from":     modelMessage.From,
	})

	logger.Info("Creating direct push SMS message")

	// Create a direct push message
	directPushMessage, err := queue.NewDirectPushSMSMessage(
		a.DB,
		a.SMSProducer.PulsarClient,
		modelMessage,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to create direct push SMS message")
		return "", err
	}

	// Log with UUID
	logger = logger.WithField("uuid", directPushMessage.UUID)

	// Push the message directly to Pulsar
	logger.Debug("Pushing SMS message directly to queue")
	if err := directPushMessage.Push(); err != nil {
		logger.WithError(err).Error("Failed to push SMS message directly to queue")
		return "", err
	}

	logger.Info("Successfully pushed SMS message directly to queue")
	return directPushMessage.UUID, nil
}
