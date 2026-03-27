package model

import "time"

type Admin struct {
	ID         int64 `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	DomainID   int64
	AccountID  int64
	IsSuper    bool // super can manage all domain, otherwise only manage own domain
	CreateTime time.Time
}
