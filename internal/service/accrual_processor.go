package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/onbehalfofhim/gofermart/internal/client"
	"github.com/onbehalfofhim/gofermart/internal/logger"
	"github.com/onbehalfofhim/gofermart/internal/models"
)

type AccrualProcessor struct {
	orderService   *OrderService
	balanceService *BalanceService
	accrualClient  *client.AccrualClient

	logger *logger.Logger

	jobs         chan models.Order
	workerCount  int
	pollInterval time.Duration

	// Rate limiting состояние
	retryMu              sync.RWMutex // защита от race condition при чтении/записи nextAllowedRequestAt
	nextAllowedRequestAt time.Time    // момент времени, после которого можно снова отправлять запросы в accrual-систему

	wg sync.WaitGroup
}

func NewAccrualProcessor(orderService *OrderService, balanceService *BalanceService, accrualClient *client.AccrualClient, logger *logger.Logger) *AccrualProcessor {
	return &AccrualProcessor{
		orderService:         orderService,
		balanceService:       balanceService,
		accrualClient:        accrualClient,
		logger:               logger,
		jobs:                 make(chan models.Order, 100),
		workerCount:          5,
		pollInterval:         5 * time.Second,
		nextAllowedRequestAt: time.Time{},
	}
}

func (p *AccrualProcessor) Start(ctx context.Context) {
	p.logger.Info("starting accrual processor")

	// стартуем workers
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	// scheduler loop
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("stopping accrual processor")
			close(p.jobs)
			p.wg.Wait()
			p.logger.Info("all workers stopped")
			return
		case <-ticker.C:
			p.scheduleProcessingOrders(ctx)
		}
	}
}

func (p *AccrualProcessor) canProcess() bool {
	p.retryMu.RLock()
	defer p.retryMu.RUnlock()

	return time.Now().After(p.nextAllowedRequestAt)
}

func (p *AccrualProcessor) setRetryAfter(duration time.Duration) {
	p.retryMu.Lock()
	defer p.retryMu.Unlock()

	next := time.Now().Add(duration)
	if next.After(p.nextAllowedRequestAt) {
		p.nextAllowedRequestAt = next
	}
}

func (p *AccrualProcessor) scheduleProcessingOrders(ctx context.Context) {
	if !p.canProcess() {
		p.retryMu.RLock()
		retryAt := p.nextAllowedRequestAt
		p.retryMu.RUnlock()

		p.logger.Info("accrual polling paused", "retry_after", time.Until(retryAt).String())
		return
	}

	orders, err := p.orderService.GetOrdersForProcessing(ctx)
	if err != nil {
		p.logger.Error("can't get orders for processing", "error", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	p.logger.Info("Sending orders to worker pool", "amount", len(orders))

	for _, order := range orders {
		select {
		case p.jobs <- order:
			// Заказ отправлен воркеру
		case <-ctx.Done():
			return
		}
	}
}

func (p *AccrualProcessor) worker(ctx context.Context, workerID int) {
	defer p.wg.Done()

	p.logger.Info("worker started", "worker_id", workerID)
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("worker stopped", "worker_id", workerID)
			return

		case order, ok := <-p.jobs:
			if !ok {
				p.logger.Info("jobs channel closed", "worker_id", workerID)
				return
			}
			p.processOrder(ctx, order)
		}
	}
}

func (p *AccrualProcessor) processOrder(ctx context.Context, order models.Order) {
	p.logger.Info("Processing order", "order", order.Number, "status", order.Status)

	resp, err := p.accrualClient.GetOrder(ctx, order.Number)
	if err != nil {
		// 429 Too Many Requests
		var retryErr *client.RetryAfterError
		if errors.As(err, &retryErr) {
			p.logger.Info("accrual rate limited", "order", order.Number, "retry_after", retryErr.RetryAfter.String())

			p.setRetryAfter(retryErr.RetryAfter)

			return
		}
		// network/internal errors
		p.logger.Info("can't get accrual status", "order", order.Number, "error", err)
		return
	}

	// 204 No Content
	if resp == nil {
		return
	}

	err = p.orderService.UpdateStatus(ctx, order.Number, resp.Status, &resp.Accrual)
	if err != nil {
		p.logger.Info("failed to update order", "order", order.Number, "error", err)
	}

	if resp.Status == "PROCESSED" && resp.Accrual > 0 {
		err := p.balanceService.CreateAccrual(ctx, order.Number, order.UserID, resp.Accrual)
		if err != nil {
			p.logger.Info("Failed to create accrual operation for order", "order", order.Number, "error", err)
		}
	}

	p.logger.Info("Order processed with status", "order", order.Number, "status", resp.Status, "accrual", resp.Accrual)
}
