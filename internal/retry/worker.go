package retry

import (
	"context"
	"log"
	"time"

	"github.com/PeaceMelodi/waka-dispatch/internal/delivery"
	"github.com/PeaceMelodi/waka-dispatch/internal/dispatch"
	"github.com/PeaceMelodi/waka-dispatch/internal/order"
	"github.com/PeaceMelodi/waka-dispatch/internal/rider"
	"github.com/PeaceMelodi/waka-dispatch/internal/ws"
)

type Worker struct {
	orderRepo    *order.Repository
	riderRepo    *rider.Repository
	deliveryRepo *delivery.Repository
	matcher      *dispatch.Matcher
	hub          *ws.Hub
	interval     time.Duration
}

func NewWorker(
	orderRepo *order.Repository,
	riderRepo *rider.Repository,
	deliveryRepo *delivery.Repository,
	matcher *dispatch.Matcher,
	hub *ws.Hub,
) *Worker {
	return &Worker{
		orderRepo:    orderRepo,
		riderRepo:    riderRepo,
		deliveryRepo: deliveryRepo,
		matcher:      matcher,
		hub:          hub,
		interval:     10 * time.Second,
	}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("retry worker stopped")
				return
			case <-ticker.C:
				w.tick(ctx)
			}
		}
	}()
	log.Printf("retry worker started (every %s)", w.interval)
}

func (w *Worker) tick(ctx context.Context) {
	pendingOrders, err := w.orderRepo.List(ctx, "pending")
	if err != nil {
		log.Printf("retry: failed to list pending orders: %v", err)
		return
	}

	for _, ord := range pendingOrders {
		existing, err := w.deliveryRepo.GetByOrderID(ctx, ord.ID)
		if err == nil && existing != nil {
			continue
		}

		nearestRider, err := w.matcher.FindNearestRider(ctx, ord.PickupLat, ord.PickupLng)
		if err != nil {
			continue
		}

		del := &delivery.Delivery{
			OrderID: ord.ID,
			RiderID: nearestRider.ID,
		}
		if err := w.deliveryRepo.Create(ctx, del); err != nil {
			log.Printf("retry: failed to create delivery for order %s: %v", ord.ID, err)
			continue
		}

		if err := w.riderRepo.SetBusy(ctx, nearestRider.ID); err != nil {
			log.Printf("retry: failed to set rider %s busy: %v", nearestRider.ID, err)
			continue
		}

		if err := w.orderRepo.UpdateStatus(ctx, ord.ID, "assigned"); err != nil {
			log.Printf("retry: failed to update order %s: %v", ord.ID, err)
			continue
		}

		log.Printf("retry: order %s matched with rider %s", ord.ID, nearestRider.ID)
		w.hub.BroadcastOrderStatus(ord.ID, "assigned")
	}
}