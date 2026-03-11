package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/domain/workflow"
	"github.com/globaltask/bank/internal/infrastructure/repository"
)

type WorkerPool struct {
	eventOutboxRepo repository.EventOutboxRepository
	workflowEngine  *workflow.WorkflowEngine
	concurrency     int
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

func NewWorkerPool(
	eventOutboxRepo repository.EventOutboxRepository,
	workflowEngine *workflow.WorkflowEngine,
	concurrency int,
) *WorkerPool {
	return &WorkerPool{
		eventOutboxRepo: eventOutboxRepo,
		workflowEngine:  workflowEngine,
		concurrency:     concurrency,
		stopCh:          make(chan struct{}),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.concurrency; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
	log.Printf("Started %d workers", wp.concurrency)
}

func (wp *WorkerPool) Stop(ctx context.Context) {
	close(wp.stopCh)
	wp.wg.Wait()
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down...", id)
			return
		case <-wp.stopCh:
			log.Printf("Worker %d stopped", id)
			return
		case <-ticker.C:
			if err := wp.processOne(ctx, id); err != nil {
				log.Printf("Worker %d error: %v", id, err)
			}
		}
	}
}

func (wp *WorkerPool) processOne(ctx context.Context, workerID int) error {
	// Get pending events with SKIP LOCKED
	events, err := wp.eventOutboxRepo.GetPending(ctx, 1)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil // No work to do
	}

	event := events[0]

	// Lock the event
	locked, err := wp.eventOutboxRepo.Lock(ctx, event.ID)
	if err != nil {
		return err
	}
	if !locked {
		return nil // Another worker got it
	}

	// Process based on event type
	log.Printf("Worker %d processing event %s (type: %s)", workerID, event.ID, event.EventType)

	switch event.EventType {
	case entity.EventLoanApplicationCreated:
		err = wp.workflowEngine.HandleLoanApplicationCreated(ctx, event)
	case entity.EventFetchBankData:
		err = wp.workflowEngine.HandleFetchBankData(ctx, event)
	case entity.EventValidateUserIdentity:
		err = wp.workflowEngine.HandleValidateUserIdentity(ctx, event)
	case entity.EventEvaluateApplicationRisk:
		err = wp.workflowEngine.HandleEvaluateRisk(ctx, event)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		err = wp.eventOutboxRepo.MarkFailed(ctx, event.ID, "unknown event type")
	}

	if err != nil {
		log.Printf("Worker %d failed to process event %s: %v", workerID, event.ID, err)
		wp.eventOutboxRepo.MarkFailed(ctx, event.ID, err.Error())
		return err
	}

	// Mark as done
	if err := wp.eventOutboxRepo.MarkDone(ctx, event.ID); err != nil {
		return err
	}

	log.Printf("Worker %d completed event %s", workerID, event.ID)
	return nil
}
