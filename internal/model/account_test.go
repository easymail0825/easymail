package model

import (
	"bytes"
	"github.com/jhillyerd/enmime"
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
	password := "123456"
	hashedPassword := "$2a$10$M9ezUs6WL1PXAEdjXCdMAugMuvw4/rpAvzauPyoeslOI0Ip2DJWCy"
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	t.Log(err)
}

func TestMigrate(t *testing.T) {
	var err error
	if err = db.AutoMigrate(&Domain{}, &Account{}, &Email{}); err != nil {
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
	err := CreateAccount(1, "admin", "123456", -1, time.Now().Add(time.Hour*24*30*24))
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

func TestCreateEmail(t *testing.T) {
	mailer := "easymail 1.0.0"
	html := "<p color='red'>hello world</p>"

	builder := enmime.Builder().
		From("admin", "admin@super.com").
		Subject("this is test email from enmime").
		HTML([]byte(html)).
		Header("X-Mailer", mailer).
		To("", "admin@super.com")

	builder = builder.AddFileAttachment("/home/bobxiao/tmp/pypolicyd-spf-1.3.2.tar.gz")
	buf := &bytes.Buffer{}
	root, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	err = root.Encode(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", buf.String())

}
