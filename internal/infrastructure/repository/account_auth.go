package repository

import (
	"context"
	"easymail/internal/application/auth"
	"easymail/internal/model"
)

type AccountAuthRepository struct{}

func NewAccountAuthRepository() *AccountAuthRepository {
	return &AccountAuthRepository{}
}

func (r *AccountAuthRepository) Authorize(_ context.Context, username, password string) (*auth.Account, error) {
	acc, err := model.Authorize(username, password)
	if err != nil {
		return nil, err
	}
	return &auth.Account{
		ID:       acc.ID,
		Username: acc.Username,
		DomainID: acc.DomainID,
		Active:   acc.Active,
		Deleted:  acc.Deleted,
	}, nil
}
