package account

import (
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
	IsDeleted  bool       `gorm:"not null;default false"`
	DeleteTime time.Time
	Identity   string `gorm:"not null"`
}
