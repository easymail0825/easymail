package storage

import (
	"easymail/internal/model"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// LocalStorage is a Storager implementation that saves files locally.
// AppRoot should read from global configure
type LocalStorage struct {
	AppRoot string
	BaseDir string
	lock    sync.Mutex
}

func (s *LocalStorage) Query(accID, folderID int64, orderField, orderDir string, page, pageSize int) (total, news int64, emails []model.Email, err error) {
	emails = make([]model.Email, 0)
	query := db.Model(&emails).Where("account_id = ?", accID).Where("folder_id = ?", folderID)
	err = query.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Model(&emails).Where("account_id = ?", accID).Where("folder_id = ?", folderID).Where("read_status=?", 0).Count(&news).Error
	if err != nil {
		return
	}
	if orderField != "" && orderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", orderField, orderDir))
	}
	err = query.Limit(pageSize).Offset(page).Find(&emails).Error
	return
}

// NewLocalStorage creates a new LocalStorage instance.
func NewLocalStorage(appRoot, baseDir string) *LocalStorage {
	return &LocalStorage{
		AppRoot: appRoot,
		BaseDir: baseDir,
		lock:    sync.Mutex{},
	}
}

// Save saves the content of the file at the given filename.
func (s *LocalStorage) Save(accountName string, email *model.Email, content io.Reader) (identify string, err error) {
	// check model status
	if len(accountName) < 1 {
		return "", errors.New("model name is empty")
	}
	var acc *model.Account
	acc, err = model.FindAccountByName(accountName)
	if err != nil {
		return "", errors.New("model not found")
	}
	if !model.ValidateAccount(*acc) {
		return "", errors.New("model is disabled")
	}

	// check domain status
	var domain *model.Domain
	domain, err = model.FindDomainByID(acc.DomainID)
	if err != nil {
		return "", errors.New("domain not found")
	}
	if !model.ValidateDomain(*domain) {
		return "", errors.New("domain is disabled")
	}

	// jobid is uuid
	jobID, err := uuid.Parse(email.JobID)
	if err != nil {
		return "", errors.New("email job id is invalid")
	}

	// generate identify,
	// domain/model/jobid_%32/jobid_%64/jobid
	hashID := jobID.ID()
	filePath := filepath.Join(s.BaseDir, fmt.Sprintf("%s/%s/%d/%d", domain.Name, acc.Username, hashID%32, hashID%64))

	// make directory if not exists
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return "", err
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	// save file
	filePath = filepath.Join(filePath, email.JobID)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	if err != nil {
		return "", err
	}
	email.Identity = filePath
	email.AccountID = acc.ID
	email.SaveTime = time.Now()
	email.SavePath = filePath

	// update mail database
	err = model.SaveEmail(email)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

type Attach struct {
	Name        string
	ContentType string
	Size        int64
}

type MailContent struct {
	Sender    mail.Address
	Recipient []mail.Address
	Text      string
	Html      string
	Attaches  []Attach
}

var AddressRe = regexp.MustCompile("(.*) <(.*)>")

func (s *LocalStorage) Read(filename string) (m *MailContent, err error) {
	m = &MailContent{}
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	email, err := enmime.ReadEnvelope(fd)
	if err != nil {
		return nil, err
	}

	m.Text = email.Text
	m.Html = email.HTML
	sender, err := mail.ParseAddress(email.GetHeader("From"))

	for _, to := range email.GetHeaderValues("To") {
		if recipient, err := mail.ParseAddress(to); err == nil {
			m.Recipient = append(m.Recipient, *recipient)
		}
	}

	m.Sender = *sender
	for _, cc := range email.GetHeaderValues("Cc") {
		if recipient, err := mail.ParseAddress(cc); err != nil {
			continue
		} else {
			m.Recipient = append(m.Recipient, *recipient)
		}
	}

	for _, att := range email.Attachments {
		attach := Attach{
			Name:        att.FileName,
			ContentType: att.ContentType,
			Size:        int64(len(att.Content)),
		}
		m.Attaches = append(m.Attaches, attach)
	}
	return m, nil
}

type AttachBuffer struct {
	Name string
	Data []byte
}

func (s *LocalStorage) GetAttach(mailPath, attachName string, all bool) (data []AttachBuffer, err error) {
	fd, err := os.Open(mailPath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	email, err := enmime.ReadEnvelope(fd)
	if err != nil {
		return nil, err
	}

	data = make([]AttachBuffer, 0, len(email.Attachments))
	if all {
		for _, att := range email.Attachments {
			data = append(data, AttachBuffer{att.FileName, att.Content})
		}
		return
	} else {
		for _, att := range email.Attachments {
			if !all && att.FileName == attachName {
				data = append(data, AttachBuffer{att.FileName, att.Content})
				return
			}
		}
	}
	return nil, errors.New("not found")
}
