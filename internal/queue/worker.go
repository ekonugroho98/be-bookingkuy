package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Worker handles message consumption from queues
type Worker struct {
	rabbitMQ *RabbitMQClient
	handlers map[string]MessageHandler
	running  bool
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// MessageHandler handles a message from queue
type MessageHandler func(ctx context.Context, payload map[string]interface{}) error

// NewWorker creates a new queue worker
func NewWorker(rabbitMQ *RabbitMQClient) *Worker {
	return &Worker{
		rabbitMQ: rabbitMQ,
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler registers a message handler for a queue
func (w *Worker) RegisterHandler(queue string, handler MessageHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.handlers[queue] = handler
	logger.Infof("Handler registered for queue: %s", queue)
}

// Start starts consuming messages from all registered queues
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.running = true
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.mu.Unlock()

	logger.Info("Queue worker started")

	// Start consumer for each registered queue
	for queue := range w.handlers {
		w.wg.Add(1)
		go w.consumeQueue(queue)
	}

	return nil
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	w.mu.Unlock()

	logger.Info("Stopping queue worker...")
	w.cancel()
	w.wg.Wait()
	logger.Info("Queue worker stopped")
}

// consumeQueue consumes messages from a specific queue
func (w *Worker) consumeQueue(queue string) {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			logger.Infof("Consumer for queue %s stopped", queue)
			return
		default:
			// Check connection
			if !w.rabbitMQ.IsConnected() {
				logger.Warnf("RabbitMQ not connected, retrying in 5s...")
				time.Sleep(5 * time.Second)
				continue
			}

			// Start consuming
			msgs, err := w.rabbitMQ.Consume(w.ctx, queue)
			if err != nil {
				logger.Errorf("Failed to consume from queue %s: %v", queue, err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Process messages
			for {
				select {
				case <-w.ctx.Done():
					return
				case msg, ok := <-msgs:
					if !ok {
						logger.Warnf("Consumer channel closed for queue %s", queue)
						return
					}

					// Process message
					w.processMessage(queue, msg)
				}
			}
		}
	}
}

// processMessage processes a single message
func (w *Worker) processMessage(queue string, msg amqp.Delivery) {
	// Get handler
	w.mu.RLock()
	handler, exists := w.handlers[queue]
	w.mu.RUnlock()

	if !exists {
		logger.Errorf("No handler registered for queue: %s", queue)
		w.rabbitMQ.Nack(msg, false) // Don't requeue
		return
	}

	// Parse message
	var message Message
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		logger.Errorf("Failed to unmarshal message: %v", err)
		w.rabbitMQ.Nack(msg, false) // Don't requeue malformed messages
		return
	}

	// Handle message with timeout
	ctx, cancel := context.WithTimeout(w.ctx, 30*time.Second)
	defer cancel()

	if err := handler(ctx, message.Payload); err != nil {
		logger.Errorf("Handler failed for queue %s: %v", queue, err)
		// Requeue on error for retry
		w.rabbitMQ.Nack(msg, true)
		return
	}

	// Acknowledge successful processing
	if err := w.rabbitMQ.Ack(msg); err != nil {
		logger.Errorf("Failed to ack message: %v", err)
		return
	}

	logger.Debugf("Message processed successfully from queue: %s", queue)
}

// IsRunning returns worker status
func (w *Worker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}
