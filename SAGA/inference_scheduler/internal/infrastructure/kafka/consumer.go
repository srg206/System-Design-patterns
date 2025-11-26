package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer interface {
	ReadMessage(ctx context.Context) (*Message, error)
	CommitMessages(ctx context.Context, msgs ...*Message) error
	Close() error
}

type KafkaConsumer struct {
	reader      *kafka.Reader
	logger      *zap.Logger
	lastMessage *kafka.Message
}

func NewKafkaConsumer(cfg *Config, topics []string, consumerGroup string, logger *zap.Logger) (Consumer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid kafka config: %w", err)
	}

	if len(topics) == 0 {
		return nil, fmt.Errorf("topics list cannot be empty")
	}

	if cfg.ConsumerGroup == "" {
		return nil, fmt.Errorf("consumer group cannot be empty")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        consumerGroup,
		GroupTopics:    topics,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
		CommitInterval: 0, // Отключаем автоматический commit
		Logger:         kafka.LoggerFunc(logger.Sugar().Debugf),
		ErrorLogger:    kafka.LoggerFunc(logger.Sugar().Errorf),
	})

	logger.Info("kafka consumer initialized",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("consumer_group", cfg.ConsumerGroup),
		zap.Strings("topics", topics),
	)

	return &KafkaConsumer{
		reader: reader,
		logger: logger,
	}, nil
}

func (c *KafkaConsumer) ReadMessage(ctx context.Context) (*Message, error) {
	kafkaMsg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		c.logger.Error("failed to fetch message from kafka",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}

	// Сохраняем последнее сообщение для commit
	c.lastMessage = &kafkaMsg

	var key *string
	if len(kafkaMsg.Key) > 0 {
		keyStr := string(kafkaMsg.Key)
		key = &keyStr
	}

	headers := make(map[string][]byte)
	for _, header := range kafkaMsg.Headers {
		headers[header.Key] = header.Value
	}

	msg := &Message{
		Topic:   kafkaMsg.Topic,
		Key:     key,
		Value:   kafkaMsg.Value,
		Headers: headers,
	}

	logFields := []zap.Field{
		zap.String("topic", msg.Topic),
		zap.Int("partition", kafkaMsg.Partition),
		zap.Int64("offset", kafkaMsg.Offset),
	}
	if msg.Key != nil {
		logFields = append(logFields, zap.String("key", *msg.Key))
	}
	c.logger.Debug("message fetched successfully", logFields...)

	return msg, nil
}

func (c *KafkaConsumer) CommitMessages(ctx context.Context, msgs ...*Message) error {
	if c.lastMessage == nil {
		return fmt.Errorf("no message to commit")
	}

	if err := c.reader.CommitMessages(ctx, *c.lastMessage); err != nil {
		c.logger.Error("failed to commit message",
			zap.Error(err),
			zap.String("topic", c.lastMessage.Topic),
			zap.Int("partition", c.lastMessage.Partition),
			zap.Int64("offset", c.lastMessage.Offset),
		)
		return fmt.Errorf("failed to commit message: %w", err)
	}

	c.logger.Debug("message committed successfully",
		zap.String("topic", c.lastMessage.Topic),
		zap.Int("partition", c.lastMessage.Partition),
		zap.Int64("offset", c.lastMessage.Offset),
	)

	return nil
}

func (c *KafkaConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}

	if err := c.reader.Close(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to close kafka reader", zap.Error(err))
		}
		return fmt.Errorf("close kafka reader: %w", err)
	}

	if c.logger != nil {
		c.logger.Info("kafka consumer closed")
	}

	return nil
}
