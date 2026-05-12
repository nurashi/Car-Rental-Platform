package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/api"
	httpapi "github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/api/http"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/config"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/db"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/db/migration"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/messaging"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/metrics"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/repository"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/service"
	pb "github.com/nurashi/Car-Rental-Platform/vehicle-inventory/proto"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := config.Load()
	ctx := context.Background()

	postgresDB, err := db.NewPostgresDB(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer postgresDB.Close()

	if err := migration.RunMigrations(ctx, postgresDB.Pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	redisDB, err := db.NewRedisDB(ctx, cfg.RedisAddr(), cfg.RedisPassword)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisDB.Close()

	natsClient, err := messaging.NewNATSClient(cfg.NATSURL)
	if err != nil {
		log.Printf("nats not available: %v", err)
		natsClient = nil
	}
	if natsClient != nil {
		defer natsClient.Close()
	}

	vehicleRepo := repository.NewPostgresVehicleRepository(postgresDB.Pool)
	cacheRepo := repository.NewRedisCacheRepository(redisDB.Client, time.Minute*10)

	vehicleSvc := service.NewVehicleService(vehicleRepo, cacheRepo, natsClient)

	grpcHandler := api.NewVehicleInventoryHandler(vehicleSvc)
	httpHandler := httpapi.NewHandler(vehicleSvc, cfg.JWTSecret)
	httpHandler.Engine().Use(metrics.Middleware())
	httpHandler.Engine().GET("/metrics", metrics.Handler())

	go func() {
		addr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen on %s: %v", addr, err)
		}

		s := grpc.NewServer()
		pb.RegisterVehicleInventoryServiceServer(s, grpcHandler)

		log.Printf("grpc server listening on %s", addr)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("grpc server failed: %v", err)
		}
	}()

	go func() {
		addr := ":" + cfg.HTTPPort
		log.Printf("http server listening on %s", addr)
		if err := httpHandler.Engine().Run(addr); err != nil {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	log.Println("vehicle-inventory service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
}
