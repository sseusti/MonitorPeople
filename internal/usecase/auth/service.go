package auth

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
)

type Repository interface {
	Authenticate(ctx context.Context, login, password string) (User, bool, error)
}

type User struct {
	Login string
	Role  string
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(ctx context.Context, login, password string) (User, error) {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	if login == "" || password == "" {
		return User{}, ErrInvalidCredentials
	}

	user, ok, err := s.repo.Authenticate(ctx, login, password)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}
