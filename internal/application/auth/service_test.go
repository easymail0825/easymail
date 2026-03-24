package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	account *Account
	err     error
}

func (f *fakeRepository) Authorize(_ context.Context, _, _ string) (*Account, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.account, nil
}

func TestServiceAuthenticate(t *testing.T) {
	svc := NewService(&fakeRepository{
		account: &Account{
			ID:       1,
			Username: "root@localhost",
			Active:   true,
		},
	})

	acc, err := svc.Authenticate(context.Background(), "root@localhost", "123456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if acc == nil || acc.Username != "root@localhost" {
		t.Fatalf("unexpected account: %#v", acc)
	}
}

func TestServiceAuthenticateInvalid(t *testing.T) {
	svc := NewService(&fakeRepository{err: errors.New("db error")})

	_, err := svc.Authenticate(context.Background(), "root@localhost", "123456")
	if !errors.Is(err, ErrInvalidCredential) {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
}
