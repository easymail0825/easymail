package account

import (
	"easymail/internal/database"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

// crypt加密用户密码
func TestCrypt(t *testing.T) {
	password := "admin"
	hashedPassword, err := GeneratePassword(password)
	t.Log(hashedPassword, err)
}

func TestCompare(t *testing.T) {
	password := "admin"
	hashedPassword := "$2a$10$yAanFMv4U7tX21Q9yDhpqOhLWVGhlpYl4ZZMKVfjc0x6N/g/wKrDS"
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	t.Log(err)
}

func TestMigrate(t *testing.T) {
	db := database.DB
	var err error
	//if err = db.AutoMigrate(&Domain{}); err != nil {
	//	t.Fail()
	//}
	err = db.AutoMigrate(&Email{})
	if err != nil {
		t.Fail()
	}
}

func TestCreateDomain(t *testing.T) {
	as := NewService()
	err := as.CreateDomain("super.com")
	if err != nil {
		t.Fail()
	}
}

func TestCreateAccount(t *testing.T) {
	as := NewService()
	err := as.CreateAccount("admin", "admin", "super.com")
	if err != nil {
		t.Fail()
	}
}

func TestAuthAccount(t *testing.T) {
	as := NewService()
	err := as.Authorize("admin@super.com", "admin")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetDomain(t *testing.T) {
	as := NewService()
	domain, err := as.FindDomainByName("super.com")
	if err != nil {
		t.Fail()
	}
	t.Log(domain)
}

func TestValidateAccount(t *testing.T) {
	as := NewService()
	acc, err := as.FindAccountByName("admin@super.com")
	if err == nil {
		v := as.ValidateAccount(*acc)
		t.Log(v)

	}
}
