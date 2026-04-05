package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"MonitorPeople/internal/repository/postgres"
	httpapi "MonitorPeople/internal/transport/http"
	"MonitorPeople/internal/usecase/people"

	"github.com/lib/pq"
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
		user := strings.TrimSpace(os.Getenv("USER"))
		if user == "" {
			user = "postgres"
		}
		dsn = fmt.Sprintf("postgres://%s@localhost:5432/monitor_people?sslmode=disable", user)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "3D000" {
			if createErr := ensureDatabaseExists(dsn); createErr != nil {
				return nil, createErr
			}

			db, err = sql.Open("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if err := db.Ping(); err != nil {
				db.Close()
				return nil, err
			}
			return db, nil
		}
		return nil, err
	}

	return db, nil
}

func ensureDatabaseExists(dsn string) error {
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		return errors.New("database name is empty in DATABASE_URL")
	}

	adminURL := *parsedURL
	adminURL.Path = "/postgres"
	adminDSN := adminURL.String()

	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return err
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		return err
	}

	_, err = adminDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "42P04" {
			return nil
		}
		return err
	}

	return nil
}
