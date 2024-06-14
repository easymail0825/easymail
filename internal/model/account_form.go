package model

import "time"

type IndexDomainRequest struct {
	DataTableRequest
	Keyword string `json:"keyword"`
}

type IndexDomainResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	TotalAccount int       `json:"totalAccount"`
	MX           string    `json:"mx"`
	SPF          string    `json:"spf"`
	DMARC        string    `json:"dmarc"`
	Status       int       `json:"status"`
	CreateTime   time.Time `json:"createTime"`
}

type CreateDomainRequest struct {
	//ID          int    `json:"id"`
	Name        string `json:"domainName" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"required,min=3,max=255"`
}

type IndexAccountRequest struct {
	DataTableRequest
	DomainID int    `json:"domainID"`
	Keyword  string `json:"keyword"`
	Status   int    `json:"status"`
}

type IndexAccountResponse struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Status       int       `json:"status"`
	CreateTime   time.Time `json:"createTime"`
	StorageQuota int64     `json:"storageQuota"`
	StorageUsage int64     `json:"storageUsage"`
	MailQuantity int64     `json:"mailQuantity"`
	MailUsage    int64     `json:"mailUsage"`
	ExpiredTime  time.Time `json:"expiredTime"`
}

type CreateAccountRequest struct {
	Name                string    `json:"accountName" binding:"required,min=3,max=64"`
	DomainID            int64     `json:"domainID" binding:"required"`
	Password            string    `json:"password" binding:"required,min=3,max=64"`
	PasswordAgain       string    `json:"passwordRepeat" binding:"required,min=3,max=64"`
	StorageQuota        int64     `json:"storageQuota" binding:"required,min=-1,max=100000"`
	PasswordExpired     string    `json:"passwordExpired"`
	PasswordExpiredTime time.Time `json:"passwordExpiredTime,omitempty"`
}

type EditAccountRequest struct {
	ID                  int64     `json:"accountID" binding:"required"`
	Password            string    `json:"editPassword"`
	StorageQuota        string    `json:"editStorageQuota"`
	PasswordExpired     string    `json:"editPasswordExpired"`
	StorageQuotaNumber  int64     `json:"storageQuotaNumber,omitempty"`
	PasswordExpiredTime time.Time `json:"passwordExpiredTime,omitempty"`
}
