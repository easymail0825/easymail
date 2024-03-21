package account

import (
	"easymail/internal/database"
	"errors"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Service struct {
	db *gorm.DB
}

func NewService() *Service {
	return &Service{db: database.DB}
}

func (s *Service) FindDomainByID(id int64) (domain *Domain, err error) {
	err = s.db.Model(&domain).Where("id = ?", id).Scan(&domain).Error
	if err == gorm.ErrRecordNotFound {
		return domain, errors.New("domain not exists")
	}
	return domain, nil
}

func (s *Service) FindDomainByName(name string) (domain *Domain, err error) {
	err = s.db.Model(&domain).Where("name = ?", name).Scan(&domain).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("domain not exists")
	}
	return domain, nil
}

func (s *Service) FindValidateDomainByName(name string) (domain *Domain, err error) {
	err = s.db.Model(&domain).Where("name = ? AND active=? AND deleted=?", name, true, false).Scan(&domain).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("domain not exists")
	}
	return domain, nil
}

/*
FindAccountByID
@Desc
Find an account by id.
*/
func (s *Service) FindAccountByID(id int64) (a *Account, err error) {
	err = s.db.Model(&a).Where("id= ?", id).Scan(&a).Error
	if err != nil {
		return a, err
	}
	return a, err
}

/*
FindAccountByName
@Desc
Find an account by username with the given domain.
if domain is not exist, use default domain.
*/
func (s *Service) FindAccountByName(username string) (a *Account, err error) {
	if !emailRegex.MatchString(username) {
		return nil, errors.New("invalid username")
	}

	username = strings.ToLower(username)
	parts := strings.SplitN(username, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid username")
	}

	domain, err := s.FindDomainByName(parts[1])
	if err != nil || domain == nil || domain.Id <= 0 {
		return nil, errors.New("domain not exists")
	}

	a = &Account{}
	err = s.db.Model(&a).Where("username=? AND domain_id=? AND active=? AND deleted=?", parts[0], domain.Id, true, false).Take(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("account not exists")
	}

	return a, nil
}

func (s *Service) Authorize(username, password string) (err error) {
	acc, err := s.FindAccountByName(username)
	if err != nil {
		return err
	}

	if err := VerifyPassword(acc.Password, password); err != nil {
		return errors.New("invalid password")
	}

	return nil
}

func (s *Service) CreateDomain(name string) (err error) {
	if _, err := s.FindDomainByName(name); err == nil {
		return errors.New("domain already exists")
	}

	domain := &Domain{
		Name:      name,
		Active:    true,
		CreatedAt: time.Now(),
	}
	if err := s.db.Create(domain).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateAccount(name, password, domainName string) (err error) {
	domain, err := s.FindDomainByName(domainName)
	if err != nil || domain == nil || domain.Id <= 0 {
		return errors.New("domain not exists")
	}

	if _, err := s.FindAccountByName(name); err == nil {
		return errors.New("account already exists")
	}

	passwordHash, err := GeneratePassword(password)
	if err != nil {
		return nil
	}

	account := &Account{
		Username:  name,
		Password:  passwordHash,
		Domain:    domain,
		Active:    true,
		CreatedAt: time.Now(),
	}
	if err := s.db.Create(account).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) ValidateDomain(domain Domain) bool {
	return domain.Active && !domain.Deleted
}

func (s *Service) ValidateAccount(acc Account) bool {
	return acc.Active && !acc.Deleted
}

func (s *Service) SaveEmail(email *Email) (err error) {
	// check account_id, jobid, identity must exist
	if email.AccountID <= 0 || email.JobID == "" || email.Identity == "" {
		return errors.New("invalid email")
	}
	return s.db.Model(&email).Create(email).Error
}
