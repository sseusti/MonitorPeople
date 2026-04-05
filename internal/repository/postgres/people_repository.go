package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"MonitorPeople/internal/domain"
	"MonitorPeople/internal/usecase/people"

	"github.com/lib/pq"
)

type PeopleRepository struct {
	db *sql.DB
}

func NewPeopleRepository(db *sql.DB) *PeopleRepository {
	return &PeopleRepository{db: db}
}

func (r *PeopleRepository) AddPerson(ctx context.Context, name, surname, studyDirection string, visitedEvent bool) (domain.Person, error) {
	query := `
		INSERT INTO visitors (name, surname, study_direction, visited_event, check_in_time)
		VALUES ($1, $2, $3, $4, CASE WHEN $4 THEN NOW() ELSE NULL END)
		RETURNING order_number, name, surname, study_direction, visited_event, check_in_time
	`

	var person domain.Person
	err := r.db.QueryRowContext(ctx, query, name, surname, studyDirection, visitedEvent).Scan(
		&person.OrderNumber,
		&person.Name,
		&person.Surname,
		&person.StudyDirection,
		&person.VisitedEvent,
		&person.CheckInTime,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.Person{}, people.ErrPersonAlreadyExists
		}
		return domain.Person{}, err
	}

	return person, nil
}

func (r *PeopleRepository) MarkPersonAsVisited(ctx context.Context, name, surname string) (domain.Person, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Person{}, err
	}
	defer tx.Rollback()

	var person domain.Person
	selectQuery := `
		SELECT order_number, name, surname, study_direction, visited_event, check_in_time
		FROM visitors
		WHERE name = $1 AND surname = $2
		ORDER BY order_number
		LIMIT 1
		FOR UPDATE
	`

	err = tx.QueryRowContext(ctx, selectQuery, name, surname).Scan(
		&person.OrderNumber,
		&person.Name,
		&person.Surname,
		&person.StudyDirection,
		&person.VisitedEvent,
		&person.CheckInTime,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Person{}, people.ErrPersonNotFound
	}
	if err != nil {
		return domain.Person{}, err
	}

	if person.VisitedEvent {
		return domain.Person{}, people.ErrPersonAlreadyPassed
	}

	updateQuery := `
		UPDATE visitors
		SET visited_event = true, check_in_time = NOW()
		WHERE order_number = $1
		RETURNING order_number, name, surname, study_direction, visited_event, check_in_time
	`
	err = tx.QueryRowContext(ctx, updateQuery, person.OrderNumber).Scan(
		&person.OrderNumber,
		&person.Name,
		&person.Surname,
		&person.StudyDirection,
		&person.VisitedEvent,
		&person.CheckInTime,
	)
	if err != nil {
		return domain.Person{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Person{}, err
	}

	return person, nil
}

func (r *PeopleRepository) DeletePerson(ctx context.Context, name, surname string) error {
	const query = `
		DELETE FROM visitors
		WHERE name = $1 AND surname = $2
	`

	result, err := r.db.ExecContext(ctx, query, name, surname)
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		return people.ErrPersonNotFound
	}

	return nil
}

func InitSchema(db *sql.DB) error {
	visitorsQuery := `
		CREATE TABLE IF NOT EXISTS visitors (
			order_number BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			surname TEXT NOT NULL,
			study_direction TEXT NOT NULL,
			visited_event BOOLEAN NOT NULL DEFAULT FALSE,
			check_in_time TIMESTAMPTZ NULL,
			CONSTRAINT visitors_name_surname_unique UNIQUE (name, surname)
		)
	`
	_, err := db.Exec(visitorsQuery)
	if err != nil {
		return err
	}

	adminUsersQuery := `
		CREATE TABLE IF NOT EXISTS admin_users (
			id BIGSERIAL PRIMARY KEY,
			login TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	_, err = db.Exec(adminUsersQuery)
	return err
}

func EnsureAdminUser(db *sql.DB, login, password string) error {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	if login == "" || password == "" {
		return nil
	}

	const query = `
		INSERT INTO admin_users (login, password)
		VALUES ($1, $2)
		ON CONFLICT (login) DO NOTHING
	`
	_, err := db.Exec(query, login, password)
	return err
}
