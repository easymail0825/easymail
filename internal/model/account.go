package model

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

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@([a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}|localhost)$`)

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

func CreateAccount(req CreateAccountRequest) (err error) {
	domain, err := FindDomainByID(req.DomainID)
	if err != nil || domain == nil || domain.ID <= 0 {
		return errors.New("domain not exists")
	}

	if _, err := FindAccountByName(req.Name); err == nil {
		return errors.New("model already exists")
	}

	passwordHash, err := GeneratePassword(req.Password)
	if err != nil {
		return nil
	}

	account := &Account{
		Username:           req.Name,
		Password:           passwordHash,
		Domain:             domain,
		Active:             true,
		CreateTime:         time.Now(),
		UpdateTime:         time.Now(),
		StorageQuota:       req.StorageQuota,
		PasswordExpireTime: req.PasswordExpiredTime,
	}
	if err := db.Create(account).Error; err != nil {
		return err
	}

	return nil
}

/*
FindAccountByID
@Desc
Find an model by id.
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
Find an model by username with the given domain.
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
		return nil, errors.New("model not exists")
	}

	return a, nil
}

func CountDomainAccount(id int64) (total int64, err error) {
	err = db.Model(&Account{}).Where("domain_id=?", id).Count(&total).Error
	return total, err
}

func Index(did, status int, keyword, orderField, orderDir string, page, pageSize int) (int64, []Account, error) {
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

	if orderField != "" && orderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", orderField, orderDir))
	}
	query = query.Offset(page).Limit(pageSize)
	err := query.Find(&accounts).Error
	if err != nil {
		return 0, nil, err
	}
	return total, accounts, nil
}

func ToggleAccount(id int64) error {
	account := Account{}
	if err := db.First(&account, id).Error; err != nil {
		return err
	}
	if err := db.Model(&account).Where("id", id).Update("active", !account.Active).Error; err != nil {
		return err
	}
	return nil
}

func SaveEmail(email *Email) (err error) {
	// check account_id, jobid, identity must exist
	if email.AccountID <= 0 || email.JobID == "" || email.Identity == "" {
		return errors.New("invalid email")
	}
	return db.Model(&email).Create(email).Error
}

func DeleteMail(accID int64, mailID int64) error {
	return db.Model(&Email{}).Where("account_id = ? AND id = ?", accID, mailID).Updates(
		map[string]interface{}{"Deleted": true, "DeleteTime": time.Now()},
	).Error
}

func MoveMail(accID int64, mailID int64, fid FolderID) error {
	return db.Model(&Email{}).Where("account_id = ? AND id = ?", accID, mailID).Update("folder_id", fid).Error
}

func DeleteAccount(id int64) (err error) {
	// transmit
	tx := db.Begin()

	// mark mails deleted
	err = tx.Model(&Email{}).Where("account_id = ?", id).Updates(
		map[string]interface{}{"Deleted": true, "DeleteTime": time.Now()},
	).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// mark model deleted
	if err = tx.Model(&Account{}).Where("id = ?", id).Updates(map[string]interface{}{"Deleted": true, "DeleteTime": time.Now()}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func EditAccount(req EditAccountRequest) error {
	tx := db.Begin()
	if len(req.Password) >= 6 && len(req.Password) <= 64 {
		hashPassword, err := GeneratePassword(req.Password)
		if err != nil {
			return err
		}
		if err := tx.Model(&Account{}).Where("id = ?", req.ID).Update("password", hashPassword).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if req.StorageQuotaNumber >= -1 && req.StorageQuotaNumber < 1000000 {
		if err := tx.Model(&Account{}).Where("id = ?", req.ID).Update("storage_quota", req.StorageQuotaNumber).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if req.PasswordExpiredTime.IsZero() {
		if err := tx.Model(&Account{}).Where("id = ?", req.ID).Update("password_expire_time", nil).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Model(&Account{}).Where("id = ?", req.ID).Update("password_expire_time", req.PasswordExpiredTime).Error; err != nil {
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
