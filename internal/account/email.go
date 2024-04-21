package account

import (
	"gorm.io/gorm"
	"time"
)

type ReadStatus uint8

const (
	UnRead ReadStatus = iota
	WebRead
	ImapRead
)

type Email struct {
	ID         int64  `gorm:"primaryKey;AUTO_INCREMENT"`
	JobID      string `gorm:"char(16);not null;index"`
	AccountID  int64  `gorm:"not null;index"`
	Date       time.Time
	Sender     string `gorm:"null" sql:"varchar(255)"`
	Recipient  string `gorm:"not null" sql:"varchar(1024)"` //clean recipient joined by ,
	CarbonCopy string `gorm:"null" sql:"varchar(1024)"`
	BlindCopy  string `gorm:"null" sql:"varchar(1024)"`
	Subject    string `gorm:"null" sql:"varchar(255)"`
	MailTime   time.Time
	SaveTime   time.Time
	Size       int64      `gorm:"default 0"`
	FolderId   int64      `gorm:"not null;default 0"`
	ReadStatus ReadStatus `gorm:"not null;default 0"`
	Deleted    bool       `gorm:"not null;default false"`
	DeleteTime time.Time
	Identity   string `gorm:"not null"`
}

func GetMailQuantity(accountId int64) (total int64, err error) {
	err = db.Model(&Email{}).Where("account_id = ?", accountId).Count(&total).Error
	return
}

func GetMailUsage(accountId int64) (total int64, err error) {
	err = db.Model(&Email{}).Select("SUM(size) as total").Where("account_id = ?", accountId).Scan(&total).Error
	return
}

/*
MarkAccountMailDeleted
Describe: mark mails of the account as deleted
Input: accountId
Output: err
*/
func MarkAccountMailDeleted(accountId int64, db *gorm.DB) (err error) {
	err = db.Model(&Email{}).Where("account_id = ?", accountId).Update("deleted", true).Error
	return
}
