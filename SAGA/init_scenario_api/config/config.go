package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	API      APIConfig
	Producer ProducerConfig
	Database DatabaseConfig
	Pool     PoolConfig
}

type APIConfig struct {
	Port        int
	CORSOrigins []string
}

type ProducerConfig struct {
	Port              int
	KafkaBrokers      []string
	KafkaUsername     string
	KafkaPassword     string
	KafkaAnswersTopic string
	KafkaResultsTopic string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{}

	apiPort, err := getEnvAsInt("API_PORT", 3000)
	if err != nil {
		return nil, fmt.Errorf("invalid API_PORT: %w", err)
	}
	cfg.API.Port = apiPort

	corsOrigins := getEnv("CORS_ORIGINS", "*")
	cfg.API.CORSOrigins = parseList(corsOrigins)

	producerPort, err := getEnvAsInt("PRODUCER_PORT", 3001)
	if err != nil {
		return nil, fmt.Errorf("invalid PRODUCER_PORT: %w", err)
	}
	cfg.Producer.Port = producerPort

	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	cfg.Producer.KafkaBrokers = parseList(kafkaBrokers)

	cfg.Producer.KafkaUsername = getEnv("KAFKA_USERNAME", "")
	cfg.Producer.KafkaPassword = getEnv("KAFKA_PASSWORD", "")
	cfg.Producer.KafkaAnswersTopic = getEnv("KAFKA_ANSWERS_TOPIC", "answers")
	cfg.Producer.KafkaResultsTopic = getEnv("KAFKA_RESULTS_TOPIC", "results")

	cfg.Database.Host = getEnv("DB_HOST", "localhost")

	dbPort, err := getEnvAsInt("DB_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.Database.Port = dbPort

	cfg.Database.User = getEnv("DB_USER", "user")
	cfg.Database.Password = getEnv("DB_PASSWORD", "password")
	cfg.Database.Name = getEnv("DB_NAME", "db")

	// Load Pool configuration
	maxConns, err := getEnvAsInt("DB_POOL_MAX_CONNS", 25)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_POOL_MAX_CONNS: %w", err)
	}
	cfg.Pool.MaxConns = int32(maxConns)

	minConns, err := getEnvAsInt("DB_POOL_MIN_CONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_POOL_MIN_CONNS: %w", err)
	}
	cfg.Pool.MinConns = int32(minConns)

	cfg.Pool.MaxConnLifetime, err = getEnvAsDuration("DB_POOL_MAX_CONN_LIFETIME", 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_POOL_MAX_CONN_LIFETIME: %w", err)
	}

	cfg.Pool.MaxConnIdleTime, err = getEnvAsDuration("DB_POOL_MAX_CONN_IDLE_TIME", 1*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_POOL_MAX_CONN_IDLE_TIME: %w", err)
	}

	cfg.Pool.HealthCheckPeriod, err = getEnvAsDuration("DB_POOL_HEALTH_CHECK_PERIOD", 1*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_POOL_HEALTH_CHECK_PERIOD: %w", err)
	}

	cfg.Pool.ConnectTimeout, err = getEnvAsDuration("DB_POOL_CONNECT_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_POOL_CONNECT_TIMEOUT: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(valueStr))
	if err != nil {
		return 0, err
	}
	return value, nil
}

func parseList(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// getEnvAsDuration retrieves an environment variable as a time.Duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := time.ParseDuration(strings.TrimSpace(valueStr))
	if err != nil {
		return 0, err
	}
	return value, nil
}
