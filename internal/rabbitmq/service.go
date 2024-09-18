package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type RabbitMQService struct {
	Connection *amqp.Connection
}

type UploadResponse struct {
	UploadID string   `json:"uploadId"`
	FileURLs []string `json:"fileUrls"`
}

func NewRabbitMQService() (*RabbitMQService, error) {
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	if rabbitHost == "" {
		return nil, fmt.Errorf("Missing RabbitMQ host")
	}
	conn, err := amqp.Dial(rabbitHost)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}

	return &RabbitMQService{
		Connection: conn,
	}, nil
}

func (s *RabbitMQService) PublishToQueue(queueName string, uploadID string, fileURLs []string) error {
	channel, err := s.Connection.Channel()
	if err != nil {
		return fmt.Errorf("Failed to open RabbitMQ channel: %v", err)
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(
		queueName, // name of the queue
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare queue: %v", err)
	}

	message := UploadResponse{
		UploadID: uploadID,
		FileURLs: fileURLs,
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("Failed to marshal message: %v", err)
	}

	err = channel.Publish(
		"",        // exchange
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})

	if err != nil {
		return fmt.Errorf("Failed to publish message: %v", err)
	}

	log.Printf("Message published to queue %s: %s", queueName, messageBody)
	return nil
}

func (s *RabbitMQService) Close() error {
	return s.Connection.Close()
}
