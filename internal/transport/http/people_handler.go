package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"MonitorPeople/internal/domain"
	"MonitorPeople/internal/usecase/people"
)

type PeopleService interface {
	AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error)
	ImportPeople(ctx context.Context, drafts []domain.PersonDraft) (domain.ImportResult, error)
	CheckInPerson(ctx context.Context, name, surname string) (domain.Person, error)
	UndoCheckIn(ctx context.Context, name, surname string) (domain.Person, error)
	DeletePerson(ctx context.Context, name, surname string) error
	DeleteStudents(ctx context.Context) (int64, error)
	ListPeople(ctx context.Context, filter domain.PeopleFilter) ([]domain.Person, error)
	GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error)
	SuggestFieldValues(ctx context.Context, field, query string) ([]string, error)
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

func (h *PeopleHandler) CreatePersonHandler() http.HandlerFunc {
	return h.handleCreatePerson
}

func (h *PeopleHandler) CheckInHandler() http.HandlerFunc {
	return h.handleCheckIn
}

func (h *PeopleHandler) DeletePersonHandler() http.HandlerFunc {
	return h.handleDeletePerson
}

func (h *PeopleHandler) DeleteStudentsHandler() http.HandlerFunc {
	return h.handleDeleteStudents
}

func (h *PeopleHandler) UndoCheckInHandler() http.HandlerFunc {
	return h.handleUndoCheckIn
}

func (h *PeopleHandler) ListPeopleHandler() http.HandlerFunc {
	return h.handleListPeople
}

func (h *PeopleHandler) ProgramStatsHandler() http.HandlerFunc {
	return h.handleProgramStats
}

func (h *PeopleHandler) SuggestValuesHandler() http.HandlerFunc {
	return h.handleSuggestValues
}

func (h *PeopleHandler) ImportStudentsHandler() http.HandlerFunc {
	return h.handleImportStudents
}

func (h *PeopleHandler) ImportTeachersHandler() http.HandlerFunc {
	return h.handleImportTeachers
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
		case errors.Is(err, people.ErrInvalidProgram):
			http.Error(w, "invalid study direction", http.StatusBadRequest)
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

func (h *PeopleHandler) handleDeletePerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req checkInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePerson(r.Context(), req.Name, req.Surname); err != nil {
		switch {
		case errors.Is(err, people.ErrValidation):
			http.Error(w, "name and surname are required", http.StatusBadRequest)
		case errors.Is(err, people.ErrPersonNotFound):
			http.Error(w, "person not found", http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *PeopleHandler) handleDeleteStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deleted, err := h.service.DeleteStudents(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int64{"deleted": deleted})
}

func (h *PeopleHandler) handleUndoCheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req checkInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	person, err := h.service.UndoCheckIn(r.Context(), req.Name, req.Surname)
	if err != nil {
		switch {
		case errors.Is(err, people.ErrValidation):
			http.Error(w, "name and surname are required", http.StatusBadRequest)
		case errors.Is(err, people.ErrPersonNotFound):
			http.Error(w, "person not found", http.StatusNotFound)
		case errors.Is(err, people.ErrPersonNotVisited):
			http.Error(w, "person is not checked in", http.StatusConflict)
		case errors.Is(err, people.ErrUndoWindowExpired):
			http.Error(w, "undo window expired", http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(person)
}

func (h *PeopleHandler) handleListPeople(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filter, err := filterFromQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	peopleList, err := h.service.ListPeople(r.Context(), filter)
	if err != nil {
		if errors.Is(err, people.ErrInvalidProgram) {
			http.Error(w, "invalid study direction", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peopleList)
}

func (h *PeopleHandler) handleProgramStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filter, err := filterFromQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stats, err := h.service.GetVisitedByProgramStats(r.Context(), filter)
	if err != nil {
		if errors.Is(err, people.ErrInvalidProgram) {
			http.Error(w, "invalid study direction", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (h *PeopleHandler) handleSuggestValues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	field := strings.TrimSpace(r.URL.Query().Get("field"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	values, err := h.service.SuggestFieldValues(r.Context(), field, query)
	if err != nil {
		if errors.Is(err, people.ErrInvalidSuggestField) {
			http.Error(w, "invalid suggest field", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(values)
}

func (h *PeopleHandler) handleImportStudents(w http.ResponseWriter, r *http.Request) {
	h.handleImport(w, r, parseStudentsImport)
}

func (h *PeopleHandler) handleImportTeachers(w http.ResponseWriter, r *http.Request) {
	h.handleImport(w, r, parseTeachersImport)
}

func (h *PeopleHandler) handleImport(
	w http.ResponseWriter,
	r *http.Request,
	parse func(fileReader io.Reader, filename string) ([]domain.PersonDraft, error),
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(12 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	drafts, err := parse(file, header.Filename)
	if err != nil {
		switch {
		case errors.Is(err, errUnsupportedImportFile):
			http.Error(w, "unsupported import file type", http.StatusBadRequest)
		case errors.Is(err, errImportColumnsNotFound):
			http.Error(w, "required import columns not found", http.StatusBadRequest)
		default:
			http.Error(w, "invalid import file", http.StatusBadRequest)
		}
		return
	}
	if len(drafts) == 0 {
		http.Error(w, "no valid rows found", http.StatusBadRequest)
		return
	}

	result, err := h.service.ImportPeople(r.Context(), drafts)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func filterFromQuery(r *http.Request) (domain.PeopleFilter, error) {
	query := r.URL.Query()
	filter := domain.PeopleFilter{
		StudyDirection: strings.TrimSpace(query.Get("studyDirection")),
	}

	visitedRaw := strings.TrimSpace(query.Get("visited"))
	if visitedRaw == "" || visitedRaw == "all" {
		return filter, nil
	}
	if visitedRaw == "true" {
		visitedTrue := true
		filter.Visited = &visitedTrue
		return filter, nil
	}
	if visitedRaw == "false" {
		visitedFalse := false
		filter.Visited = &visitedFalse
		return filter, nil
	}

	return domain.PeopleFilter{}, errors.New("visited must be true, false or all")
}
