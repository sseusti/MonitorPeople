package people

import (
	"context"
	"errors"
	"strconv"
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

type Repository interface {
	AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error)
	MarkPersonAsVisited(ctx context.Context, name, surname string) (domain.Person, error)
	UndoCheckIn(ctx context.Context, name, surname string, within time.Duration) (domain.Person, error)
	DeletePerson(ctx context.Context, name, surname string) error
	DeleteStudents(ctx context.Context) (int64, error)
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

	return s.repo.AddPerson(ctx, name, surname, studyDirection, visitedEvent)
}

func (s *Service) ImportPeople(ctx context.Context, drafts []domain.PersonDraft) (domain.ImportResult, error) {
	result := domain.ImportResult{}
	for index, draft := range drafts {
		result.Processed++

		name := strings.TrimSpace(draft.Name)
		surname := strings.TrimSpace(draft.Surname)
		studyDirection := strings.TrimSpace(draft.StudyDirection)
		if name == "" || surname == "" || studyDirection == "" {
			result.SkippedInvalid++
			result.Errors = appendImportError(result.Errors, index, "пустые имя, фамилия или образовательная программа")
			continue
		}

		_, err := s.repo.AddPerson(ctx, name, surname, studyDirection, false)
		if err == nil {
			result.Imported++
			continue
		}
		if errors.Is(err, ErrPersonAlreadyExists) {
			result.SkippedDuplicates++
			continue
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return result, ctxErr
		}

		result.SkippedInvalid++
		result.Errors = appendImportError(result.Errors, index, "не удалось добавить запись: "+err.Error())
		continue
	}

	return result, nil
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

func (s *Service) DeleteStudents(ctx context.Context) (int64, error) {
	return s.repo.DeleteStudents(ctx)
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
	return s.repo.ListPeople(ctx, filter)
}

func (s *Service) GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error) {
	filter.StudyDirection = strings.TrimSpace(filter.StudyDirection)
	return s.repo.GetVisitedByProgramStats(ctx, filter)
}

func (s *Service) SuggestFieldValues(ctx context.Context, field, query string) ([]string, error) {
	field = strings.TrimSpace(field)
	query = strings.TrimSpace(query)
	if query == "" {
		return []string{}, nil
	}
	if field != "name" && field != "surname" && field != "studyDirection" {
		return nil, ErrInvalidSuggestField
	}

	return s.repo.SuggestFieldValues(ctx, field, query, 10)
}

func appendImportError(errorsList []string, rowIndex int, message string) []string {
	if len(errorsList) >= 20 {
		return errorsList
	}
	return append(errorsList, "строка "+strconv.Itoa(rowIndex+1)+": "+message)
}
