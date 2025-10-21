package api

import (
	"delivery/helper"
	"delivery/models"
	"delivery/services/queue"
	"errors"

	"gorm.io/gorm"
)

// WhatsAppRequest represents the request body for sending WhatsApp messages
type WhatsAppRequest struct {
	Messages []WhatsAppMessage `json:"messages" validate:"required,min=1"`
}

// WhatsAppMessage represents a single WhatsApp message
type WhatsAppMessage struct {
	Template    string                 `json:"template" validate:"required"`
	To          []WhatsAppRecipient    `json:"to" validate:"required,min=1"`
	Provider    string                 `json:"provider" validate:"required,uuid4"`
	RefNo       string                 `json:"refno" validate:"required"`
	Categories  []string               `json:"categories" validate:"required,min=1"`
	TenantID    string                 `json:"tenantId" validate:"required"`
	Identifiers map[string]interface{} `json:"identifiers" validate:"required"`
	Params      map[string]string      `json:"params"`
	Attachments *WhatsAppAttachments   `json:"attachments"`
}

// ToModelWhatsAppMessage converts API WhatsAppMessage to models.WhatsAppMessage
func (w *WhatsAppMessage) ToModelWhatsAppMessage() *models.WhatsAppMessage {
	modelMessage := &models.WhatsAppMessage{
		Template:    w.Template,
		Provider:    w.Provider,
		RefNo:       w.RefNo,
		Categories:  w.Categories,
		TenantID:    w.TenantID,
		Identifiers: w.Identifiers,
		Params:      w.Params,
	}

	// Convert recipients
	modelMessage.To = make([]models.WhatsAppRecipient, len(w.To))
	for i, recipient := range w.To {
		modelMessage.To[i] = models.WhatsAppRecipient{
			Name:      recipient.Name,
			Telephone: recipient.Telephone,
		}
	}

	// Convert attachments if present
	if w.Attachments != nil {
		modelAttachments := &models.WhatsAppAttachments{
			Inline: make([]models.WhatsAppInlineAttachment, len(w.Attachments.Inline)),
		}
		for i, attachment := range w.Attachments.Inline {
			modelAttachments.Inline[i] = models.WhatsAppInlineAttachment{
				Filename:  attachment.Filename,
				Type:      attachment.Type,
				Content:   attachment.Content,
				ContentID: attachment.ContentID,
			}
		}
		modelMessage.Attachments = modelAttachments
	}

	return modelMessage
}

// WhatsAppRecipient represents a recipient for a WhatsApp message
type WhatsAppRecipient struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone" validate:"required,e164"`
}

// WhatsAppAttachments represents attachments for a WhatsApp message
type WhatsAppAttachments struct {
	Inline []WhatsAppInlineAttachment `json:"inline"`
}

// WhatsAppInlineAttachment represents an inline attachment for a WhatsApp message
type WhatsAppInlineAttachment struct {
	Filename  string `json:"filename" validate:"required"`
	Type      string `json:"type" validate:"required"`
	Content   string `json:"content" validate:"required"`
	ContentID string `json:"contentId" validate:"required"`
}

// WhatsAppResponse represents the response body for sending WhatsApp messages
type WhatsAppResponse struct {
	Messages []WhatsAppMessageResponse `json:"messages"`
}

// WhatsAppMessageResponse represents a single WhatsApp message response
type WhatsAppMessageResponse struct {
	RefNo string `json:"refno"`
	UUID  string `json:"uuid"`
}

// CreateWhatsAppMessageResponse creates a new WhatsAppMessageResponse
func CreateWhatsAppMessageResponse(refNo string, uuid string) WhatsAppMessageResponse {
	return WhatsAppMessageResponse{
		RefNo: refNo,
		UUID:  uuid,
	}
}

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
func (a *WhatsAppAPI) ProcessMessageBatch(request WhatsAppRequest) ([]WhatsAppMessageResponse, error) {
	batchLogger := helper.Log.WithFields(map[string]interface{}{
		"batchSize": len(request.Messages),
	})

	batchLogger.Info("Starting to process WhatsApp message batch")
	responses := []WhatsAppMessageResponse{}

	for idx, message := range request.Messages {
		messageLogger := batchLogger.WithFields(map[string]interface{}{
			"messageIndex": idx,
			"template":     message.Template,
			"provider":     message.Provider,
			"refNo":        message.RefNo,
			"tenantId":     message.TenantID,
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
		messageResponse := CreateWhatsAppMessageResponse(message.RefNo, messageUUID)

		// Add to responses
		responses = append(responses, messageResponse)

		// Convert to model message
		modelMessage := message.ToModelWhatsAppMessage()

		// Send to queue
		messageLogger.Debug("Sending WhatsApp message to queue")
		if err := a.MessageProducer.ProduceWhatsAppMessage(modelMessage, messageUUID); err != nil {
			messageLogger.WithError(err).Error("Failed to produce WhatsApp message to queue")
			return nil, errors.New("failed to process message: " + err.Error())
		}
		messageLogger.Debug("Successfully sent WhatsApp message to queue")
	}

	batchLogger.WithField("responseCount", len(responses)).Info("Successfully processed WhatsApp message batch")
	return responses, nil
}

// DirectPushWhatsAppMessage pushes a WhatsApp message directly to Pulsar queue
func (a *WhatsAppAPI) DirectPushWhatsAppMessage(modelMessage *models.WhatsAppMessage) (string, error) {
	logger := helper.Log.WithFields(map[string]interface{}{
		"template": modelMessage.Template,
		"provider": modelMessage.Provider,
		"refNo":    modelMessage.RefNo,
		"tenantId": modelMessage.TenantID,
	})

	logger.Info("Creating direct push WhatsApp message")

	// Create a direct push message
	directPushMessage, err := queue.NewDirectPushWhatsAppMessage(
		a.DB,
		a.MessageProducer.PulsarClient,
		modelMessage,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to create direct push WhatsApp message")
		return "", err
	}

	// Log with UUID
	logger = logger.WithField("uuid", directPushMessage.UUID)

	// Push the message directly to Pulsar
	logger.Debug("Pushing WhatsApp message directly to queue")
	if err := directPushMessage.Push(); err != nil {
		logger.WithError(err).Error("Failed to push WhatsApp message directly to queue")
		return "", err
	}

	logger.Info("Successfully pushed WhatsApp message directly to queue")
	return directPushMessage.UUID, nil
}
