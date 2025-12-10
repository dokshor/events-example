package clients

import "database/sql"

type pgEventStore struct {
	db *sql.DB
}

func NewPGEventStore(db *sql.DB) *pgEventStore {
	return &pgEventStore{db: db}
}
