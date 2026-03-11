package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globaltask/bank/internal/application/service"
	"github.com/globaltask/bank/internal/domain/workflow"
	"github.com/globaltask/bank/internal/infrastructure/database"
	"github.com/globaltask/bank/internal/infrastructure/handler"
	"github.com/globaltask/bank/internal/infrastructure/middleware"
	"github.com/globaltask/bank/internal/infrastructure/repository"
	"github.com/globaltask/bank/internal/infrastructure/websocket"
)

func main() {
	// ==========================================
	// 1. Configuration
	// ==========================================
	port := os.Getenv("API_PORT")
	logLevel := os.Getenv("LOG_LEVEL")

	// ==========================================
	// 2. Database Connection
	// ==========================================
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected")

	// ==========================================
	// 3. Repositories (Infrastructure Layer)
	// ==========================================
	dbPool := db.Pool()
	loanAppRepo := repository.NewLoanApplicationRepository(dbPool)
	countryRepo := repository.NewCountryRepository(dbPool)
	bankProviderRepo := repository.NewBankProviderRepository(dbPool)
	eventOutboxRepo := repository.NewEventOutboxRepository(dbPool)
	uow := repository.NewPgUnitOfWork(dbPool)

	// ==========================================
	// 3.5 WebSocket Hub
	// ==========================================
	hub := websocket.NewHub()
	go hub.Run()

	// Listen for Postgres notifications and broadcast to WebSocket hub
	go database.ListenForNotifications(context.Background(), dbPool, "loan_application_updates", func(payload []byte) {
		hub.Broadcast(payload)
	})

	// ==========================================
	// 4. Services (Application Layer)
	// ==========================================
	workflowEngine := workflow.NewWorkflowEngine(
		uow,
		loanAppRepo,
		countryRepo,
		bankProviderRepo,
		eventOutboxRepo,
	)

	loanService := service.NewLoanApplicationService(
		uow,
		loanAppRepo,
		countryRepo,
		bankProviderRepo,
		eventOutboxRepo,
		workflowEngine,
	)

	// ==========================================
	// 5. Handlers (Interface Layer)
	// ==========================================
	loanHandler := handler.NewLoanApplicationHandler(loanService)

	// ==========================================
	// 6. Router Setup
	// ==========================================
	if logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// WebSocket endpoint (no JWT for initial connection, or JWT in query param if needed)
		v1.GET("/ws", handler.ServeWS(hub))

		v1.Use(middleware.JWTAuth()) // All other routes require JWT
		// Loan Applications
		applications := v1.Group("/applications")
		{
			applications.POST("", loanHandler.CreateApplication)
			applications.GET("", loanHandler.ListApplications)
			applications.GET("/:id", loanHandler.GetApplication)
			applications.PATCH("/:id", loanHandler.UpdateApplication)
		}

		// Countries
		countries := v1.Group("/countries")
		{
			countries.GET("", handler.ListCountries(countryRepo))
			countries.GET("/:id", handler.GetCountry(countryRepo))
		}
	}

	// Webhooks (external systems - no JWT, but signature verification)
	webhooks := router.Group("/webhook")
	{
		webhooks.POST("/bank-update", loanHandler.HandleBankWebhook)
	}

	// ==========================================
	// 7. Server Setup
	// ==========================================
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}
