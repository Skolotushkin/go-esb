package adapter

import (
	"context"
	"fmt"
	"time"

	"go-esb/internal/models"

	"github.com/streadway/amqp"
)

// AMQPAdapter реализует AMQP протокол (RabbitMQ)
type AMQPAdapter struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewAMQPAdapter создает новый AMQP адаптер
func NewAMQPAdapter() *AMQPAdapter {
	return &AMQPAdapter{}
}

// Connect подключается к RabbitMQ
func (a *AMQPAdapter) Connect(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to AMQP: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	a.conn = conn
	a.ch = ch
	return nil
}

// Send отправляет сообщение в очередь (endpoint содержит queue name, action содержит exchange)
func (a *AMQPAdapter) Send(ctx context.Context, endpoint string, action string, headers map[string]string, body []byte) ([]byte, int, error) {
	if a.ch == nil {
		return nil, 0, fmt.Errorf("AMQP channel not initialized. Call Connect() first")
	}

	queueName := endpoint
	if queueName == "" {
		queueName = "default"
	}

	exchangeName := action // action используется для exchange name в AMQP

	// Declare queue
	_, err := a.ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Prepare message
	msgHeaders := amqp.Table{}
	for k, v := range headers {
		msgHeaders[k] = v
	}

	err = a.ch.Publish(
		exchangeName, // может быть пустым для default exchange
		queueName,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Headers:      msgHeaders,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to publish message: %w", err)
	}

	return []byte(`{"status":"ok","message":"published to queue"}`), 200, nil
}

// Authenticate выполняет аутентификацию для AMQP (обычно через URL)
func (a *AMQPAdapter) Authenticate(auth *models.ConnectionAuthentication, endpoint string) (map[string]string, error) {
	// AMQP аутентификация обычно происходит через URL
	// (amqp://user:password@host:port/vhost)
	headers := make(map[string]string)
	return headers, nil
}

// Close закрывает соединение
func (a *AMQPAdapter) Close() error {
	if a.ch != nil {
		a.ch.Close()
	}
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

