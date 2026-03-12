package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/globaltask/bank/internal/domain/workflow"
	"github.com/globaltask/bank/internal/infrastructure/auth"
	"github.com/globaltask/bank/internal/infrastructure/database"
	"github.com/globaltask/bank/internal/infrastructure/repository"
	"github.com/globaltask/bank/internal/infrastructure/worker"
)

func main() {
	// ==========================================
	// 1. Configuration
	// ==========================================
	concurrency := 5
	if c := os.Getenv("WORKER_CONCURRENCY"); c != "" {
		if parsed, err := parseInt(c); err == nil {
			concurrency = parsed
		}
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "debug"
	}

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
	// 3. Repositories
	// ==========================================
	dbPool := db.Pool()
	loanAppRepo := repository.NewLoanApplicationRepository(dbPool)
	eventOutboxRepo := repository.NewEventOutboxRepository(dbPool)
	countryRepo := repository.NewCountryRepository(dbPool)
	bankProviderRepo := repository.NewBankProviderRepository(dbPool)
	uow := repository.NewPgUnitOfWork(dbPool)

	identityService := auth.NewSupabaseIdentityService()

	workflowEngine := workflow.NewWorkflowEngine(
		uow,
		loanAppRepo,
		countryRepo,
		bankProviderRepo,
		eventOutboxRepo,
		identityService,
	)

	// ==========================================
	// 5. Worker Pool
	// ==========================================
	workerPool := worker.NewWorkerPool(
		eventOutboxRepo,
		workflowEngine,
		concurrency,
	)

	log.Printf("🚀 Starting worker pool with %d concurrent workers", concurrency)

	// Start the worker pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workerPool.Start(ctx)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down workers...")

	// Graceful shutdown - wait for current jobs to finish
	gracefulCtx, gracefulCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer gracefulCancel()

	workerPool.Stop(gracefulCtx)
	log.Println("✅ Workers stopped")
}

// Helper to parse int (simple version to avoid imports)
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
