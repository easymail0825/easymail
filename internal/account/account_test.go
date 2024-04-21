package account

import (
	"easymail/internal/database"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

// crypt加密用户密码
func TestCrypt(t *testing.T) {
	password := "123456"
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
	db := database.GetDB()
	var err error
	if err = db.AutoMigrate(&Domain{}, &Account{}); err != nil {
		t.Fail()
	}
	err = db.AutoMigrate(&Email{})
	if err != nil {
		t.Fail()
	}
}

func TestCreateDomain(t *testing.T) {
	err := CreateDomain("super.com", "test domain")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestCreateAccount(t *testing.T) {
	err := CreateAccount("super.com", "admin", "123456", -1, time.Now().Add(time.Hour*24*30*24))
	if err != nil {
		t.Fail()
	}
}

func TestAuthAccount(t *testing.T) {
	_, err := Authorize("admin@super.com", "admin")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetDomain(t *testing.T) {
	domain, err := FindDomainByName("super.com")
	if err != nil {
		t.Fail()
	}
	t.Log(domain)
}

func TestValidateAccount(t *testing.T) {
	acc, err := FindAccountByName("admin@super.com")
	if err == nil {
		v := ValidateAccount(*acc)
		t.Log(v)

	}
}

func TestGenerateRandomString(t *testing.T) {
	s, err := GenerateRandomString(16)

	if err != nil {
		t.Fail()
	}
	t.Log(s)
}
