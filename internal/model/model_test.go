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
	if err = db.AutoMigrate(&SsdeepHash{}); err != nil {
		t.Fail()
	}

	/*	fs := []FilterField{
			FilterField{
				Name:       "ClientIP",
				Description:   "client ip",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Sender",
				Description:   "client real ip",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "HeaderFrom",
				Description:   "from address in the header",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Nick",
				Description:   "nick of from address",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Rcpt",
				Description:   "Receipts of the mail",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Size",
				Description:   "Size of the mail",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Mailer",
				Description:   "Mailer of the mail",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Subject",
				Description:   "Subject of the mail",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Text",
				Description:   "Text content or cleaned html content of the mail",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "Html",
				Description:   "Raw html content of the mail",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "TextHash",
				Description:   "Text hash in ssdeep, maximum 2000 characters",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "AttachName",
				Description:   "Attach names in the mail, split with ';'",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "AttachHash",
				Description:   "Attach hashes in ssdeep, split with ';'",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "AttachMd5",
				Description:   "Attach hashes in md5, split with ';'",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "AttachContent",
				Description:   "Content of the attach maybe parsed",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
			FilterField{
				Name:       "URL",
				Description:   "URL of the text content and html content",
				AccountID:  1,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     1,
			},
		}
		for _, f := range fs {
			db.Create(&f)
		}
	*/

}

func TestCreateDomain(t *testing.T) {
	req := CreateDomainRequest{
		Name:        "super.com",
		Description: "test domain",
	}
	err := CreateDomain(req)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestCreateAccount(t *testing.T) {
	req := CreateAccountRequest{
		Name:                "admin",
		Password:            "123456",
		PasswordExpiredTime: time.Now().Add(time.Hour * 24 * 30 * 24),
	}

	err := CreateAccount(req)
	if err != nil {
		t.Fail()
	}
}

func TestAuthAccount(t *testing.T) {
	acc, err := GetAccountByID(1)
	if err != nil {
		t.Fatal(err)
	}
	acc.Password, err = GeneratePassword("123456")
	if err != nil {
		t.Fatal(err)
	}
	db.Save(&acc)

	_, err = Authorize("root@localhost", "123456")
	//_, err := Authorize("admin@super.com", "admin")
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

func TestRuleConvert(t *testing.T) {
	var rule FilterRule
	err := db.Model(&rule).Where("id=?", 1).Scan(&rule).Error
	if err != nil {
		t.Fatal(err)
	}
	drl, err := rule.Convert2DRL()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(drl)
}
