package services

import (
	"context"
	"events/structures"

	"github.com/google/uuid"
)

type EventService interface {
	CreateEvent(ctx context.Context, e *structures.Event) (*structures.Event, error)
	ListEvents(ctx context.Context) ([]structures.Event, error)
	GetEvent(ctx context.Context, id uuid.UUID) (*structures.Event, error)
}

type eventService struct {
	store EventService
}

func NewEventService(store EventService) EventService {
	return &eventService{store: store}
}

func (s *eventService) CreateEvent(ctx context.Context, e *structures.Event) (*structures.Event, error) {
	return s.store.CreateEvent(ctx, e)
}

func (s *eventService) ListEvents(ctx context.Context) ([]structures.Event, error) {
	return s.store.ListEvents(ctx)
}

func (s *eventService) GetEvent(ctx context.Context, id uuid.UUID) (*structures.Event, error) {
	return s.store.GetEvent(ctx, id)
}
