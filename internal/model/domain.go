package model

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

/*
Domain
status = 0 indicated active=true deleted=0
status = 1 indicated active=false deleted=true
status = 2 indicated active=false deleted=false
*/
type Domain struct {
	ID          int64     `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Name        string    `gorm:"unique" json:"name"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	Deleted     bool      `json:"deleted"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	DeleteTime  time.Time `json:"delete_time"`
}

func FindDomainByID(id int64) (domain *Domain, err error) {
	err = db.Model(&domain).Where("id = ?", id).Scan(&domain).Error
	if err == gorm.ErrRecordNotFound {
		return domain, errors.New("domain not exists")
	}
	return domain, nil
}

func FindDomainByName(name string) (domain *Domain, err error) {
	err = db.Model(&domain).Where("name = ?", name).First(&domain).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("domain not exists")
	}
	return domain, nil
}

func FindAllValidateDomain() (domains []Domain, err error) {
	err = db.Model(&domains).Where("active=? AND deleted=?", true, false).Order("name").Scan(&domains).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("domain not exists")
	}
	return domains, nil
}

func FindValidateDomainByName(name string) (domain *Domain, err error) {
	err = db.Model(&domain).Where("name = ? AND active=? AND deleted=?", name, true, false).Scan(&domain).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("domain not exists")
	}
	return domain, nil
}

func CreateDomain(req CreateDomainRequest) (err error) {
	if d, _ := FindDomainByName(req.Name); d != nil {
		return errors.New("domain already exists")
	}

	domain := &Domain{
		Name:        req.Name,
		Description: req.Description,
		Active:      true,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	if err := db.Create(domain).Error; err != nil {
		return err
	}

	return nil
}

func ValidateDomain(domain Domain) bool {
	return domain.Active && !domain.Deleted
}

func IndexDomain(keyword, orderField, orderDir string, page, pageSize int) (int64, []Domain, error) {
	domains := make([]Domain, 0)
	query := db.Model(&domains)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	if orderField != "" && orderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", orderField, orderDir))
	}
	query = query.Offset(page).Limit(pageSize)
	err := query.Find(&domains).Error
	if err != nil {
		return 0, nil, err
	}
	return total, domains, nil
}

func ToggleDomainActive(id int64) error {
	domain := Domain{}
	if err := db.First(&domain, id).Error; err != nil {
		return err
	}
	if err := db.Model(&domain).Where("id", id).Updates(map[string]interface{}{
		"Active":     !domain.Active,
		"UpdateTime": time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}

/*
PhysicalDeleteDomain
@Description: Physical delete the domain, include mails, model, but it will be deleted in background
*/
func PhysicalDeleteDomain(id int64) (err error) {
	// transmit
	tx := db.Begin()

	// todo delete mails

	// delete accounts
	if err = tx.Where("domain_id = ?", id).Delete(&Account{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	//delete domain
	domain := &Domain{
		ID: id,
	}
	if err = tx.Delete(domain).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func DeleteDomain(id int64) error {
	return db.Model(&Domain{}).Where("id = ?", id).Updates(Domain{Deleted: true, UpdateTime: time.Now(), Active: false}).Error
}

func LastUpdateDomain() time.Time {
	d := Domain{}
	err := db.Model(&d).Select("update_time").Order("update_time desc").First(&d).Error
	if err == nil {
		return d.UpdateTime
	}
	return time.Time{}
}
