package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	sqlBytes, err := ioutil.ReadFile("migrations/bootstrap.sql")
	if err != nil {
		log.Fatalf("read migration file: %v", err)
	}
	sqlText := string(sqlBytes)
	if sqlText == "" {
		log.Fatal("migration file is empty")
	}

	if _, err := db.ExecContext(ctx, sqlText); err != nil {
		log.Fatalf("exec migration: %v", err)
	}

	fmt.Println("migration applied successfully")
}
