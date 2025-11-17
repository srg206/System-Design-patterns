package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// CreateTopic создает топик с указанными параметрами
func CreateTopic(ctx context.Context, brokers []string, topic string, partitions int, replicationFactor int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("brokers list cannot be empty")
	}

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

// EnsureTopic проверяет существование топика и создает его при необходимости
// Использует параметры по умолчанию: 1 партиция, replication factor 1
func EnsureTopic(ctx context.Context, brokers []string, topic string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("brokers list cannot be empty")
	}

	// Пробуем получить метаданные топика
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err == nil && len(partitions) > 0 {
		fmt.Println("Topic", topic, "already exists")
		// Топик уже существует
		return nil
	}

	// Топик не существует, создаем его
	return CreateTopic(ctx, brokers, topic, 3, 3)
}

// TopicExists проверяет существование топика
func TopicExists(ctx context.Context, brokers []string, topic string) (bool, error) {
	if len(brokers) == 0 {
		return false, fmt.Errorf("brokers list cannot be empty")
	}

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return false, fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return false, nil
	}

	return len(partitions) > 0, nil
}
