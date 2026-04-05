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
	ValidateAdmin(ctx context.Context, login, password string) (bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(ctx context.Context, login, password string) error {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	if login == "" || password == "" {
		return ErrInvalidCredentials
	}

	ok, err := s.repo.ValidateAdmin(ctx, login, password)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidCredentials
	}
	return nil
}
