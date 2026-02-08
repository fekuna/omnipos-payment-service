package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	GRPCPort   string
	LoggerLvl  string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		GRPCPort:   getEnv("GRPC_PORT", "50054"), // Priority 4 -> 50054 (Product 50051, Order 50053, Customer 8084(5005x?)) Let's check Customer port. Customer was 8084 http, grpc?
		LoggerLvl:  getEnv("LOGGER_LEVEL", "debug"),
		DBHost:     getEnv("POSTGRES_HOST", "localhost"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:     getEnv("POSTGRES_DB", "omnipos_payment_db"),
		DBSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
