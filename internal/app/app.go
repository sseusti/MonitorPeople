package app

import (
	"fmt"
	"net/http"

	"MonitorPeople/internal/app/config"
	dbinfra "MonitorPeople/internal/infrastructure/postgres"
	"MonitorPeople/internal/repository/postgres"
	httpapi "MonitorPeople/internal/transport/http"
	"MonitorPeople/internal/usecase/auth"
	"MonitorPeople/internal/usecase/people"
)

func Run(cfg config.Config) error {
	db, err := dbinfra.OpenDatabase(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer db.Close()

	if err := postgres.InitSchema(db); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	if err := postgres.EnsureAuthUser(db, cfg.AdminLogin, cfg.AdminPassword, "admin"); err != nil {
		return fmt.Errorf("ensure admin user: %w", err)
	}
	if err := postgres.EnsureAuthUser(db, cfg.EntranceLogin, cfg.EntrancePassword, "entrance"); err != nil {
		return fmt.Errorf("ensure entrance user: %w", err)
	}

	peopleRepo := postgres.NewPeopleRepository(db)
	peopleService := people.NewService(peopleRepo)
	peopleHandler := httpapi.NewPeopleHandler(peopleService)

	authRepo := postgres.NewAuthRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := httpapi.NewAuthHandler(authService)

	router := httpapi.NewRouter(authHandler, peopleHandler)
	address := ":" + cfg.HTTPPort
	if err := http.ListenAndServe(address, router); err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}
