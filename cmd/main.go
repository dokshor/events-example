package main

import (
	"context"
	"database/sql"
	"errors"
	"events/controller"
	"events/providers"
	"events/services"
	"events/utils"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type server struct {
	store services.EventService
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	repo := providers.NewPGEventStore(db)
	svc := services.NewEventService(repo)
	ec := controller.NewEventController(svc)

	mux := http.NewServeMux()
	ec.RegisterRoutes(mux)

	addr := ":8080"
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      utils.LoggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
