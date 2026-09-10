package rider

import (
	"context"
	"errors"
	"time"

	"github.com/PeaceMelodi/waka-dispatch/internal/ws"
)

type Service struct {
	repo *Repository
	hub  *ws.Hub
}

func NewService(repo *Repository, hub *ws.Hub) *Service {
	return &Service{
		repo: repo,
		hub:  hub,
	}
}

func (s *Service) Register(ctx context.Context, name string, phone string) (*Rider, error) {
	rider := &Rider{
		Name:   name,
		Phone:  phone,
		Status: "offline",
	}

	if err := s.repo.Create(ctx, rider); err != nil {
		return nil, err
	}

	return rider, nil
}

func (s *Service) UpdateLocation(ctx context.Context, id string, lat float64, lng float64) error {
	rider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rider == nil {
		return errors.New("rider not found")
	}

	if err := s.repo.UpdateLocation(ctx, id, lat, lng); err != nil {
		return err
	}

	s.hub.BroadcastRiderLocation(id, lat, lng)

	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	if status != "available" && status != "busy" && status != "offline" {
		return errors.New("status must be available, busy, or offline")
	}

	rider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rider == nil {
		return errors.New("rider not found")
	}

	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) List(ctx context.Context) ([]Rider, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) ListAvailable(ctx context.Context) ([]Rider, error) {
	return s.repo.ListAvailable(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Rider, error) {
	rider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rider == nil {
		return nil, errors.New("rider not found")
	}
	return rider, nil
}

func (s *Service) SetBusy(ctx context.Context, id string) error {
	return s.repo.SetBusy(ctx, id)
}

func (s *Service) SetAvailable(ctx context.Context, id string) error {
	return s.repo.SetAvailable(ctx, id)
}

func (s *Service) TouchRider(ctx context.Context, id string) error {
	rider, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rider == nil {
		return errors.New("rider not found")
	}

	if time.Since(rider.UpdatedAt) > 5*time.Minute {
		return s.repo.UpdateStatus(ctx, id, "offline")
	}
	return nil
}