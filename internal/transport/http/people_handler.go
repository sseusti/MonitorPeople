package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"MonitorPeople/internal/domain"
	"MonitorPeople/internal/usecase/people"
)

type PeopleService interface {
	AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error)
	CheckInPerson(ctx context.Context, name, surname string) (domain.Person, error)
}

type PeopleHandler struct {
	service PeopleService
}

type createPersonRequest struct {
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	StudyDirection string `json:"studyDirection"`
	VisitedEvent   bool   `json:"visitedEvent"`
}

type checkInRequest struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

func NewPeopleHandler(service PeopleService) *PeopleHandler {
	return &PeopleHandler{service: service}
}

func (h *PeopleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/people", h.handleCreatePerson)
	mux.HandleFunc("/people/check-in", h.handleCheckIn)
}

func (h *PeopleHandler) handleCreatePerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createPersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	person, err := h.service.AddPerson(r.Context(), req.Name, req.Surname, req.StudyDirection, req.VisitedEvent)
	if err != nil {
		switch {
		case errors.Is(err, people.ErrValidation):
			http.Error(w, "name, surname and studyDirection are required", http.StatusBadRequest)
		case errors.Is(err, people.ErrPersonAlreadyExists):
			http.Error(w, "person already exists", http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(person); err != nil {
		log.Printf("encode response error: %v", err)
	}
}

func (h *PeopleHandler) handleCheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req checkInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	person, err := h.service.CheckInPerson(r.Context(), req.Name, req.Surname)
	if err != nil {
		switch {
		case errors.Is(err, people.ErrValidation):
			http.Error(w, "name and surname are required", http.StatusBadRequest)
		case errors.Is(err, people.ErrPersonAlreadyPassed):
			http.Error(w, "such person already passed", http.StatusConflict)
		case errors.Is(err, people.ErrPersonNotFound):
			http.Error(w, "person not found", http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(person); err != nil {
		log.Printf("encode response error: %v", err)
	}
}
