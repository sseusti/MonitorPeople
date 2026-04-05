package postgres

import (
	"context"
	"database/sql"

	"MonitorPeople/internal/usecase/auth"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Authenticate(ctx context.Context, login, password string) (auth.User, bool, error) {
	const query = `
		SELECT login, role
		FROM admin_users
		WHERE login = $1 AND password = $2 AND is_active = true
		LIMIT 1
	`

	var user auth.User
	err := r.db.QueryRowContext(ctx, query, login, password).Scan(&user.Login, &user.Role)
	if err == sql.ErrNoRows {
		return auth.User{}, false, nil
	}
	if err != nil {
		return auth.User{}, false, err
	}

	return user, true, nil
}
