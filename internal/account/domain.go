package account

import "time"

type Domain struct {
	Id          int64     `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Name        string    `gorm:"unique" json:"name"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	Deleted     bool      `json:"deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}
