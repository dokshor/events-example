package providers

import (
	"context"
	"database/sql"
	"errors"
	"events/structures"

	"github.com/google/uuid"
)

type pgEventStore struct {
	db *sql.DB
}

func NewPGEventStore(db *sql.DB) *pgEventStore {
	return &pgEventStore{db: db}
}

func (s *pgEventStore) CreateEvent(ctx context.Context, e *structures.Event) (*structures.Event, error) {
	const q = `
        INSERT INTO events (id, title, description, start_time, end_time, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := s.db.ExecContext(ctx, q,
		e.ID,
		e.Title,
		e.Description,
		e.StartTime,
		e.EndTime,
		e.CreatedAt,
	)
	return e, err
}

func (s *pgEventStore) ListEvents(ctx context.Context) ([]structures.Event, error) {
	const q = `
        SELECT id, title, COALESCE(description, ''), start_time, end_time, created_at
        FROM events
        ORDER BY start_time ASC
    `
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]structures.Event, 0)
	for rows.Next() {
		var e structures.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *pgEventStore) GetEvent(ctx context.Context, id uuid.UUID) (*structures.Event, error) {
	const q = `
        SELECT id, title, COALESCE(description, ''), start_time, end_time, created_at
        FROM events
        WHERE id = $1
    `
	var e structures.Event
	err := s.db.QueryRowContext(ctx, q, id).
		Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}
