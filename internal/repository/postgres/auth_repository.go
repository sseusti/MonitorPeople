package postgres

import (
	"context"
	"database/sql"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) ValidateAdmin(ctx context.Context, login, password string) (bool, error) {
	const query = `
		SELECT 1
		FROM admin_users
		WHERE login = $1 AND password = $2 AND is_active = true
		LIMIT 1
	`

	var marker int
	err := r.db.QueryRowContext(ctx, query, login, password).Scan(&marker)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
