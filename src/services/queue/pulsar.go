package queue

import (
	"context"
	"encoding/json"
	"errors"
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

	return nil
}

// Close closes the Pulsar client connection
func (cm *ConsumerManager) Close() {
	if cm.pulsarClient != nil {
		cm.pulsarClient.Close()
	}
}
