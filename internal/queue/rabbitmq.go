package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// RabbitMQClient represents RabbitMQ client
type RabbitMQClient struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	url           string
	mu            sync.Mutex
	reconnectDelay time.Duration
	isConnected   bool
}

// Config represents RabbitMQ configuration
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	VHost           string
	ReconnectDelay  time.Duration
}

// NewClient creates a new RabbitMQ client
func NewClient(config Config) (*RabbitMQClient, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.VHost,
	)

	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = 5 * time.Second
	}

	client := &RabbitMQClient{
		url:            url,
		reconnectDelay: config.ReconnectDelay,
	}

	// Connect to RabbitMQ
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	logger.Info("✅ RabbitMQ client connected")
	return client, nil
}

// Connect establishes connection to RabbitMQ
func (c *RabbitMQClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	c.conn, err = amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to dial RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	c.isConnected = true
	return nil
}

// Close closes RabbitMQ connection
func (c *RabbitMQClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			logger.ErrorWithErr(err, "Failed to close RabbitMQ channel")
		}
	}

	if c.conn != nil && !c.conn.IsClosed() {
		if err := c.conn.Close(); err != nil {
			logger.ErrorWithErr(err, "Failed to close RabbitMQ connection")
		}
	}

	c.isConnected = false
	logger.Info("RabbitMQ connection closed")
	return nil
}

// DeclareQueue declares a queue
func (c *RabbitMQClient) DeclareQueue(name string, durable bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return errors.New("not connected to RabbitMQ")
	}

	_, err := c.channel.QueueDeclare(
		name,   // queue name
		durable, // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Infof("Queue declared: %s (durable: %v)", name, durable)
	return nil
}

// Publish publishes a message to a queue
func (c *RabbitMQClient) Publish(ctx context.Context, queue string, message interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return errors.New("not connected to RabbitMQ")
	}

	// Serialize message to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish message
	err = c.channel.PublishWithContext(
		ctx,
		"",    // exchange (default exchange)
		queue, // routing key (queue name)
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // persistent message
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Debugf("Message published to queue: %s", queue)
	return nil
}

// Consume starts consuming messages from a queue
func (c *RabbitMQClient) Consume(ctx context.Context, queue string) (<-chan amqp.Delivery, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return nil, errors.New("not connected to RabbitMQ")
	}

	msgs, err := c.channel.Consume(
		queue, // queue name
		"",    // consumer tag
		false, // auto-ack (we'll manually ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	logger.Infof("Started consuming from queue: %s", queue)
	return msgs, nil
}

// Ack acknowledges a message
func (c *RabbitMQClient) Ack(msg amqp.Delivery) error {
	if err := msg.Ack(false); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}
	return nil
}

// Nack negatively acknowledges a message
func (c *RabbitMQClient) Nack(msg amqp.Delivery, requeue bool) error {
	if err := msg.Nack(false, requeue); err != nil {
		return fmt.Errorf("failed to nack message: %w", err)
	}
	return nil
}

// IsConnected returns connection status
func (c *RabbitMQClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected && c.conn != nil && !c.conn.IsClosed()
}

// Queue names
const (
	QueueEmail        = "notifications.email"
	QueueSMS          = "notifications.sms"
	QueueBookingSync  = "bookings.sync"
	QueuePaymentSync  = "payments.sync"
)

// Message represents a queue message
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}
