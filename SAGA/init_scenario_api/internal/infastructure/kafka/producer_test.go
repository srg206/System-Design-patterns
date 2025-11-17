package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestKafkaProducerSendJSONMessage(t *testing.T) {
	brokers := []string{"localhost:9092", "localhost:9093", "localhost:9094"}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	cfg := DefaultConfig(brokers...)
	cfg.AllowAutoTopicCreation = false

	producer, err := NewKafkaProducer(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create Kafka producer: %v", err)
	}
	defer func() {
		_ = producer.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создаем топик перед отправкой сообщения
	t.Log("ensuring topic exists...")
	if err := EnsureTopic(ctx, cfg.Brokers, "test-topic"); err != nil {
		t.Fatalf("failed to ensure topic exists: %v", err)
	}

	// Небольшая задержка для синхронизации метаданных кластера
	time.Sleep(100 * time.Millisecond)

	payload := map[string]string{
		"type":    "test",
		"message": "hello, kafka",
	}

	value, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	key := "test-key"
	msg := NewMessage("test-topic", &key, value, nil)

	if err := producer.SendMessage(ctx, msg); err != nil {
		t.Fatalf("failed to send message to Kafka: %v", err)
	}

	t.Log("message sent successfully")
}

func TestKafkaProducerSendMessagesWithoutKey(t *testing.T) {
	brokers := []string{"localhost:9092", "localhost:9093", "localhost:9094"}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	cfg := DefaultConfig(brokers...)
	cfg.AllowAutoTopicCreation = false

	producer, err := NewKafkaProducer(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create Kafka producer: %v", err)
	}
	defer func() {
		_ = producer.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Создаем топик перед отправкой сообщений
	topicName := "test-topic-without-key-1"
	t.Log("ensuring topic exists...")
	if err := EnsureTopic(ctx, cfg.Brokers, topicName); err != nil {
		t.Fatalf("failed to ensure topic exists: %v", err)
	}

	// Небольшая задержка для синхронизации метаданных кластера
	time.Sleep(1000 * time.Millisecond)

	messagesCount := 100
	messages := make([]*Message, 0, messagesCount)

	for i := 0; i < messagesCount; i++ {
		payload := map[string]interface{}{
			"type":      "test",
			"message":   "message without key - should be distributed by balancer",
			"index":     i,
			"timestamp": time.Now().Unix(),
		}

		value, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload for message %d: %v", i, err)
		}

		messages = append(messages, NewMessage(topicName, nil, value, nil))
	}

	startTime := time.Now()

	if err := producer.SendMessages(ctx, messages); err != nil {
		t.Fatalf("failed to send messages batch to Kafka: %v", err)
	}

	duration := time.Since(startTime)

	t.Logf("=== Test Results ===")
	t.Logf("Total messages sent in batch: %d", messagesCount)
	t.Logf("Duration: %v", duration)
	t.Logf("Average time per message: %v", duration/time.Duration(messagesCount))
	t.Logf("Messages are sent without keys and should be distributed across partitions by LeastBytes balancer")
}
