package main

import (
	"log"

	"MonitorPeople/internal/app"
	"MonitorPeople/internal/app/config"
)

func main() {
	cfg := config.Load()
	log.Printf("server starting on :%s", cfg.HTTPPort)
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
