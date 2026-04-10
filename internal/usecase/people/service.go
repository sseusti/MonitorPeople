package people

import (
	"context"
	"errors"
	"strings"
	"time"

	"MonitorPeople/internal/domain"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrPersonNotFound      = errors.New("person not found")
	ErrPersonAlreadyPassed = errors.New("such person already passed")
	ErrPersonAlreadyExists = errors.New("person already exists")
	ErrPersonNotVisited    = errors.New("person is not checked in")
	ErrInvalidProgram      = errors.New("invalid study direction")
	ErrInvalidSuggestField = errors.New("invalid suggest field")
	ErrUndoWindowExpired   = errors.New("undo window expired")
)

var allowedPrograms = map[string]struct{}{
	"КНТ":  {},
	"МББЭ": {},
	"ЦМ":   {},
	"Ю":    {},
}

type Repository interface {
	AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error)
	MarkPersonAsVisited(ctx context.Context, name, surname string) (domain.Person, error)
	UndoCheckIn(ctx context.Context, name, surname string, within time.Duration) (domain.Person, error)
	DeletePerson(ctx context.Context, name, surname string) error
	ListPeople(ctx context.Context, filter domain.PeopleFilter) ([]domain.Person, error)
	GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error)
	SuggestFieldValues(ctx context.Context, field, query string, limit int) ([]string, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error) {
	name = strings.TrimSpace(name)
	surname = strings.TrimSpace(surname)
	studyDirection = strings.TrimSpace(studyDirection)
	if name == "" || surname == "" || studyDirection == "" {
		return domain.Person{}, ErrValidation
	}
	if !isAllowedProgram(studyDirection) {
		return domain.Person{}, ErrInvalidProgram
	}

	return s.repo.AddPerson(ctx, name, surname, studyDirection, visitedEvent)
}

func (s *Service) CheckInPerson(ctx context.Context, name, surname string) (domain.Person, error) {
	name = strings.TrimSpace(name)
	surname = strings.TrimSpace(surname)
	if name == "" || surname == "" {
		return domain.Person{}, ErrValidation
	}

	return s.repo.MarkPersonAsVisited(ctx, name, surname)
}

func (s *Service) DeletePerson(ctx context.Context, name, surname string) error {
	name = strings.TrimSpace(name)
	surname = strings.TrimSpace(surname)
	if name == "" || surname == "" {
		return ErrValidation
	}

	return s.repo.DeletePerson(ctx, name, surname)
}

func (s *Service) UndoCheckIn(ctx context.Context, name, surname string) (domain.Person, error) {
	name = strings.TrimSpace(name)
	surname = strings.TrimSpace(surname)
	if name == "" || surname == "" {
		return domain.Person{}, ErrValidation
	}

	return s.repo.UndoCheckIn(ctx, name, surname, time.Minute)
}

func (s *Service) ListPeople(ctx context.Context, filter domain.PeopleFilter) ([]domain.Person, error) {
	filter.StudyDirection = strings.TrimSpace(filter.StudyDirection)
	if filter.StudyDirection != "" && !isAllowedProgram(filter.StudyDirection) {
		return nil, ErrInvalidProgram
	}
	return s.repo.ListPeople(ctx, filter)
}

func (s *Service) GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error) {
	filter.StudyDirection = strings.TrimSpace(filter.StudyDirection)
	if filter.StudyDirection != "" && !isAllowedProgram(filter.StudyDirection) {
		return nil, ErrInvalidProgram
	}
	return s.repo.GetVisitedByProgramStats(ctx, filter)
}

func isAllowedProgram(program string) bool {
	_, ok := allowedPrograms[program]
	return ok
}

func (s *Service) SuggestFieldValues(ctx context.Context, field, query string) ([]string, error) {
	field = strings.TrimSpace(field)
	query = strings.TrimSpace(query)
	if query == "" {
		return []string{}, nil
	}
	if field != "name" && field != "surname" {
		return nil, ErrInvalidSuggestField
	}

	return s.repo.SuggestFieldValues(ctx, field, query, 10)
}
