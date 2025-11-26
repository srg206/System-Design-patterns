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
	ReadMessages(ctx context.Context, n int) ([]*Message, error)
	Close() error
}

type KafkaConsumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

func NewKafkaConsumer(cfg *Config, topics []string, logger *zap.Logger) (Consumer, error) {
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
		GroupID:        cfg.ConsumerGroup,
		GroupTopics:    topics,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
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
	kafkaMsg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		c.logger.Error("failed to read message from kafka",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

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
	c.logger.Debug("message read successfully", logFields...)

	return msg, nil
}

func (c *KafkaConsumer) ReadMessages(ctx context.Context, n int) ([]*Message, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be greater than 0")
	}

	messages := make([]*Message, 0, n)

	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			c.logger.Warn("context cancelled while reading messages",
				zap.Int("read", i),
				zap.Int("requested", n),
			)
			if len(messages) > 0 {
				return messages, nil
			}
			return nil, ctx.Err()
		default:
		}

		kafkaMsg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			c.logger.Error("failed to read message from kafka",
				zap.Int("message_index", i),
				zap.Error(err),
			)
			// If we've already read some messages, return them
			if len(messages) > 0 {
				return messages, nil
			}
			return nil, fmt.Errorf("failed to read message at index %d: %w", i, err)
		}

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

		messages = append(messages, msg)
	}

	c.logger.Debug("messages batch read successfully",
		zap.Int("messages_count", len(messages)),
	)

	return messages, nil
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
