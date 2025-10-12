package api

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services/queue"
	"errors"

	"gorm.io/gorm"
)

// EmailRequest represents the request body for sending Email messages
type EmailRequest struct {
	Messages []EmailMessage `json:"messages" validate:"required,min=1"`
}

// EmailMessage represents a single email message request
type EmailMessage struct {
	Template    string               `json:"template" validate:"required"`
	To          []EmailRecipient     `json:"to" validate:"required,min=1"`
	Provider    string               `json:"provider" validate:"required,uuid4"`
	RefNo       string               `json:"refno" validate:"required"`
	Categories  []string             `json:"categories" validate:"required,min=1"`
	Identifiers EmailIdentifiers     `json:"identifiers" validate:"required"`
	Params      map[string]string    `json:"params"`
	Subject     string               `json:"subject,omitempty"`
	Attachments []AttachmentMetadata `json:"attachments,omitempty"`
}

// ToModelEmailMessage converts API EmailMessage to models.EmailMessage
func (e *EmailMessage) ToModelEmailMessage() *models.EmailMessage {
	modelMessage := &models.EmailMessage{
		Template:   e.Template,
		Provider:   e.Provider,
		RefNo:      e.RefNo,
		Categories: e.Categories,
		Identifiers: models.EmailIdentifiers{
			Tenant:     e.Identifiers.Tenant,
			EventUUID:  e.Identifiers.EventUUID,
			ActionUUID: e.Identifiers.ActionUUID,
			ActionCode: e.Identifiers.ActionCode,
		},
		Params:  e.Params,
		Subject: e.Subject,
	}

	// Convert recipients
	modelMessage.To = make([]models.EmailRecipient, len(e.To))
	for i, recipient := range e.To {
		modelMessage.To[i] = models.EmailRecipient{
			Name:  recipient.Name,
			Email: recipient.Email,
		}
	}

	// Convert attachments if present
	if len(e.Attachments) > 0 {
		modelMessage.Attachments = make([]models.AttachmentMetadata, len(e.Attachments))
		for i, attachment := range e.Attachments {
			modelMessage.Attachments[i] = models.AttachmentMetadata{
				Filename:    attachment.Filename,
				ContentType: attachment.ContentType,
				Content:     attachment.Content,
			}
		}
	}

	return modelMessage
}

// EmailRecipient represents an email recipient with name and email
type EmailRecipient struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email" validate:"required,email"`
}

// EmailIdentifiers represents identifiers for an email message
type EmailIdentifiers struct {
	Tenant     string `json:"tenant" validate:"required"`
	EventUUID  string `json:"eventUuid" validate:"omitempty,uuid4"`
	ActionUUID string `json:"actionUuid" validate:"omitempty,uuid4"`
	ActionCode string `json:"actionCode"`
}

// AttachmentMetadata represents metadata for an email attachment
type AttachmentMetadata struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"contentType" validate:"required"`
	Content     string `json:"content" validate:"required"` // base64 encoded
}

// EmailResponse represents the response body for sending Email messages
type EmailResponse struct {
	Messages []EmailMessageResponse `json:"messages"`
}

// EmailMessageResponse represents a response for a single Email message
type EmailMessageResponse struct {
	RefNo string `json:"refno"`
	UUID  string `json:"uuid"`
}

// CreateEmailMessageResponse creates a new EmailMessageResponse
func CreateEmailMessageResponse(refNo string, uuid string) EmailMessageResponse {
	return EmailMessageResponse{
		RefNo: refNo,
		UUID:  uuid,
	}
}

// EmailAPI handles Email business logic
type EmailAPI struct {
	DB              *gorm.DB
	ReaderDB        *gorm.DB
	MessageProducer *queue.EmailProducer
}

// NewEmailAPI creates a new Email API
func NewEmailAPI(db *gorm.DB, readerDB *gorm.DB, pulsarClient *queue.PulsarClient) (*EmailAPI, error) {
	helper.Log.Debug("Initializing Email API")

	if db == nil {
		helper.Log.Error("Failed to initialize Email API: writer database connection is nil")
		return nil, errors.New("writer database connection is nil")
	}
	if readerDB == nil {
		helper.Log.Error("Failed to initialize Email API: reader database connection is nil")
		return nil, errors.New("reader database connection is nil")
	}
	if pulsarClient == nil {
		helper.Log.Error("Failed to initialize Email API: pulsar client is nil")
		return nil, errors.New("pulsar client is nil")
	}

	producer := queue.NewEmailProducer(pulsarClient, db)
	helper.Log.Info("Email API initialized successfully")

	return &EmailAPI{
		DB:              db,
		ReaderDB:        readerDB,
		MessageProducer: producer,
	}, nil
}

// ProcessMessageBatch processes a batch of Email messages
func (a *EmailAPI) ProcessMessageBatch(request EmailRequest) ([]EmailMessageResponse, error) {
	batchLogger := helper.Log.WithFields(map[string]interface{}{
		"batchSize": len(request.Messages),
	})

	batchLogger.Info("Starting to process Email message batch")
	responses := []EmailMessageResponse{}

	for idx, message := range request.Messages {
		messageLogger := batchLogger.WithFields(map[string]interface{}{
			"messageIndex": idx,
			"template":     message.Template,
			"provider":     message.Provider,
			"refNo":        message.RefNo,
			"tenant":       message.Identifiers.Tenant,
			"to":           message.To,
		})

		messageLogger.Debug("Processing individual Email message")

		//TODO need to handle de duplication based on RefNo
		messageLogger.Debug("Generating UUID for Email message")
		// Generate a random UUID
		messageUUID, err := helper.GenerateUUID()
		if err != nil {
			messageLogger.WithError(err).Error("Failed to generate UUID for Email message")
			return nil, errors.New("failed to generate message ID: " + err.Error())
		}

		// Log message information with UUID
		messageLogger = messageLogger.WithField("uuid", messageUUID)
		messageLogger.Info("Processing Email message")

		// Create a message response
		messageResponse := CreateEmailMessageResponse(message.RefNo, messageUUID)

		// Add to responses
		responses = append(responses, messageResponse)

		// Convert to model message
		modelMessage := message.ToModelEmailMessage()

		// Send to queue
		messageLogger.Debug("Sending Email message to queue")
		if err := a.MessageProducer.ProduceEmailMessage(modelMessage, messageUUID); err != nil {
			messageLogger.WithError(err).Error("Failed to produce Email message to queue")
			return nil, errors.New("failed to process message: " + err.Error())
		}
		messageLogger.Debug("Successfully sent Email message to queue")
	}

	batchLogger.WithField("responseCount", len(responses)).Info("Successfully processed Email message batch")
	return responses, nil
}
