package order

import (
	"context"
	"errors"
	"log"

	"github.com/PeaceMelodi/waka-dispatch/internal/delivery"
	"github.com/PeaceMelodi/waka-dispatch/internal/dispatch"
	"github.com/PeaceMelodi/waka-dispatch/internal/rider"
	"github.com/PeaceMelodi/waka-dispatch/internal/ws"
)

type Service struct {
	repo         *Repository
	matcher      *dispatch.Matcher
	riderSvc     *rider.Service
	deliveryRepo *delivery.Repository
	hub          *ws.Hub
}

func NewService(repo *Repository, matcher *dispatch.Matcher, riderSvc *rider.Service, deliveryRepo *delivery.Repository, hub *ws.Hub) *Service {
	return &Service{
		repo:         repo,
		matcher:      matcher,
		riderSvc:     riderSvc,
		deliveryRepo: deliveryRepo,
		hub:          hub,
	}
}

func (s *Service) Create(ctx context.Context, customerName string, pickupLat float64, pickupLng float64, dropoffLat float64, dropoffLng float64) (*Order, error) {
	if customerName == "" {
		return nil, errors.New("customer name is required")
	}
	if pickupLat == 0 || pickupLng == 0 {
		return nil, errors.New("pickup coordinates are required")
	}
	if dropoffLat == 0 || dropoffLng == 0 {
		return nil, errors.New("dropoff coordinates are required")
	}

	order := &Order{
		CustomerName: customerName,
		PickupLat:    pickupLat,
		PickupLng:    pickupLng,
		DropoffLat:   dropoffLat,
		DropoffLng:   dropoffLng,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	s.hub.BroadcastOrderStatus(order.ID, "pending")

	if err := s.autoMatch(ctx, order); err != nil {
		log.Printf("auto-match failed for order %s: %v", order.ID, err)
	}

	return order, nil
}

func (s *Service) autoMatch(ctx context.Context, order *Order) error {
	nearestRider, err := s.matcher.FindNearestRider(ctx, order.PickupLat, order.PickupLng)
	if err != nil {
		return err
	}


	del := &delivery.Delivery{
		OrderID: order.ID,
		RiderID: nearestRider.ID,
	}
	if err := s.deliveryRepo.Create(ctx, del); err != nil {
		return err
	}

	if err := s.riderSvc.SetBusy(ctx, nearestRider.ID); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, order.ID, "assigned"); err != nil {
		return err
	}

	order.Status = "assigned"
	s.hub.BroadcastOrderStatus(order.ID, "assigned")
	return nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (s *Service) List(ctx context.Context, status string) ([]Order, error) {
	return s.repo.List(ctx, status)
}

func (s *Service) Cancel(ctx context.Context, id string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}
	if order.Status != "pending" && order.Status != "assigned" {
		return errors.New("only pending or assigned orders can be cancelled")
	}
	if err := s.repo.UpdateStatus(ctx, id, "cancelled"); err != nil {
		return err
	}
	s.hub.BroadcastOrderStatus(id, "cancelled")
	return nil
}