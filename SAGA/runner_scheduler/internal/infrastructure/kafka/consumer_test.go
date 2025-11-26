package kafka

import (
	"context"
	"testing"
	"time"

	modelKafka "runner_scheduler/internal/models/kafka"

	"go.uber.org/zap"
)

func TestKafkaConsumerReadMessage(t *testing.T) {
	brokers := []string{"localhost:9092", "localhost:9093", "localhost:9094"}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	topicName := modelKafka.OutboxScenarioApi

	t.Log("message sent successfully, now creating consumer...")

	// Setup consumer
	consumerCfg := DefaultConfig(brokers...)
	consumerCfg.ConsumerGroup = modelKafka.KafkaConsumerGroup
	consumerCfg.AllowAutoTopicCreation = false

	consumer, err := NewKafkaConsumer(consumerCfg, []string{topicName}, modelKafka.KafkaConsumerGroup, logger)
	if err != nil {
		t.Fatalf("failed to create Kafka consumer: %v", err)
	}
	defer func() {
		_ = consumer.Close()
	}()

	t.Log("consumer created, reading message...")

	time.Sleep(5000 * time.Millisecond)
	// Read message
	readMsg, err := consumer.ReadMessage(context.Background())
	if err != nil {
		t.Fatalf("failed to read message from Kafka: %v", err)
	}

	t.Logf("message read successfully from topic: %s", readMsg.Topic)

	// Выводим детальную информацию о сообщении
	t.Log("=== Message Details ===")
	t.Logf("Topic: %s", readMsg.Topic)

	if readMsg.Key != nil {
		t.Logf("Key: %s", *readMsg.Key)
	} else {
		t.Log("Key: <nil>")
	}

	t.Logf("Value: %s", string(readMsg.Value))

	if len(readMsg.Headers) > 0 {
		t.Log("Headers:")
		for k, v := range readMsg.Headers {
			t.Logf("  %s: %s", k, v)
		}
	} else {
		t.Log("Headers: <empty>")
	}

	t.Log("message validation passed")
}
