package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Kafka       KafkaConfig
	LogLevel    string
	Service     ServiceConfig
	MQTT		MQTTConfig
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

type MQTTConfig struct {
	BrokerURL string
}

type ServiceConfig struct {
	RetryCount                         int
	RetryDelay                         time.Duration
	AllowedOrigins                       string
	PhysicalOperationTimeout        time.Duration
	PhysicalOperationTimeoutCheckInterval time.Duration
}

func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "warehouse"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:   getEnv("KAFKA_TOPIC", "inventory_events"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Service: ServiceConfig{
			RetryCount:                         parseInt(getEnv("RETRY_COUNT", "5")),
			RetryDelay:                         parseDuration(getEnv("RETRY_DELAY", "2s")),
			AllowedOrigins:                       getEnv("ALLOW_ORIGINS", "*"),
			PhysicalOperationTimeout:        parseDuration(getEnv("PHYSICAL_OPERATION_TIMEOUT", "5m")),
			PhysicalOperationTimeoutCheckInterval: parseDuration(getEnv("PHYSICAL_OPERATION_TIMEOUT_CHECK_INTERVAL", "1m")),
		},
		MQTT: MQTTConfig{
			BrokerURL: getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0 // Default or handle error
	}
	return val
}

func parseDuration(s string) time.Duration {
	val, err := time.ParseDuration(s)
	if err != nil {
		return 0 // Default or handle error
	}
	return val
}
