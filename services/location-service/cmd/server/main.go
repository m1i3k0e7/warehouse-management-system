package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/infrastructure/database"
	grpc_server "github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/interfaces/grpc"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/interfaces/grpc/handlers"
)

func main() {
	// Configuration
	port := getEnv("PORT", ":50052")
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("DB_NAME", "location_service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize MongoDB repository
	repo, err := database.NewMongoRepository(ctx, mongoURI, dbName)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	defer func() {
		if err := repo.Disconnect(ctx); err != nil {
			log.Printf("failed to disconnect from mongo: %v", err)
		}
	}()

	// Create the location server
	locationServer := handlers.NewLocationServer(repo, repo) // repo implements both interfaces

	// Create and start the gRPC server
	server := grpc_server.NewServer(port, locationServer)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	server.Stop()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}