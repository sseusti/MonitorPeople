package people

import (
	"context"
	"errors"
	"testing"
	"time"

	"MonitorPeople/internal/domain"
)

type importRepoStub struct {
	failAt map[int]error
	calls  int
}

func (r *importRepoStub) AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error) {
	r.calls++
	if err := r.failAt[r.calls]; err != nil {
		return domain.Person{}, err
	}
	return domain.Person{Name: name, Surname: surname, StudyDirection: studyDirection}, nil
}

func (r *importRepoStub) MarkPersonAsVisited(ctx context.Context, name, surname string) (domain.Person, error) {
	return domain.Person{}, nil
}

func (r *importRepoStub) UndoCheckIn(ctx context.Context, name, surname string, within time.Duration) (domain.Person, error) {
	return domain.Person{}, nil
}

func (r *importRepoStub) DeletePerson(ctx context.Context, name, surname string) error {
	return nil
}

func (r *importRepoStub) DeleteStudents(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *importRepoStub) ListPeople(ctx context.Context, filter domain.PeopleFilter) ([]domain.Person, error) {
	return nil, nil
}

func (r *importRepoStub) GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error) {
	return nil, nil
}

func (r *importRepoStub) SuggestFieldValues(ctx context.Context, field, query string, limit int) ([]string, error) {
	return nil, nil
}

func TestImportPeopleSkipsUnexpectedRowErrors(t *testing.T) {
	repo := &importRepoStub{failAt: map[int]error{31: errors.New("bad row")}}
	service := NewService(repo)
	drafts := make([]domain.PersonDraft, 40)
	for i := range drafts {
		drafts[i] = domain.PersonDraft{
			Name:           "Name",
			Surname:        "Surname" + string(rune('A'+i)),
			StudyDirection: "Program",
		}
	}

	result, err := service.ImportPeople(context.Background(), drafts)
	if err != nil {
		t.Fatalf("ImportPeople returned error: %v", err)
	}
	if result.Imported != 39 || result.SkippedInvalid != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.calls != 40 {
		t.Fatalf("expected all rows to be attempted, got %d", repo.calls)
	}
}
