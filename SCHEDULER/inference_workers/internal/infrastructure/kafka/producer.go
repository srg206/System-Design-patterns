package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer interface {
	SendMessage(ctx context.Context, msg *Message) error
	SendMessages(ctx context.Context, msgs []*Message) error
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewKafkaProducer(cfg *Config, logger *zap.Logger) (Producer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid kafka config: %w", err)
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		ReadTimeout:            time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:           time.Duration(cfg.WriteTimeout) * time.Second,
		RequiredAcks:           kafka.RequiredAcks(cfg.RequiredAcks),
		Compression:            kafka.Compression(cfg.Compression),
		MaxAttempts:            cfg.MaxAttempts,
		Async:                  cfg.Async,
		AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
		Logger:                 kafka.LoggerFunc(logger.Sugar().Debugf),
		ErrorLogger:            kafka.LoggerFunc(logger.Sugar().Errorf),
	}

	logger.Info("kafka producer initialized",
		zap.Strings("brokers", cfg.Brokers),
	)

	return &KafkaProducer{
		writer: writer,
		logger: logger,
	}, nil
}

func (p KafkaProducer) SendMessage(ctx context.Context, msg *Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	var keyBytes []byte
	if msg.Key != nil {
		keyBytes = []byte(*msg.Key)
	} else {
		keyBytes = nil
	}

	kafkaMsg := kafka.Message{
		Topic:   msg.Topic,
		Key:     keyBytes,
		Value:   msg.Value,
		Headers: convertHeaders(msg.Headers),
		Time:    time.Now(),
	}

	err := p.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		logFields := []zap.Field{
			zap.String("topic", msg.Topic),
			zap.Error(err),
		}
		if msg.Key != nil {
			logFields = append(logFields, zap.String("key", *msg.Key))
		}
		p.logger.Error("failed to send message to kafka", logFields...)
		return fmt.Errorf("failed to send message: %w", err)
	}

	logFields := []zap.Field{
		zap.String("topic", msg.Topic),
	}
	if msg.Key != nil {
		logFields = append(logFields, zap.String("key", *msg.Key))
	}
	p.logger.Debug("message sent successfully", logFields...)

	return nil
}

func (p KafkaProducer) SendMessages(ctx context.Context, msgs []*Message) error {
	if len(msgs) == 0 {
		return nil
	}

	for i, msg := range msgs {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}

	kafkaMsgs := make([]kafka.Message, 0, len(msgs))
	now := time.Now()

	for _, msg := range msgs {
		var keyBytes []byte
		if msg.Key != nil {
			keyBytes = []byte(*msg.Key)
		} else {
			keyBytes = nil
		}

		kafkaMsg := kafka.Message{
			Topic:   msg.Topic,
			Key:     keyBytes,
			Value:   msg.Value,
			Headers: convertHeaders(msg.Headers),
			Time:    now,
		}
		kafkaMsgs = append(kafkaMsgs, kafkaMsg)
	}

	err := p.writer.WriteMessages(ctx, kafkaMsgs...)
	if err != nil {
		p.logger.Error("failed to send messages batch to kafka",
			zap.Int("messages_count", len(msgs)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send messages batch: %w", err)
	}

	p.logger.Debug("messages batch sent successfully",
		zap.Int("messages_count", len(msgs)),
	)

	return nil
}

func (p *KafkaProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}

	if err := p.writer.Close(); err != nil {
		if p.logger != nil {
			p.logger.Error("failed to close kafka writer", zap.Error(err))
		}
		return fmt.Errorf("close kafka writer: %w", err)
	}

	if p.logger != nil {
		p.logger.Info("kafka producer closed")
	}

	return nil
}

func convertHeaders(headers map[string][]byte) []kafka.Header {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for key, value := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   key,
			Value: value,
		})
	}
	return kafkaHeaders
}
