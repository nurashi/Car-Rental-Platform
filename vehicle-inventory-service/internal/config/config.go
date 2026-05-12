package config

import (
	"os"
)

type Config struct {
	HTTPPort string
	GRPCPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	RedisHost     string
	RedisPort     string
	RedisPassword string

	NATSURL   string
	JWTSecret string
}

func Load() Config {
	return Config{
		HTTPPort: getEnv("HTTP_PORT", "8081"),
		GRPCPort: getEnv("GRPC_PORT", "50052"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "vehicle_db"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		NATSURL:   getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret: getEnv("JWT_SECRET", "default-secret"),
	}
}

func (c Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

func (c Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
