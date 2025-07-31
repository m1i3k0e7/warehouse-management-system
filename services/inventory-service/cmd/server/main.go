package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"WMS/services/inventory-service/internal/config"
	"WMS/services/inventory-service/internal/infrastructure/cache"
	"WMS/services/inventory-service/internal/infrastructure/database"
	"WMS/services/inventory-service/internal/application/commands"
	"WMS/services/inventory-service/internal/application/queries"
	"WMS/services/inventory-service/internal/domain/repositories"
	"WMS/services/inventory-service/internal/domain/services"
	"WMS/services/inventory-service/internal/interfaces/http/handlers"
	"WMS/services/inventory-service/internal/interfaces/http/router"
	"WMS/services/inventory-service/internal/interfaces/mqtt"
	"WMS/services/inventory-service/pkg/utils/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.LogLevel)

	// Initialize database connection
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Redis client
	redisClient := cache.NewRedisClient(cfg.Redis)
	defer redisClient.Close()

	// Initialize Kafka event service
	eventService, err := services.NewEventService(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Fatal("Failed to initialize event service:", err)
	}
	defer eventService.Close()

	// Initialize all other services
	lockService := services.NewLockService(redisClient)
	cacheService := services.NewCacheService(redisClient)
	retryService := services.NewRetryService(cfg.ServiceConfig.RetryCount, cfg.ServiceConfig.RetryDelay)
	auditService := services.NewAuditService(eventService)
	alertService := services.NewAlertService(eventService)

	// Initialize repositories
	materialRepo := repositories.NewMaterialRepository(db)
	slotRepo := repositories.NewSlotRepository(db)
	operationRepo := repositories.NewOperationRepository(db)
	alertRepo := repositories.NewAlertRepository(db)
	failedEventRepo := repositories.NewFailedEventRepository(db)

	// Initialize inventory service
	inventoryService := services.NewInventoryService(
		materialRepo,
		slotRepo,
		operationRepo,
		alertRepo,
		lockService,
		eventService,
		cacheService,
		auditService,
		alertService,
		failedEventRepo,
	)

	// Initialize command and query handlers
	placeMaterialHandler := commands.NewPlaceMaterialCommandHandler(inventoryService)
	removeMaterialHandler := commands.NewRemoveMaterialCommandHandler(inventoryService)
	moveMaterialHandler := commands.NewMoveMaterialCommandHandler(inventoryService)
	reserveSlotsHandler := commands.NewReserveSlotsCommandHandler(inventoryService)
	batchPlaceMaterialsHandler := commands.NewBatchPlaceMaterialsCommandHandler(inventoryService)
	handleSlotErrorHandler := commands.NewHandleSlotErrorCommandHandler(inventoryService)
	updateShelfStatusHandler := commands.NewUpdateShelfStatusCommandHandler(inventoryService)

	getShelfStatusHandler := queries.NewGetShelfStatusQueryHandler(inventoryService)
	findOptimalSlotHandler := queries.NewFindOptimalSlotQueryHandler(inventoryService)
	searchMaterialsHandler := queries.NewSearchMaterialsQueryHandler(inventoryService)
	healthCheckShelfHandler := queries.NewHealthCheckShelfQueryHandler(inventoryService)
	getOperationsHandler := queries.NewGetOperationsQueryHandler(operationRepo)

	// Initialize MQTT handler
	mqttHandler := mqtt.NewMQTTHandler(
		cfg.MQTT.BrokerURL,
		placeMaterialHandler,
		removeMaterialHandler,
		handleSlotErrorHandler,
		updateShelfStatusHandler,
		inventoryService, // Pass inventoryService here
		retryService,
	)
	if err := mqttHandler.Connect(); err != nil {
		log.Fatal("Failed to connect to MQTT broker:", err)
	}

	// Timeout Scheduler for Pending Physical Confirmations
	go func() {
		ticker := time.NewTicker(cfg.Service.PhysicalOperationTimeoutCheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			// Query operations that are pending physical confirmation and have timed out
			timedOutPlacementOperations, err := operationRepo.GetTimedOutPendingPhysicalConfirmations(ctx, cfg.Service.PhysicalOperationTimeout)
			if err != nil {
				logger.Error("Failed to query timed out operations", err)
			} else {
				for _, op := range timedOutPlacementOperations {
					logger.Warn(fmt.Sprintf("Physical placement operation %s timed out. Initiating rollback.", op.ID))
					if err := inventoryService.HandlePhysicalPlacementTimeout(ctx, op.ID); err != nil {
						logger.Error(fmt.Sprintf("Failed to handle timeout for operation %s", op.ID), err)
					}
				}
			}

			timeOutRemovalOperations, err := operationRepo.GetTimedOutPendingRemovalConfirmations(ctx, cfg.Service.PhysicalOperationTimeout)
			if err != nil {
				logger.Error("Failed to query timed out removal operations", err)
				continue
			}
			for _, op := range timeOutRemovalOperations {
				logger.Warn(fmt.Sprintf("Physical removal operation %s timed out. Initiating rollback.", op.ID))
				if err := inventoryService.HandlePhysicalRemovalTimeout(ctx, op.ID); err != nil {
					logger.Error(fmt.Sprintf("Failed to handle timeout for removal operation %s", op.ID), err)
				}
			}
		}
	}()

	// Initialize HTTP handlers
	materialHandler := handlers.NewMaterialHandler(placeMaterialHandler, removeMaterialHandler, moveMaterialHandler, searchMaterialsHandler)
	slotHandler := handlers.NewSlotHandler(reserveSlotsHandler, findOptimalSlotHandler, getShelfStatusHandler, healthCheckShelfHandler)
	operationHandler := handlers.NewOperationHandler(getOperationsHandler)

	// Initialize http router
	gin.SetMode(cfg.Server.Mode)
	r := router.SetupRoutes(gin.Default(), materialHandler, slotHandler, operationHandler)

	// configure http server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// start the server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Info("Inventory service started successfully")

	// handle shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}