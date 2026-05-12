package config

import (
	"os"
)

type Config struct {
	Port string

	IdentityServiceAddr string
	VehicleServiceAddr  string
	BookingServiceAddr  string

	JWTSecret string
}

func Load() Config {
	return Config{
		Port: getEnv("PORT", "8080"),

		IdentityServiceAddr: getEnv("IDENTITY_SERVICE_ADDR", "localhost:8080"),
		VehicleServiceAddr:  getEnv("VEHICLE_SERVICE_ADDR", "localhost:8081"),
		BookingServiceAddr:  getEnv("BOOKING_SERVICE_ADDR", "localhost:8082"),

		JWTSecret: getEnv("JWT_SECRET", "default-secret"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
