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
	"MonitorPeople/internal/usecase/auth"
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

	adminLogin := envOrDefault("ADMIN_LOGIN", "admin")
	adminPassword := envOrDefault("ADMIN_PASSWORD", "admin123")
	entranceLogin := envOrDefault("ENTRANCE_LOGIN", "entrance")
	entrancePassword := envOrDefault("ENTRANCE_PASSWORD", "entrance123")
	if err := postgres.EnsureAuthUser(db, adminLogin, adminPassword, "admin"); err != nil {
		log.Fatalf("ensure admin user: %v", err)
	}
	if err := postgres.EnsureAuthUser(db, entranceLogin, entrancePassword, "entrance"); err != nil {
		log.Fatalf("ensure entrance user: %v", err)
	}

	repo := postgres.NewPeopleRepository(db)
	service := people.NewService(repo)
	handler := httpapi.NewPeopleHandler(service)
	authRepo := postgres.NewAuthRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := httpapi.NewAuthHandler(authService)

	mux := http.NewServeMux()
	authHandler.RegisterPublicRoutes(mux)
	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("/styles.css", fileServer)
	mux.Handle("/login.js", fileServer)
	mux.Handle("/admin.js", fileServer)
	mux.Handle("/entrance.js", fileServer)

	mux.Handle("/", authHandler.RequireAuth(http.HandlerFunc(authHandler.HandleHome)))
	mux.Handle("/admin", authHandler.RequireAuth(http.HandlerFunc(authHandler.HandleAdminPage)))
	mux.Handle("/entrance", authHandler.RequireAuth(http.HandlerFunc(authHandler.HandleEntrancePage)))

	mux.Handle("/people/check-in", authHandler.RequireAuth(http.HandlerFunc(handler.CheckInHandler())))
	mux.Handle("/people", authHandler.RequireAdmin(http.HandlerFunc(handler.CreatePersonHandler())))
	mux.Handle("/people/delete", authHandler.RequireAdmin(http.HandlerFunc(handler.DeletePersonHandler())))
	mux.Handle("/people/list", authHandler.RequireAdmin(http.HandlerFunc(handler.ListPeopleHandler())))
	mux.Handle("/people/stats/programs", authHandler.RequireAdmin(http.HandlerFunc(handler.ProgramStatsHandler())))

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

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
