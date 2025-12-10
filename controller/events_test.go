package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"events/structures"

	"github.com/google/uuid"
)

// --- mock service ---

type mockEventService struct {
	createCalled bool
	createReq    *structures.Event
	createResp   *structures.Event
	createErr    error

	listCalled bool
	listResp   []structures.Event
	listErr    error

	getCalled bool
	getID     uuid.UUID
	getResp   *structures.Event
	getErr    error
}

func (m *mockEventService) CreateEvent(ctx context.Context, e *structures.Event) (*structures.Event, error) {
	m.createCalled = true
	m.createReq = e
	return m.createResp, m.createErr
}

func (m *mockEventService) ListEvents(ctx context.Context) ([]structures.Event, error) {
	m.listCalled = true
	return m.listResp, m.listErr
}

func (m *mockEventService) GetEvent(ctx context.Context, id uuid.UUID) (*structures.Event, error) {
	m.getCalled = true
	m.getID = id
	return m.getResp, m.getErr
}

// --- tests ---

func TestHandleCreateEvent_Success(t *testing.T) {
	now := time.Now().UTC()
	respEvent := &structures.Event{
		ID:          uuid.New(),
		Title:       "Test",
		Description: "desc",
		StartTime:   now,
		EndTime:     now.Add(time.Hour),
		CreatedAt:   now,
	}

	mockSvc := &mockEventService{
		createResp: respEvent,
		createErr:  nil,
	}

	ctrl := NewEventController(mockSvc).(*eventController)

	body, _ := json.Marshal(structures.CreateEventRequest{
		Title:       "Test",
		Description: "desc",
		StartTime:   now,
		EndTime:     now.Add(time.Hour),
	})
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.handleCreateEvent(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}
	if !mockSvc.createCalled {
		t.Fatalf("expected CreateEvent to be called on service")
	}

	var got structures.Event
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != respEvent.ID {
		t.Fatalf("unexpected event ID: got %v, want %v", got.ID, respEvent.ID)
	}
}

func TestHandleCreateEvent_InvalidJSON(t *testing.T) {
	mockSvc := &mockEventService{}
	ctrl := NewEventController(mockSvc).(*eventController)

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString("{invalid-json"))
	w := httptest.NewRecorder()

	ctrl.handleCreateEvent(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	if mockSvc.createCalled {
		t.Fatalf("service should not be called on invalid JSON")
	}
}

func TestHandleCreateEvent_ServiceError(t *testing.T) {
	now := time.Now().UTC()
	mockSvc := &mockEventService{
		createErr: errors.New("validation error"),
	}
	ctrl := NewEventController(mockSvc).(*eventController)

	body, _ := json.Marshal(structures.CreateEventRequest{
		Title:       "Bad",
		Description: "desc",
		StartTime:   now,
		EndTime:     now.Add(time.Hour),
	})
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.handleCreateEvent(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	if !mockSvc.createCalled {
		t.Fatalf("expected service to be called")
	}
}

func TestHandleListEvents_Success(t *testing.T) {
	events := []structures.Event{
		{ID: uuid.New(), Title: "A"},
		{ID: uuid.New(), Title: "B"},
	}
	mockSvc := &mockEventService{
		listResp: events,
	}
	ctrl := NewEventController(mockSvc).(*eventController)

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	ctrl.handleListEvents(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if !mockSvc.listCalled {
		t.Fatalf("expected ListEvents to be called")
	}

	var got []structures.Event
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(got))
	}
}

func TestHandleGetEventByID_Success(t *testing.T) {
	id := uuid.New()
	mockSvc := &mockEventService{
		getResp: &structures.Event{ID: id, Title: "Found"},
	}
	ctrl := NewEventController(mockSvc).(*eventController)

	req := httptest.NewRequest(http.MethodGet, "/events/"+id.String(), nil)
	w := httptest.NewRecorder()

	ctrl.handleGetEventByID(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if !mockSvc.getCalled {
		t.Fatalf("expected GetEvent to be called")
	}
	if mockSvc.getID != id {
		t.Fatalf("service called with wrong ID: got %v, want %v", mockSvc.getID, id)
	}
}

func TestHandleGetEventByID_NotFound(t *testing.T) {
	id := uuid.New()
	mockSvc := &mockEventService{
		getResp: nil,
		getErr:  nil,
	}
	ctrl := NewEventController(mockSvc).(*eventController)

	req := httptest.NewRequest(http.MethodGet, "/events/"+id.String(), nil)
	w := httptest.NewRecorder()

	ctrl.handleGetEventByID(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestHandleGetEventByID_InvalidUUID(t *testing.T) {
	mockSvc := &mockEventService{}
	ctrl := NewEventController(mockSvc).(*eventController)

	req := httptest.NewRequest(http.MethodGet, "/events/not-a-uuid", nil)
	w := httptest.NewRecorder()

	ctrl.handleGetEventByID(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	if mockSvc.getCalled {
		t.Fatalf("service should not be called on invalid UUID")
	}
}
