package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/api"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/config"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/db"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/messaging"
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

	// Initialize Postgres
	postgresDB, err := db.NewPostgresDB(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer postgresDB.Close()

	// Initialize Redis
	redisDB, err := db.NewRedisDB(ctx, cfg.RedisAddr(), cfg.RedisPassword)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisDB.Close()

	// Initialize NATS
	natsClient, err := messaging.NewNATSClient(cfg.NATSURL)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	// Initialize Repositories
	vehicleRepo := repository.NewPostgresVehicleRepository(postgresDB.Pool)
	cacheRepo := repository.NewRedisCacheRepository(redisDB.Client, time.Minute*10)

	// Initialize Service
	vehicleSvc := service.NewVehicleService(vehicleRepo, cacheRepo, natsClient)

	// Initialize gRPC Handler
	handler := api.NewVehicleInventoryHandler(vehicleSvc)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterVehicleInventoryServiceServer(grpcServer, handler)

	log.Printf("Vehicle Inventory gRPC service listening on port %s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
