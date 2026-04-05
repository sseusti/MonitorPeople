package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (r *PeopleRepository) ListPeople(ctx context.Context, filter domain.PeopleFilter) ([]domain.Person, error) {
	query := `
		SELECT order_number, name, surname, study_direction, visited_event, check_in_time
		FROM visitors
	`
	whereClause, args := buildVisitorsFilter(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += " ORDER BY order_number"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	peopleList := make([]domain.Person, 0)
	for rows.Next() {
		var person domain.Person
		if err := rows.Scan(
			&person.OrderNumber,
			&person.Name,
			&person.Surname,
			&person.StudyDirection,
			&person.VisitedEvent,
			&person.CheckInTime,
		); err != nil {
			return nil, err
		}
		peopleList = append(peopleList, person)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return peopleList, nil
}

func (r *PeopleRepository) GetVisitedByProgramStats(ctx context.Context, filter domain.PeopleFilter) ([]domain.ProgramStat, error) {
	statsFilter := filter
	visitedTrue := true
	statsFilter.Visited = &visitedTrue

	query := `
		SELECT study_direction, COUNT(*)
		FROM visitors
	`
	whereClause, args := buildVisitorsFilter(statsFilter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += " GROUP BY study_direction ORDER BY COUNT(*) DESC, study_direction"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]domain.ProgramStat, 0)
	for rows.Next() {
		var stat domain.ProgramStat
		if err := rows.Scan(&stat.StudyDirection, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func buildVisitorsFilter(filter domain.PeopleFilter) (string, []any) {
	parts := make([]string, 0, 2)
	args := make([]any, 0, 2)
	argPos := 1

	if filter.Visited != nil {
		parts = append(parts, fmt.Sprintf("visited_event = $%d", argPos))
		args = append(args, *filter.Visited)
		argPos++
	}

	if filter.StudyDirection != "" {
		parts = append(parts, fmt.Sprintf("study_direction = $%d", argPos))
		args = append(args, filter.StudyDirection)
	}

	return strings.Join(parts, " AND "), args
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
			role TEXT NOT NULL DEFAULT 'admin',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT admin_users_role_check CHECK (role IN ('admin', 'entrance'))
		)
	`
	_, err = db.Exec(adminUsersQuery)
	if err != nil {
		return err
	}

	alterRoleColumnQuery := `
		ALTER TABLE admin_users
		ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'admin'
	`
	_, err = db.Exec(alterRoleColumnQuery)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE admin_users SET role = 'admin' WHERE role IS NULL")
	if err != nil {
		return err
	}

	return nil
}

func EnsureAuthUser(db *sql.DB, login, password, role string) error {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	role = strings.TrimSpace(role)
	if login == "" || password == "" || role == "" {
		return nil
	}

	const query = `
		INSERT INTO admin_users (login, password, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (login) DO NOTHING
	`
	_, err := db.Exec(query, login, password, role)
	return err
}
