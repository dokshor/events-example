package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"events/services"
	"events/structures"

	"github.com/google/uuid"
)

type EventController interface {
	RegisterRoutes(mux *http.ServeMux)
}

type eventController struct {
	svc services.EventService
}

func NewEventController(svc services.EventService) EventController {
	return &eventController{svc: svc}
}

func (c *eventController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /events", c.handleCreateEvent)
	mux.HandleFunc("GET /events", c.handleListEvents)
	mux.HandleFunc("GET /events/", c.handleGetEventByID)
}

func (c *eventController) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req structures.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if len(req.Title) > 100 {
		http.Error(w, "title must be at most 100 characters", http.StatusBadRequest)
		return
	}
	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		http.Error(w, "start_time and end_time are required", http.StatusBadRequest)
		return
	}
	if !req.StartTime.Before(req.EndTime) {
		http.Error(w, "start_time must be before end_time", http.StatusBadRequest)
		return
	}

	e, err := c.svc.CreateEvent(ctx, &structures.Event{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		CreatedAt:   time.Now(),
	})

	if err != nil {
		log.Printf("Create error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, e)
}

func (c *eventController) handleListEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	events, err := c.svc.ListEvents(ctx)
	if err != nil {
		log.Printf("List error: %v", err)
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (c *eventController) handleGetEventByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Path[len("/events/"):]
	if idStr == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid UUID", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	e, err := c.svc.GetEvent(ctx, id)
	if err != nil {
		log.Printf("Get error: %v", err)
		http.Error(w, "failed to get event", http.StatusInternalServerError)
		return
	}
	if e == nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
