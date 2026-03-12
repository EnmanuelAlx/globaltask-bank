package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/domain/workflow"
	"github.com/globaltask/bank/internal/infrastructure/repository"
)

type WorkerPool struct {
	uowManager     repository.UnitOfWork
	workflowEngine *workflow.WorkflowEngine
	concurrency    int
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

func NewWorkerPool(
	uowManager repository.UnitOfWork,
	workflowEngine *workflow.WorkflowEngine,
	concurrency int,
) *WorkerPool {
	return &WorkerPool{
		uowManager:     uowManager,
		workflowEngine: workflowEngine,
		concurrency:    concurrency,
		stopCh:         make(chan struct{}),
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
	return wp.uowManager.Do(ctx, func(uow repository.UnitOfWork) error {
		// a. Claim one event atomically
		events, err := uow.EventOutbox().Claim(ctx, 1)
		if err != nil {
			return err
		}

		// b. If no events, return nil to end transaction early
		if len(events) == 0 {
			return nil
		}

		event := events[0]

		// c. Log worker ID and event processing start
		log.Printf("Worker %d processing event %s (type: %s)", workerID, event.ID, event.EventType)

		// d. Pass 'uow' into handlers
		switch event.EventType {
		case entity.EventLoanApplicationCreated:
			err = wp.workflowEngine.HandleLoanApplicationCreated(ctx, uow, event)
		case entity.EventFetchBankData:
			err = wp.workflowEngine.HandleFetchBankData(ctx, uow, event)
		case entity.EventValidateUserIdentity:
			err = wp.workflowEngine.HandleValidateUserIdentity(ctx, uow, event)
		case entity.EventEvaluateApplicationRisk:
			err = wp.workflowEngine.HandleEvaluateRisk(ctx, uow, event)
		default:
			log.Printf("Unknown event type: %s", event.EventType)
			return fmt.Errorf("unknown event type: %s", event.EventType)
		}

		if err != nil {
			// e. Return error for rollback
			return err
		}

		// f. Mark as done
		return uow.EventOutbox().MarkDone(ctx, event.ID)
	})
}
