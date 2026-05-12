package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	grpcapi "github.com/nurashi/car-rental-booking/internal/api/grpc"
	"github.com/nurashi/car-rental-booking/internal/api/grpc/pb"
	httpapi "github.com/nurashi/car-rental-booking/internal/api/http"
	"github.com/nurashi/car-rental-booking/internal/config"
	"github.com/nurashi/car-rental-booking/internal/db"
	"github.com/nurashi/car-rental-booking/internal/db/migration"
	"github.com/nurashi/car-rental-booking/internal/messaging"
	"github.com/nurashi/car-rental-booking/internal/repository"
	"github.com/nurashi/car-rental-booking/internal/service"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	pgDB, err := db.NewPostgresDB(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pgDB.Close()

	if err := migration.RunMigrations(ctx, pgDB.Pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	// Redis is used for pricing tier cache (cache-aside pattern)
	var redisClient *redis.Client
	redisDB, err := db.NewRedisDB(ctx, cfg.RedisAddr(), cfg.RedisPassword)
	if err != nil {
		log.Printf("redis not available (pricing cache disabled): %v", err)
		// Use a no-op client so the service degrades gracefully
		redisClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr()})
	} else {
		redisClient = redisDB.Client
		defer redisDB.Close()
	}

	// ── Vehicle inventory gRPC client ─────────────────────────────────────────
	var vehicleClient service.VehicleServiceClient
	vc, err := service.NewVehicleClient(cfg.VehicleServiceAddr)
	if err != nil {
		log.Printf("vehicle service not reachable (degraded mode): %v", err)
		vehicleClient = &service.NoopVehicleClient{}
	} else {
		vehicleClient = vc
		defer vc.Close()
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	bookingRepo := repository.NewBookingRepo(pgDB.Pool)
	pricingRepo := repository.NewPricingRepo(pgDB.Pool)
	refundRepo := repository.NewRefundRepo(pgDB.Pool)

	// ── Business logic ────────────────────────────────────────────────────────
	bookingSvc := service.NewBookingService(
		bookingRepo,
		pricingRepo,
		refundRepo,
		redisClient,
		vehicleClient,
	)

	// ── NATS Messaging ────────────────────────────────────────────────────────
	pub, err := messaging.NewPublisher(cfg.NATSURL)
	if err != nil {
		log.Printf("nats publisher not available: %v", err)
		pub = nil
	} else {
		defer pub.Close()
	}

	sub, err := messaging.NewSubscriber(cfg.NATSURL)
	if err != nil {
		log.Printf("nats subscriber not available: %v", err)
		sub = nil
	} else {
		defer sub.Close()
	}

	if pub != nil {
		_ = messaging.NewEventPublisher(pub)
	}

	if sub != nil {
		messaging.SetupInboundHandlers(sub, bookingSvc.HandleVehicleMaintenanceEvent)
	}

	// ── gRPC server ───────────────────────────────────────────────────────────
	grpcHandler := grpcapi.NewServer(bookingSvc)
	httpHandler := httpapi.NewHandler(bookingSvc, cfg.JWTSecret)

	go func() {
		addr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen on %s: %v", addr, err)
		}

		s := grpc.NewServer()
		pb.RegisterBookingServiceServer(s, grpcHandler)

		log.Printf("grpc server listening on %s", addr)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("grpc server failed: %v", err)
		}
	}()

	// ── HTTP server ───────────────────────────────────────────────────────────
	go func() {
		addr := ":" + cfg.HTTPPort
		log.Printf("http server listening on %s", addr)
		if err := httpHandler.Engine().Run(addr); err != nil {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	log.Println("booking-and-pricing service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
}
