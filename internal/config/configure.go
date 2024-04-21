package config

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type DataType uint

const (
	DataTypeString DataType = iota
	DataTypeInt
	DataTypeFloat
	DataTypeBool
	DataTypeNull
)

type Configure struct {
	gorm.Model
	Name     string `gorm:"index:idx_name,unique"`
	Value    string
	DataType DataType
	ParentID uint       `gorm:"index:idx_name,unique"`
	Parent   *Configure `gorm:"foreignkey:ParentID"`
}

func CreateRoot(name string) (err error) {
	now := time.Now()
	return db.Exec("INSERT INTO `configures` (`name`, `value`, `data_type`, `parent_id`, `created_at`, `updated_at`) "+
		"VALUES (?, ?, ?, ?, ?, ?)", name, "", nil, nil, now, now).Error
}

func FindConfigureByName(name string, pid uint) (c *Configure, err error) {
	query := db.Model(&c).Where("name = ?", name)
	if pid > 0 {
		query = query.Where("parent_id = ?", pid)
	}
	err = query.First(&c).Error
	if err != nil {
		return nil, err
	}
	return
}

func CreateNode(parent, name string, value string, dataType DataType) (*Configure, error) {
	now := time.Now()
	p, err := FindConfigureByName(parent, 0)
	if err != nil {
		return nil, err
	}

	c, err := FindConfigureByName(name, p.ID)
	if err == nil || c != nil {
		return nil, errors.New("name already exists, " + name)
	}

	c = &Configure{
		Name:     name,
		Value:    value,
		DataType: dataType,
		ParentID: p.ID,
		Parent:   p,
	}
	c.CreatedAt = now
	c.UpdatedAt = now

	err = db.Create(c).Error
	if err != nil {
		return nil, err
	}
	return c, nil
}

func CreateNodeFromParent(parent *Configure, name string, value string, dataType DataType) (*Configure, error) {
	return CreateNode(parent.Name, name, value, dataType)
}

func GetConfigure(names ...string) (*Configure, error) {
	var pid uint
	var err error
	var c *Configure
	for i, name := range names {
		c, err = FindConfigureByName(name, pid)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			pid = c.ID
		}
	}
	return c, nil
}
