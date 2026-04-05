package people

import (
	"context"
	"errors"
	"strings"

	"MonitorPeople/internal/domain"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrPersonNotFound      = errors.New("person not found")
	ErrPersonAlreadyPassed = errors.New("such person already passed")
	ErrPersonAlreadyExists = errors.New("person already exists")
)

type Repository interface {
	AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error)
	MarkPersonAsVisited(ctx context.Context, name, surname string) (domain.Person, error)
	DeletePerson(ctx context.Context, name, surname string) error
	ListPeople(ctx context.Context, filter domain.PeopleFilter) ([]domain.Person, error)
	GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error)
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

func (s *Service) ListPeople(ctx context.Context, filter domain.PeopleFilter) ([]domain.Person, error) {
	filter.StudyDirection = strings.TrimSpace(filter.StudyDirection)
	return s.repo.ListPeople(ctx, filter)
}

func (s *Service) GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error) {
	filter.StudyDirection = strings.TrimSpace(filter.StudyDirection)
	return s.repo.GetVisitedByProgramStats(ctx, filter)
}
