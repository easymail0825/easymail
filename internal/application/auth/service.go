package auth

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidCredential = errors.New("invalid credential")
)

type Account struct {
	ID       int64
	Username string
	DomainID int64
	Active   bool
	Deleted  bool
}

type Repository interface {
	Authorize(ctx context.Context, username, password string) (*Account, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (*Account, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(password) < 6 {
		return nil, ErrInvalidCredential
	}

	acc, err := s.repo.Authorize(ctx, username, password)
	if err != nil {
		return nil, ErrInvalidCredential
	}
	if acc == nil || !acc.Active || acc.Deleted {
		return nil, ErrInvalidCredential
	}
	return acc, nil
}
