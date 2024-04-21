package account

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

/*
Account
@Desc
account model
*/
type Account struct {
	ID                 int64  `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Username           string `gorm:"index:idx_username,unique" json:"username"`
	Password           string `json:"password"`
	DomainID           int64  `gorm:"index:idx_username,unique"`
	Domain             *Domain
	Active             bool      `json:"active"`
	Deleted            bool      `json:"deleted"`
	UpdateTime         time.Time `json:"update_time"`
	DeleteTime         time.Time `json:"delete_time"`
	CreateTime         time.Time `json:"create_time"`
	PasswordExpireTime time.Time `json:"password_expire_time"`
	StorageQuota       int64     `json:"storage_quota"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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

// GenerateRandomString generate random string
func GenerateRandomString(length int) (string, error) {
	// make byte array
	//source := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+|<>?")
	source := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	// make random string
	result := make([]rune, length)
	for i := range result {
		result[i] = source[rand.Intn(len(source))]
	}
	return string(result), nil
}

func Authorize(username, password string) (acc *Account, err error) {
	acc, err = FindAccountByName(username)
	if err != nil {
		return nil, err
	}

	if err := VerifyPassword(acc.Password, password); err != nil {
		return nil, errors.New("invalid password")
	}

	return acc, nil
}

func (account *Account) SaveEmail(email *Email) (err error) {
	// check account_id, jobid, identity must exist
	if email.AccountID <= 0 || email.JobID == "" || email.Identity == "" {
		return errors.New("invalid email")
	}
	return db.Model(&email).Create(email).Error
}

func CreateAccount(domainID int64, name, password string, quota int64, expired time.Time) (err error) {
	domain, err := FindDomainByID(domainID)
	if err != nil || domain == nil || domain.ID <= 0 {
		return errors.New("domain not exists")
	}

	if _, err := FindAccountByName(name); err == nil {
		return errors.New("account already exists")
	}

	passwordHash, err := GeneratePassword(password)
	if err != nil {
		return nil
	}

	account := &Account{
		Username:           name,
		Password:           passwordHash,
		Domain:             domain,
		Active:             true,
		CreateTime:         time.Now(),
		StorageQuota:       quota,
		PasswordExpireTime: expired,
	}
	if err := db.Create(account).Error; err != nil {
		return err
	}

	return nil
}

/*
FindAccountByID
@Desc
Find an account by id.
*/
func FindAccountByID(id int64) (a *Account, err error) {
	err = db.Model(&a).Where("id= ?", id).Scan(&a).Error
	if err != nil {
		return a, err
	}
	return a, err
}

func ValidateAccount(acc Account) bool {
	return acc.Active && !acc.Deleted
}

func GetAccountByID(i int) (*Account, error) {
	var acc Account
	err := db.First(&acc, i).Error
	return &acc, err
}

/*
FindAccountByName
@Desc
Find an account by username with the given domain.
if domain is not exist, use default domain.
*/
func FindAccountByName(username string) (a *Account, err error) {
	if !emailRegex.MatchString(username) {
		return nil, errors.New("invalid username")
	}

	username = strings.ToLower(username)
	parts := strings.SplitN(username, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid username")
	}

	domain, err := FindDomainByName(parts[1])
	if err != nil || domain == nil || domain.ID <= 0 {
		return nil, errors.New("domain not exists")
	}

	a = &Account{}
	err = db.Model(&a).Where("username=? AND domain_id=? AND active=? AND deleted=?", parts[0], domain.ID, true, false).Take(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("account not exists")
	}

	return a, nil
}

func CountDomainAccount(id int64) (total int64, err error) {
	err = db.Model(&Account{}).Where("domain_id=?", id).Count(&total).Error
	return total, err
}

func Index(did, status int, keyword, orderFiled, orderDir string, page, pageSize int) (int64, []Account, error) {
	accounts := make([]Account, 0)
	query := db.Model(&accounts).Where("domain_id=?", did)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if status > -1 {
		if status == 0 {
			query = query.Where("active=? AND deleted=?", true, false)
		} else if status == 1 {
			query = query.Where("active=? AND deleted=?", false, false)
		} else if status == 2 {
			query = query.Where("deleted=?", true)
		}
	}

	var total int64
	query.Count(&total)

	if orderFiled != "" && orderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", orderFiled, orderDir))
	}
	query = query.Offset(page).Limit(pageSize)
	err := query.Find(&accounts).Error
	if err != nil {
		return 0, nil, err
	}
	return total, accounts, nil
}

func ToggleAccountActive(id int64) error {
	account := Account{}
	if err := db.First(&account, id).Error; err != nil {
		return err
	}
	if err := db.Model(&account).Where("id", id).Update("active", !account.Active).Error; err != nil {
		return err
	}
	return nil
}

func DeleteAccount(id int64) (err error) {
	// transmit
	tx := db.Begin()

	// mark mails deleted
	err = MarkAccountMailDeleted(id, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// mark account deleted
	if err = tx.Model(&Account{}).Where("id = ?", id).Update("deleted", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func EditAccount(id int64, password string, storageQuota int64, passwordExpiredTime time.Time) error {
	tx := db.Begin()
	if len(password) >= 6 && len(password) <= 64 {
		hashPassword, err := GeneratePassword(password)
		if err != nil {
			return err
		}
		if err := tx.Model(&Account{}).Where("id = ?", id).Update("password", hashPassword).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if storageQuota >= -1 && storageQuota < 1000000 {
		if err := tx.Model(&Account{}).Where("id = ?", id).Update("storage_quota", storageQuota).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if passwordExpiredTime.IsZero() {
		if err := tx.Model(&Account{}).Where("id = ?", id).Update("password_expire_time", nil).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Model(&Account{}).Where("id = ?", id).Update("password_expire_time", passwordExpiredTime).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
