test:
	go test ./... -v
run:
	go run main.go
migrate:
	go run migrations/migrate.go