package kafka

import "fmt"

type Config struct {
	Brokers                []string
	ReadTimeout            int
	WriteTimeout           int
	RequiredAcks           int
	Compression            int
	MaxAttempts            int
	Async                  bool
	AllowAutoTopicCreation bool
}

func DefaultConfig(brokers ...string) *Config {
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}
	return &Config{
		Brokers:                brokers,
		ReadTimeout:            10,
		WriteTimeout:           10,
		RequiredAcks:           -1,
		Compression:            0,
		MaxAttempts:            3,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}
}

func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("brokers list cannot be empty")
	}

	for i, broker := range c.Brokers {
		if broker == "" {
			return fmt.Errorf("broker at index %d cannot be empty", i)
		}
	}

	if c.ReadTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative")
	}

	if c.WriteTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative")
	}

	if c.RequiredAcks < -1 || c.RequiredAcks > 1 {
		return fmt.Errorf("required acks must be -1, 0 or 1")
	}

	if c.Compression < 0 || c.Compression > 4 {
		return fmt.Errorf("compression must be between 0 and 4")
	}

	if c.MaxAttempts < 1 {
		return fmt.Errorf("max attempts must be greater than 0")
	}

	return nil
}

func (c *Config) WithRequiredAcks(acks int) *Config {
	c.RequiredAcks = acks
	return c
}

func (c *Config) WithCompression(compression int) *Config {
	c.Compression = compression
	return c
}

func (c *Config) WithMaxAttempts(attempts int) *Config {
	c.MaxAttempts = attempts
	return c
}

func (c *Config) WithAsync(async bool) *Config {
	c.Async = async
	return c
}
