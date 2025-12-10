package providers

import (
	"context"
	"regexp"
	"testing"
	"time"

	"events/structures"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCreateEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := &pgEventStore{db: db}

	e := &structures.Event{
		ID:          uuid.New(),
		Title:       "Test Event",
		Description: "desc",
		StartTime:   time.Now().Add(time.Hour),
		EndTime:     time.Now().Add(2 * time.Hour),
		CreatedAt:   time.Now(),
	}

	query := regexp.QuoteMeta(`
        INSERT INTO events (id, title, description, start_time, end_time, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `)

	mock.ExpectExec(query).
		WithArgs(e.ID, e.Title, e.Description, e.StartTime, e.EndTime, e.CreatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := store.CreateEvent(context.Background(), e)
	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}
	if got == nil || got.ID != e.ID {
		t.Fatalf("CreateEvent returned wrong event: got %+v, want %+v", got, e)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := &pgEventStore{db: db}

	now := time.Now().UTC()
	eID := uuid.New()

	query := regexp.QuoteMeta(`
        SELECT id, title, COALESCE(description, ''), start_time, end_time, created_at
        FROM events
        ORDER BY start_time ASC
    `)

	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "start_time", "end_time", "created_at",
	}).AddRow(eID, "Test Event", "desc", now, now.Add(time.Hour), now)

	mock.ExpectQuery(query).WillReturnRows(rows)

	result, err := store.ListEvents(context.Background())
	if err != nil {
		t.Fatalf("ListEvents returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result))
	}
	if result[0].ID != eID {
		t.Fatalf("unexpected event ID: got %v, want %v", result[0].ID, eID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetEvent_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := &pgEventStore{db: db}

	now := time.Now().UTC()
	eID := uuid.New()

	query := regexp.QuoteMeta(`
        SELECT id, title, COALESCE(description, ''), start_time, end_time, created_at
        FROM events
        WHERE id = $1
    `)

	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "start_time", "end_time", "created_at",
	}).AddRow(eID, "Test Event", "desc", now, now.Add(time.Hour), now)

	mock.ExpectQuery(query).
		WithArgs(eID).
		WillReturnRows(rows)

	e, err := store.GetEvent(context.Background(), eID)
	if err != nil {
		t.Fatalf("GetEvent returned error: %v", err)
	}
	if e == nil || e.ID != eID {
		t.Fatalf("unexpected event: %+v", e)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := &pgEventStore{db: db}

	query := regexp.QuoteMeta(`
        SELECT id, title, COALESCE(description, ''), start_time, end_time, created_at
        FROM events
        WHERE id = $1
    `)

	mock.ExpectQuery(query).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "start_time", "end_time", "created_at",
		})) // no rows

	e, err := store.GetEvent(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetEvent returned error: %v", err)
	}
	if e != nil {
		t.Fatalf("expected nil event when not found, got %+v", e)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
