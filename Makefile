APP_PORT ?= 8080
APP_URL ?= http://localhost:$(APP_PORT)
DATABASE_URL ?= postgres://stepan@localhost:5432/monitor_people?sslmode=disable

.PHONY: run up

run:
	DATABASE_URL="$(DATABASE_URL)" go run main.go

up:
	@echo "Starting backend on $(APP_URL)"
	@DATABASE_URL="$(DATABASE_URL)" go run main.go & \
	BACKEND_PID=$$!; \
	trap 'if kill -0 $$BACKEND_PID 2>/dev/null; then kill $$BACKEND_PID; fi' INT TERM EXIT; \
	sleep 1; \
	open "$(APP_URL)"; \
	wait $$BACKEND_PID
