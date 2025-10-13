package queue

import (
	"context"
	"delivery/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"delivery/helper"

	"github.com/apache/pulsar-client-go/pulsar"
	"gorm.io/gorm"
)

const (
	WhatsAppTopic = "delivery-whatsapp"
	SMSTopic      = "delivery-sms"
	EmailTopic    = "delivery-email"
)

// PulsarClient wraps the Pulsar client with common operations
type PulsarClient struct {
	client pulsar.Client
}

// ConsumerManager handles initializing and managing message consumers
type ConsumerManager struct {
	pulsarClient *PulsarClient
	db           *gorm.DB
	readerDB     *gorm.DB
}

// NewPulsarClient creates a new Pulsar client
func NewPulsarClient() (*PulsarClient, error) {
	pulsarURL := os.Getenv("PULSAR_URL")
	if pulsarURL == "" {
		pulsarURL = "pulsar://localhost:6650"
	}

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               pulsarURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	return &PulsarClient{
		client: client,
	}, nil
}

// Close closes the Pulsar client
func (p *PulsarClient) Close() {
	p.client.Close()
}

// ProduceMessage produces a message to a topic
func (p *PulsarClient) ProduceMessage(topic string, message interface{}) error {
	producer, err := p.client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return err
	}
	defer producer.Close()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: data,
	})

	return err
}

// ConsumeMessages consumes messages from a topic
func (p *PulsarClient) ConsumeMessages(topic, subscription string, handler func(message []byte) error) error {
	consumer, err := p.client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared, // Allow multiple consumers to process messages
	})
	if err != nil {
		return err
	}
	defer consumer.Close()

	ctx := context.Background()
	for {
		msg, err := consumer.Receive(ctx)
		if err != nil {
			helper.Log.Errorf("Error receiving message: %v", err)
			continue
		}

		err = handler(msg.Payload())
		if err != nil {
			helper.Log.Errorf("Error handling message: %v", err)
			consumer.Nack(msg)
			continue
		}

		consumer.Ack(msg)
	}
}

// CreateConsumerGroup creates a consumer group for a topic
func (p *PulsarClient) CreateConsumerGroup(topic, subscription string, numConsumers int, handler func(message []byte) error) error {
	if numConsumers <= 0 {
		return errors.New("number of consumers must be greater than 0")
	}

	for i := 0; i < numConsumers; i++ {
		go func() {
			err := p.ConsumeMessages(topic, subscription, handler)
			if err != nil {
				helper.Log.Errorf("Error consuming messages: %v", err)
			}
		}()
	}

	return nil
}

// GetPulsarClient returns the Pulsar client for external use
func (cm *ConsumerManager) GetPulsarClient() *PulsarClient {
	return cm.pulsarClient
}

// NewConsumerManager creates a new consumer manager
func NewConsumerManager(db *gorm.DB, readerDB *gorm.DB) (*ConsumerManager, error) {
	// Initialize Pulsar client
	pulsarClient, err := NewPulsarClient()
	if err != nil {
		return nil, err
	}

	return &ConsumerManager{
		pulsarClient: pulsarClient,
		db:           db,
		readerDB:     readerDB,
	}, nil
}

// StartConsumers starts all message consumers
func (cm *ConsumerManager) StartConsumers() error {
	// Start WhatsApp consumer
	whatsAppConsumer := NewWhatsAppConsumer(cm.pulsarClient, cm.db, cm.readerDB)
	err := whatsAppConsumer.Start(1) // Start with a single consumer worker
	if err != nil {
		helper.Log.Errorf("Failed to start WhatsApp consumer: %v", err)
		return err
	}

	// Start SMS consumer
	smsConsumer, err := NewSMSConsumer(cm.pulsarClient, cm.db, cm.readerDB)
	if err != nil {
		helper.Log.Errorf("Failed to create SMS consumer: %v", err)
		return err
	}

	err = smsConsumer.Start()
	if err != nil {
		helper.Log.Errorf("Failed to start SMS consumer: %v", err)
		return err
	}

	// Start Email consumer
	emailConsumer, err := NewEmailConsumer(cm.pulsarClient, cm.db, cm.readerDB)
	if err != nil {
		helper.Log.Errorf("Failed to create Email consumer: %v", err)
		return err
	}

	err = emailConsumer.Start()
	if err != nil {
		helper.Log.Errorf("Failed to start Email consumer: %v", err)
		return err
	}

	return nil
}

// Close closes the Pulsar client connection
func (cm *ConsumerManager) Close() {
	if cm.pulsarClient != nil {
		cm.pulsarClient.Close()
	}
}

// DirectPushMessage is a generic interface for all message types that can be pushed directly to Pulsar
type DirectPushMessage interface {
	GetUUID() string
	GetChannel() models.Channel
	GetRefNo() string
	GetIdentifiers() map[string]interface{}
	GetCategories() []string
	Validate() error
}

// DirectPushEmailMessage represents an email message that can be pushed directly to Pulsar
type DirectPushEmailMessage struct {
	UUID       string              `json:"uuid"`
	Message    models.EmailMessage `json:"message"`
	DB         *gorm.DB            `json:"-"`
	PulsarConn *PulsarClient       `json:"-"`
}

// DirectPushSMSMessage represents an SMS message that can be pushed directly to Pulsar
type DirectPushSMSMessage struct {
	UUID       string            `json:"uuid"`
	Message    models.SMSMessage `json:"message"`
	DB         *gorm.DB          `json:"-"`
	PulsarConn *PulsarClient     `json:"-"`
}

// DirectPushWhatsAppMessage represents a WhatsApp message that can be pushed directly to Pulsar
type DirectPushWhatsAppMessage struct {
	UUID       string                 `json:"uuid"`
	Message    models.WhatsAppMessage `json:"message"`
	DB         *gorm.DB               `json:"-"`
	PulsarConn *PulsarClient          `json:"-"`
}

// GetUUID returns the UUID of the email message
func (m *DirectPushEmailMessage) GetUUID() string {
	return m.UUID
}

// GetChannel returns the channel type of the email message
func (m *DirectPushEmailMessage) GetChannel() models.Channel {
	return models.ChannelEmail
}

// GetRefNo returns the reference number of the email message
func (m *DirectPushEmailMessage) GetRefNo() string {
	return m.Message.RefNo
}

// GetIdentifiers returns the identifiers of the email message
func (m *DirectPushEmailMessage) GetIdentifiers() map[string]interface{} {
	return map[string]interface{}{
		"tenant":     m.Message.Identifiers.Tenant,
		"eventUuid":  m.Message.Identifiers.EventUUID,
		"actionUuid": m.Message.Identifiers.ActionUUID,
		"actionCode": m.Message.Identifiers.ActionCode,
	}
}

// GetCategories returns the categories of the email message
func (m *DirectPushEmailMessage) GetCategories() []string {
	return m.Message.Categories
}

// Validate validates the email message
func (m *DirectPushEmailMessage) Validate() error {
	if m.Message.Template == "" {
		return errors.New("template is required")
	}
	if len(m.Message.To) == 0 {
		return errors.New("at least one recipient is required")
	}
	if m.Message.Provider == "" {
		return errors.New("provider is required")
	}
	if m.Message.RefNo == "" {
		return errors.New("refNo is required")
	}
	if len(m.Message.Categories) == 0 {
		return errors.New("at least one category is required")
	}
	if m.Message.Identifiers.Tenant == "" {
		return errors.New("tenant identifier is required")
	}
	return nil
}

// Push pushes the email message to Pulsar and records it in the database
func (m *DirectPushEmailMessage) Push() error {
	if err := m.Validate(); err != nil {
		return err
	}

	// Create identifiers JSON for the database
	identifiersJSON := models.JSON{
		"tenant":     m.Message.Identifiers.Tenant,
		"eventUuid":  m.Message.Identifiers.EventUUID,
		"actionUuid": m.Message.Identifiers.ActionUUID,
		"actionCode": m.Message.Identifiers.ActionCode,
	}

	// Convert categories array to JSON
	categoriesJSON := models.JSON{}
	for i, category := range m.Message.Categories {
		categoriesJSON[fmt.Sprintf("%d", i)] = category
	}

	// Create a message record in the database
	dbMessage := models.Message{
		UUID:        m.UUID,
		Channel:     models.ChannelEmail,
		Status:      models.StatusAccepted,
		Identifiers: identifiersJSON,
		RefNo:       m.Message.RefNo,
		Categories:  categoriesJSON,
	}

	// Save to database
	if err := m.DB.Create(&dbMessage).Error; err != nil {
		return err
	}

	// Create queue message
	queueMessage := EmailMessage{
		UUID:    m.UUID,
		Message: m.Message,
	}

	// Produce the message to the queue
	return m.PulsarConn.ProduceMessage(EmailTopic, queueMessage)
}

// GetUUID returns the UUID of the SMS message
func (m *DirectPushSMSMessage) GetUUID() string {
	return m.UUID
}

// GetChannel returns the channel type of the SMS message
func (m *DirectPushSMSMessage) GetChannel() models.Channel {
	return models.ChannelSMS
}

// GetRefNo returns the reference number of the SMS message
func (m *DirectPushSMSMessage) GetRefNo() string {
	return m.Message.RefNo
}

// GetIdentifiers returns the identifiers of the SMS message
func (m *DirectPushSMSMessage) GetIdentifiers() map[string]interface{} {
	return map[string]interface{}{
		"tenant":     m.Message.Identifiers.Tenant,
		"eventUuid":  m.Message.Identifiers.EventUUID,
		"actionUuid": m.Message.Identifiers.ActionUUID,
		"actionCode": m.Message.Identifiers.ActionCode,
	}
}

// GetCategories returns the categories of the SMS message
func (m *DirectPushSMSMessage) GetCategories() []string {
	return m.Message.Categories
}

// Validate validates the SMS message
func (m *DirectPushSMSMessage) Validate() error {
	if m.Message.Template == "" {
		return errors.New("template is required")
	}
	if len(m.Message.To) == 0 {
		return errors.New("at least one recipient is required")
	}
	if m.Message.From == "" {
		return errors.New("from is required")
	}
	if m.Message.Provider == "" {
		return errors.New("provider is required")
	}
	if m.Message.RefNo == "" {
		return errors.New("refNo is required")
	}
	if len(m.Message.Categories) == 0 {
		return errors.New("at least one category is required")
	}
	if m.Message.Identifiers.Tenant == "" {
		return errors.New("tenant identifier is required")
	}
	return nil
}

// Push pushes the SMS message to Pulsar and records it in the database
func (m *DirectPushSMSMessage) Push() error {
	if err := m.Validate(); err != nil {
		return err
	}

	// Create identifiers JSON for the database
	identifiersJSON := models.JSON{
		"tenant":     m.Message.Identifiers.Tenant,
		"eventUuid":  m.Message.Identifiers.EventUUID,
		"actionUuid": m.Message.Identifiers.ActionUUID,
		"actionCode": m.Message.Identifiers.ActionCode,
	}

	// Convert categories array to JSON
	categoriesJSON := models.JSON{}
	for i, category := range m.Message.Categories {
		categoriesJSON[fmt.Sprintf("%d", i)] = category
	}

	// Create a message record in the database
	dbMessage := models.Message{
		UUID:        m.UUID,
		Channel:     models.ChannelSMS,
		Status:      models.StatusAccepted,
		Identifiers: identifiersJSON,
		RefNo:       m.Message.RefNo,
		Categories:  categoriesJSON,
	}

	// Save to database
	if err := m.DB.Create(&dbMessage).Error; err != nil {
		return err
	}

	// Create queue message
	queueMessage := SMSMessage{
		UUID:    m.UUID,
		Message: m.Message,
	}

	// Produce the message to the queue
	return m.PulsarConn.ProduceMessage(SMSTopic, queueMessage)
}

// GetUUID returns the UUID of the WhatsApp message
func (m *DirectPushWhatsAppMessage) GetUUID() string {
	return m.UUID
}

// GetChannel returns the channel type of the WhatsApp message
func (m *DirectPushWhatsAppMessage) GetChannel() models.Channel {
	return models.ChannelWhatsApp
}

// GetRefNo returns the reference number of the WhatsApp message
func (m *DirectPushWhatsAppMessage) GetRefNo() string {
	return m.Message.RefNo
}

// GetIdentifiers returns the identifiers of the WhatsApp message
func (m *DirectPushWhatsAppMessage) GetIdentifiers() map[string]interface{} {
	return map[string]interface{}{
		"tenant":     m.Message.Identifiers.Tenant,
		"eventUuid":  m.Message.Identifiers.EventUUID,
		"actionUuid": m.Message.Identifiers.ActionUUID,
		"actionCode": m.Message.Identifiers.ActionCode,
	}
}

// GetCategories returns the categories of the WhatsApp message
func (m *DirectPushWhatsAppMessage) GetCategories() []string {
	return m.Message.Categories
}

// Validate validates the WhatsApp message
func (m *DirectPushWhatsAppMessage) Validate() error {
	if m.Message.Template == "" {
		return errors.New("template is required")
	}
	if len(m.Message.To) == 0 {
		return errors.New("at least one recipient is required")
	}
	if m.Message.Provider == "" {
		return errors.New("provider is required")
	}
	if m.Message.RefNo == "" {
		return errors.New("refNo is required")
	}
	if len(m.Message.Categories) == 0 {
		return errors.New("at least one category is required")
	}
	if m.Message.Identifiers.Tenant == "" {
		return errors.New("tenant identifier is required")
	}
	return nil
}

// Push pushes the WhatsApp message to Pulsar and records it in the database
func (m *DirectPushWhatsAppMessage) Push() error {
	if err := m.Validate(); err != nil {
		return err
	}

	// Create identifiers JSON for the database
	identifiersJSON := models.JSON{
		"tenant":     m.Message.Identifiers.Tenant,
		"eventUuid":  m.Message.Identifiers.EventUUID,
		"actionUuid": m.Message.Identifiers.ActionUUID,
		"actionCode": m.Message.Identifiers.ActionCode,
	}

	// Convert categories array to JSON
	categoriesJSON := models.JSON{}
	for i, category := range m.Message.Categories {
		categoriesJSON[fmt.Sprintf("%d", i)] = category
	}

	// Create a message record in the database
	dbMessage := models.Message{
		UUID:        m.UUID,
		Channel:     models.ChannelWhatsApp,
		Status:      models.StatusAccepted,
		Identifiers: identifiersJSON,
		RefNo:       m.Message.RefNo,
		Categories:  categoriesJSON,
	}

	// Save to database
	if err := m.DB.Create(&dbMessage).Error; err != nil {
		return err
	}

	// Create queue message
	queueMessage := WhatsAppMessage{
		UUID:    m.UUID,
		Message: m.Message,
	}

	// Produce the message to the queue
	return m.PulsarConn.ProduceMessage(WhatsAppTopic, queueMessage)
}

// NewDirectPushEmailMessage creates a new direct push email message
func NewDirectPushEmailMessage(db *gorm.DB, pulsarClient *PulsarClient, message *models.EmailMessage) (*DirectPushEmailMessage, error) {
	// Generate a UUID
	uuid, err := helper.GenerateUUID()
	if err != nil {
		return nil, err
	}

	return &DirectPushEmailMessage{
		UUID:       uuid,
		Message:    *message,
		DB:         db,
		PulsarConn: pulsarClient,
	}, nil
}

// NewDirectPushSMSMessage creates a new direct push SMS message
func NewDirectPushSMSMessage(db *gorm.DB, pulsarClient *PulsarClient, message *models.SMSMessage) (*DirectPushSMSMessage, error) {
	// Generate a UUID
	uuid, err := helper.GenerateUUID()
	if err != nil {
		return nil, err
	}

	return &DirectPushSMSMessage{
		UUID:       uuid,
		Message:    *message,
		DB:         db,
		PulsarConn: pulsarClient,
	}, nil
}

// NewDirectPushWhatsAppMessage creates a new direct push WhatsApp message
func NewDirectPushWhatsAppMessage(db *gorm.DB, pulsarClient *PulsarClient, message *models.WhatsAppMessage) (*DirectPushWhatsAppMessage, error) {
	// Generate a UUID
	uuid, err := helper.GenerateUUID()
	if err != nil {
		return nil, err
	}

	return &DirectPushWhatsAppMessage{
		UUID:       uuid,
		Message:    *message,
		DB:         db,
		PulsarConn: pulsarClient,
	}, nil
}
