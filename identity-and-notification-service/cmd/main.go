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

	grpcapi "github.com/nurashi/car-rental-identity/internal/api/grpc"
	httpapi "github.com/nurashi/car-rental-identity/internal/api/http"
	"github.com/nurashi/car-rental-identity/internal/api/grpc/pb"
	"github.com/nurashi/car-rental-identity/internal/config"
	"github.com/nurashi/car-rental-identity/internal/db"
	"github.com/nurashi/car-rental-identity/internal/db/migration"
	"github.com/nurashi/car-rental-identity/internal/messaging"
	"github.com/nurashi/car-rental-identity/internal/repository"
	"github.com/nurashi/car-rental-identity/internal/service"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()

	pgDB, err := db.NewPostgresDB(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pgDB.Close()

	if err := migration.RunMigrations(ctx, pgDB.Pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	_, _ = db.NewRedisDB(ctx, cfg.RedisAddr(), cfg.RedisPassword)

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

	userRepo := repository.NewUserRepo(pgDB.Pool)
	licenseRepo := repository.NewLicenseRepo(pgDB.Pool)
	emailVerRepo := repository.NewEmailVerificationRepo(pgDB.Pool)
	notifRepo := repository.NewNotificationRepo(pgDB.Pool)

	authService := service.NewAuthService(userRepo, emailVerRepo, service.Config{
		JWTSecret:      cfg.JWTSecret,
		JWTExpiryHours: cfg.JWTExpiryHours,
	})
	userService := service.NewUserService(userRepo)
	licenseService := service.NewLicenseService(licenseRepo)

	grpcHandler := grpcapi.NewServer(authService, userService, licenseService, notifRepo)
	httpHandler := httpapi.NewHandler(authService, userService)

	if pub != nil {
		eventPub := messaging.NewEventPublisher(pub)
		_ = eventPub

		if sub != nil {
			messaging.SetupInboundHandlers(sub, func(subject string, data []byte) {
				log.Printf("handling inbound event: %s", subject)
			})
		}
	}

	go func() {
		addr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen on %s: %v", addr, err)
		}

		s := grpc.NewServer()
		pb.RegisterIdentityServiceServer(s, grpcHandler)

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

	log.Println("identity service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
}
