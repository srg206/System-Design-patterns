package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestKafkaConsumerReadMessage(t *testing.T) {
	brokers := []string{"localhost:9092", "localhost:9093", "localhost:9094"}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// Setup producer to send test messages
	producerCfg := DefaultConfig(brokers...)
	producerCfg.AllowAutoTopicCreation = false

	producer, err := NewKafkaProducer(producerCfg, logger)
	if err != nil {
		t.Fatalf("failed to create Kafka producer: %v", err)
	}
	defer func() {
		_ = producer.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Создаем топик перед отправкой сообщения
	topicName := "test-consumer-topic"
	t.Log("ensuring topic exists...")
	if err := EnsureTopic(ctx, brokers, topicName, 3, 3); err != nil {
		t.Fatalf("failed to ensure topic exists: %v", err)
	}

	// Небольшая задержка для синхронизации метаданных кластера
	time.Sleep(100 * time.Millisecond)

	// Send test message
	payload := map[string]string{
		"type":    "test",
		"message": "hello from consumer test",
	}

	value, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	key := "test-consumer-key"
	msg := NewMessage(topicName, &key, value, nil)

	if err := producer.SendMessage(ctx, msg); err != nil {
		t.Fatalf("failed to send message to Kafka: %v", err)
	}

	t.Log("message sent successfully, now creating consumer...")

	// Setup consumer
	consumerCfg := DefaultConfig(brokers...)
	consumerCfg.ConsumerGroup = "test-consumer-group"

	consumer, err := NewKafkaConsumer(consumerCfg, []string{topicName}, logger)
	if err != nil {
		t.Fatalf("failed to create Kafka consumer: %v", err)
	}
	defer func() {
		_ = consumer.Close()
	}()

	t.Log("consumer created, reading message...")

	// Read message
	readMsg, err := consumer.ReadMessage(ctx)
	if err != nil {
		t.Fatalf("failed to read message from Kafka: %v", err)
	}

	t.Logf("message read successfully from topic: %s", readMsg.Topic)

	if readMsg.Key == nil || *readMsg.Key != key {
		t.Fatalf("expected key %s, got %v", key, readMsg.Key)
	}

	var readPayload map[string]string
	if err := json.Unmarshal(readMsg.Value, &readPayload); err != nil {
		t.Fatalf("failed to unmarshal message value: %v", err)
	}

	if readPayload["message"] != payload["message"] {
		t.Fatalf("expected message %s, got %s", payload["message"], readPayload["message"])
	}

	t.Log("message validation passed")
}

func TestKafkaConsumerReadMessages(t *testing.T) {
	brokers := []string{"localhost:9092", "localhost:9093", "localhost:9094"}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// Setup producer to send test messages
	producerCfg := DefaultConfig(brokers...)
	producerCfg.AllowAutoTopicCreation = false

	producer, err := NewKafkaProducer(producerCfg, logger)
	if err != nil {
		t.Fatalf("failed to create Kafka producer: %v", err)
	}
	defer func() {
		_ = producer.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Создаем топик перед отправкой сообщений
	topicName := "test-consumer-batch-topic"
	t.Log("ensuring topic exists...")
	if err := EnsureTopic(ctx, brokers, topicName, 3, 3); err != nil {
		t.Fatalf("failed to ensure topic exists: %v", err)
	}

	// Небольшая задержка для синхронизации метаданных кластера
	time.Sleep(100 * time.Millisecond)

	// Send batch of test messages
	messagesCount := 10
	messages := make([]*Message, 0, messagesCount)

	for i := 0; i < messagesCount; i++ {
		payload := map[string]interface{}{
			"type":      "test",
			"message":   "batch message",
			"index":     i,
			"timestamp": time.Now().Unix(),
		}

		value, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload for message %d: %v", i, err)
		}

		messages = append(messages, NewMessage(topicName, nil, value, nil))
	}

	if err := producer.SendMessages(ctx, messages); err != nil {
		t.Fatalf("failed to send messages batch to Kafka: %v", err)
	}

	t.Logf("sent %d messages successfully, now creating consumer...", messagesCount)

	// Setup consumer
	consumerCfg := DefaultConfig(brokers...)
	consumerCfg.ConsumerGroup = "test-consumer-batch-group"

	consumer, err := NewKafkaConsumer(consumerCfg, []string{topicName}, logger)
	if err != nil {
		t.Fatalf("failed to create Kafka consumer: %v", err)
	}
	defer func() {
		_ = consumer.Close()
	}()

	t.Logf("consumer created, reading %d messages...", messagesCount)

	// Read messages
	readMessages, err := consumer.ReadMessages(ctx, messagesCount)
	if err != nil {
		t.Fatalf("failed to read messages from Kafka: %v", err)
	}

	if len(readMessages) != messagesCount {
		t.Fatalf("expected %d messages, got %d", messagesCount, len(readMessages))
	}

	t.Logf("=== Test Results ===")
	t.Logf("Total messages read in batch: %d", len(readMessages))

	// Verify messages
	for i, msg := range readMessages {
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			t.Fatalf("failed to unmarshal message %d: %v", i, err)
		}

		if payload["type"] != "test" {
			t.Fatalf("message %d: expected type 'test', got '%v'", i, payload["type"])
		}

		t.Logf("Message %d: index=%v, topic=%s", i, payload["index"], msg.Topic)
	}

	t.Log("batch message validation passed")
}
