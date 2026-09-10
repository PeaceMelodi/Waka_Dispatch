package delivery

import (
	"context"
	"errors"

	"github.com/PeaceMelodi/waka-dispatch/internal/rider"
	"github.com/PeaceMelodi/waka-dispatch/internal/ws"
)

type OrderStatusUpdater interface {
	UpdateStatus(ctx context.Context, id string, status string) error
}

type Service struct {
	repo      *Repository
	orderRepo OrderStatusUpdater
	riderRepo *rider.Repository
	hub       *ws.Hub
}

func NewService(repo *Repository, orderRepo OrderStatusUpdater, riderRepo *rider.Repository, hub *ws.Hub) *Service {
	return &Service{
		repo:      repo,
		orderRepo: orderRepo,
		riderRepo: riderRepo,
		hub:       hub,
	}
}

func (s *Service) Create(ctx context.Context, orderID string, riderID string) (*Delivery, error) {
	delivery := &Delivery{
		OrderID: orderID,
		RiderID: riderID,
	}

	if err := s.repo.Create(ctx, delivery); err != nil {
		return nil, err
	}

	return delivery, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Delivery, error) {
	delivery, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if delivery == nil {
		return nil, errors.New("delivery not found")
	}
	return delivery, nil
}

func (s *Service) GetByOrderID(ctx context.Context, orderID string) (*Delivery, error) {
	delivery, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if delivery == nil {
		return nil, errors.New("delivery not found for this order")
	}
	return delivery, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, newStatus string) error {
	delivery, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if delivery == nil {
		return errors.New("delivery not found")
	}

	if err := validateTransition(delivery, newStatus); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, newStatus); err != nil {
		return err
	}

	orderStatus := mapDeliveryStatusToOrderStatus(newStatus)
	if orderStatus != "" {
		if err := s.orderRepo.UpdateStatus(ctx, delivery.OrderID, orderStatus); err != nil {
			return err
		}
		// Broadcast order status
		s.hub.BroadcastOrderStatus(delivery.OrderID, orderStatus)
	}

	if newStatus == "delivered" || newStatus == "cancelled" {
		if err := s.riderRepo.SetAvailable(ctx, delivery.RiderID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ListActive(ctx context.Context) ([]Delivery, error) {
	return s.repo.ListActive(ctx)
}

func validateTransition(delivery *Delivery, newStatus string) error {
	if delivery.DeliveredAt != nil {
		return errors.New("delivery already delivered")
	}
	if delivery.CancelledAt != nil {
		return errors.New("delivery already cancelled")
	}

	switch newStatus {
	case "picked_up":
		if delivery.PickedUpAt != nil {
			return errors.New("already picked up")
		}
	case "in_transit":
		if delivery.PickedUpAt == nil {
			return errors.New("must pick up before going in transit")
		}
		if delivery.InTransitAt != nil {
			return errors.New("already in transit")
		}
	case "delivered":
		if delivery.InTransitAt == nil {
			return errors.New("must be in transit before delivering")
		}
	case "cancelled":
		if delivery.CancelledAt != nil {
			return errors.New("already cancelled")
		}
	default:
		return errors.New("invalid delivery status")
	}

	return nil
}

func mapDeliveryStatusToOrderStatus(deliveryStatus string) string {
	switch deliveryStatus {
	case "picked_up":
		return "picked_up"
	case "in_transit":
		return "in_transit"
	case "delivered":
		return "delivered"
	case "cancelled":
		return "cancelled"
	default:
		return ""
	}
}