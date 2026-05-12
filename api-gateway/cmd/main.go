package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/nurashi/car-rental-gateway/internal/config"
	"github.com/nurashi/car-rental-gateway/internal/health"
	"github.com/nurashi/car-rental-gateway/internal/middleware"
	"github.com/nurashi/car-rental-gateway/internal/proxy"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	r := gin.Default()

	r.Use(middleware.CORS())

	proxyHandler := proxy.NewHandler(cfg.IdentityServiceAddr, cfg.VehicleServiceAddr, cfg.BookingServiceAddr)
	proxyHandler.SetupRoutes(r)

	healthHandler := health.New(map[string]string{
		"identity": "http://" + cfg.IdentityServiceAddr,
		"vehicle":  "http://" + cfg.VehicleServiceAddr,
		"booking":  "http://" + cfg.BookingServiceAddr,
	})
	healthHandler.SetupRoutes(r)

	addr := ":" + cfg.Port
	log.Printf("api-gateway listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start gateway: %v", err)
	}
}
