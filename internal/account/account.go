package account

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

/*
Account
@Desc
account model
*/
type Account struct {
	ID        int64  `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	DomainID  int64
	Domain    *Domain
	Active    bool      `json:"active"`
	Deleted   bool      `json:"deleted"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
	CreatedAt time.Time `json:"created_at"`
}

// GeneratePassword create hashed password
func GeneratePassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verify hashed password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
