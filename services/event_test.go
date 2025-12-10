package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"events/structures"

	"github.com/google/uuid"
)

// mockEventService implements EventService and lets us control responses / capture calls.
type mockEventService struct {
	createCalled bool
	createArg    *structures.Event
	createResp   *structures.Event
	createErr    error

	listCalled bool
	listResp   []structures.Event
	listErr    error

	getCalled bool
	getArgID  uuid.UUID
	getResp   *structures.Event
	getErr    error
}

func (m *mockEventService) CreateEvent(ctx context.Context, e *structures.Event) (*structures.Event, error) {
	m.createCalled = true
	m.createArg = e
	return m.createResp, m.createErr
}

func (m *mockEventService) ListEvents(ctx context.Context) ([]structures.Event, error) {
	m.listCalled = true
	return m.listResp, m.listErr
}

func (m *mockEventService) GetEvent(ctx context.Context, id uuid.UUID) (*structures.Event, error) {
	m.getCalled = true
	m.getArgID = id
	return m.getResp, m.getErr
}

func TestEventService_CreateEvent_DelegatesToInner(t *testing.T) {
	ctx := context.Background()

	input := &structures.Event{
		ID:          uuid.New(),
		Title:       "Test",
		Description: "desc",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		CreatedAt:   time.Now(),
	}
	expected := &structures.Event{
		ID:          input.ID,
		Title:       "Created",
		Description: "created",
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		CreatedAt:   input.CreatedAt,
	}

	mockInner := &mockEventService{
		createResp: expected,
		createErr:  nil,
	}

	svc := NewEventService(mockInner)

	got, err := svc.CreateEvent(ctx, input)
	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}
	if !mockInner.createCalled {
		t.Fatalf("expected inner CreateEvent to be called")
	}
	if mockInner.createArg != input {
		t.Fatalf("inner CreateEvent called with wrong arg: got %+v, want %+v", mockInner.createArg, input)
	}
	if got != expected {
		t.Fatalf("CreateEvent returned wrong value: got %+v, want %+v", got, expected)
	}
}

func TestEventService_ListEvents_DelegatesToInner(t *testing.T) {
	ctx := context.Background()

	expected := []structures.Event{
		{ID: uuid.New(), Title: "A"},
		{ID: uuid.New(), Title: "B"},
	}

	mockInner := &mockEventService{
		listResp: expected,
		listErr:  nil,
	}

	svc := NewEventService(mockInner)

	got, err := svc.ListEvents(ctx)
	if err != nil {
		t.Fatalf("ListEvents returned error: %v", err)
	}
	if !mockInner.listCalled {
		t.Fatalf("expected inner ListEvents to be called")
	}
	if len(got) != len(expected) {
		t.Fatalf("ListEvents returned wrong length: got %d, want %d", len(got), len(expected))
	}
}

func TestEventService_GetEvent_DelegatesToInner(t *testing.T) {
	ctx := context.Background()

	id := uuid.New()
	expected := &structures.Event{
		ID:    id,
		Title: "Found",
	}

	mockInner := &mockEventService{
		getResp: expected,
		getErr:  nil,
	}

	svc := NewEventService(mockInner)

	got, err := svc.GetEvent(ctx, id)
	if err != nil {
		t.Fatalf("GetEvent returned error: %v", err)
	}
	if !mockInner.getCalled {
		t.Fatalf("expected inner GetEvent to be called")
	}
	if mockInner.getArgID != id {
		t.Fatalf("inner GetEvent called with wrong id: got %v, want %v", mockInner.getArgID, id)
	}
	if got != expected {
		t.Fatalf("GetEvent returned wrong value: got %+v, want %+v", got, expected)
	}
}

func TestEventService_PropagatesErrors(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("inner error")

	mockInner := &mockEventService{
		createErr: wantErr,
		listErr:   wantErr,
		getErr:    wantErr,
	}

	svc := NewEventService(mockInner)

	if _, err := svc.CreateEvent(ctx, &structures.Event{}); err != wantErr {
		t.Fatalf("CreateEvent did not propagate error: got %v, want %v", err, wantErr)
	}
	if _, err := svc.ListEvents(ctx); err != wantErr {
		t.Fatalf("ListEvents did not propagate error: got %v, want %v", err, wantErr)
	}
	if _, err := svc.GetEvent(ctx, uuid.New()); err != wantErr {
		t.Fatalf("GetEvent did not propagate error: got %v, want %v", err, wantErr)
	}
}
