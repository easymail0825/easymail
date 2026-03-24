package identity

import (
	"context"
	"easymail/internal/application/auth"
	sessionkey "easymail/internal/application/session"
	authrepo "easymail/internal/infrastructure/repository"
)

type Service struct {
	auth *auth.Service
}

func NewService() *Service {
	return &Service{
		auth: auth.NewService(authrepo.NewAccountAuthRepository()),
	}
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (*auth.Account, error) {
	return s.auth.Authenticate(ctx, username, password)
}

func (s *Service) SessionKeys() map[string]string {
	return map[string]string{
		"admin":   sessionkey.KeyAdminAccount,
		"user_id": sessionkey.KeyUserID,
		"mailbox": sessionkey.KeyMailbox,
	}
}

