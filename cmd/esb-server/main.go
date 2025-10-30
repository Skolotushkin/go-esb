package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-esb/internal/config"
	"go-esb/internal/database"
	"go-esb/internal/handler"
	"go-esb/internal/repository"
	"go-esb/internal/service"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)
	defer db.Close()

	database.RunMigrations(db)

	log.Println("💫 Go ESB is initialized and database is ready")

	// Инициализация репозиториев
	systemRepo := repository.NewSystemRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	threadRepo := repository.NewThreadRepository(db)
	threadRouteRepo := repository.NewThreadRouteRepository(db)
	connectionRepo := repository.NewConnectionRepository(db)

	// Инициализация сервисов
	messageService := service.NewMessageService(
		threadRouteRepo,
		routeRepo,
		connectionRepo,
		systemRepo,
	)

	orchestrator := service.NewOrchestrator(
		messageService,
		threadRouteRepo,
		routeRepo,
		connectionRepo,
		systemRepo,
	)

	// Инициализация HTTP обработчика
	httpHandler := handler.NewHTTPHandler(messageService, orchestrator)
	router := httpHandler.SetupRoutes()

	// Настройка HTTP сервера
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Go ESB server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	log.Println("✅ Go ESB server is running")
	log.Println("📡 Available endpoints:")
	log.Println("   GET  /health")
	log.Println("   POST /api/v1/messages/process/{threadId}")
	log.Println("   POST /api/v1/orchestrate/{processName}")
	log.Println("   POST /api/v1/webhooks/stripe")

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
