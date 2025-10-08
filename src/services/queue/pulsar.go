package queue

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"delivery/helper"

	"github.com/apache/pulsar-client-go/pulsar"
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
