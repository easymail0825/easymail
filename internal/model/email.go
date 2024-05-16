package model

import (
	"bytes"
	"github.com/jhillyerd/enmime"
	"net/mail"
	"time"
)

type ReadStatus uint8

const (
	UnRead ReadStatus = iota
	WebRead
	ImapRead
)

type FolderID uint8

const (
	Inbox FolderID = iota
	Sent
	Draft
	Trash
	Spam
	Quarantine
)

type Email struct {
	ID         int64  `gorm:"primaryKey;AUTO_INCREMENT"`
	JobID      string `gorm:"char(128);not null;index"`
	QueueID    string `gorm:"char(16);not null;index"`
	AccountID  int64  `gorm:"not null;index"`
	Date       time.Time
	Sender     string `gorm:"null" sql:"varchar(255)"`
	Recipient  string `gorm:"not null" sql:"varchar(1024)"` //clean recipient joined by ,
	CarbonCopy string `gorm:"null" sql:"varchar(1024)"`
	BlindCopy  string `gorm:"null" sql:"varchar(1024)"`
	Subject    string `gorm:"null" sql:"varchar(255)"`
	MailTime   time.Time
	SaveTime   time.Time
	SavePath   string
	Size       int64      `gorm:"default 0"`
	FolderId   int64      `gorm:"not null;default 0"`
	ReadStatus ReadStatus `gorm:"not null;default 0"`
	Deleted    bool       `gorm:"not null;default false"`
	DeleteTime time.Time
	Identity   string `gorm:"not null"`
}

type Attachment struct {
	Name        string
	ContentType string
	Data        []byte
}

func GetMailQuantity(accID int64) (total int64, err error) {
	err = db.Model(&Email{}).Where("account_id = ?", accID).Count(&total).Error
	return
}

func GetMailUsage(accID int64) (total int64, err error) {
	err = db.Model(&Email{}).Select("SUM(size) as total").Where("account_id = ?", accID).Scan(&total).Error
	return
}

func MarkRead(accID int64, mailID int64, readSource ReadStatus) error {
	return db.Model(&Email{}).Where("account_id = ? and id = ? and read_status=?", accID, mailID, UnRead).Update("read_status", readSource).Error
}

func GetMail(accID int64, mailID int64) (mail *Email, err error) {
	mail = &Email{}
	err = db.Where("account_id = ? and id = ?", accID, mailID).First(mail).Error
	return
}

func CreateMail(sender mail.Address, receipts []mail.Address, subject, text, html string, attaches []Attachment) ([]byte, error) {
	mailer := "easymail 1.0.0"

	builder := enmime.Builder().
		From(sender.Name, sender.Address).
		Subject(subject).
		Header("X-Mailer", mailer)

	for _, to := range receipts {
		builder = builder.To(to.Name, to.Address)
	}
	if text != "" {
		builder = builder.Text([]byte(text))
	}
	if html != "" {
		builder = builder.HTML([]byte(html))
	}

	if attaches != nil {
		for _, att := range attaches {
			builder = builder.AddAttachment(att.Data, att.ContentType, att.Name)
		}
	}

	buf := &bytes.Buffer{}
	root, err := builder.Build()
	if err != nil {
		return nil, err
	}
	err = root.Encode(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
