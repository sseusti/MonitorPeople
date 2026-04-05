package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"MonitorPeople/internal/repository/postgres"
	httpapi "MonitorPeople/internal/transport/http"
	"MonitorPeople/internal/usecase/people"

	_ "github.com/lib/pq"
)

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if err := postgres.InitSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	repo := postgres.NewPeopleRepository(db)
	service := people.NewService(repo)
	handler := httpapi.NewPeopleHandler(service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Println("server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func connectDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/monitor_people?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
