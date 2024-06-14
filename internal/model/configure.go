package model

import (
	"errors"
	"fmt"
	"time"
)

type DataType uint

const (
	DataTypeString DataType = iota
	DataTypeInt
	DataTypeFloat
	DataTypeBool
	DataTypeNull
	DataTypeChild
)

type Configure struct {
	ID          uint   `gorm:"autoIncrement;primaryKey"`
	Name        string `gorm:"type:varchar(255);uniqueIndex:idx_name"`
	Value       string `gorm:"type:varchar(255)"`
	DataType    DataType
	ParentID    uint       `gorm:"uniqueIndex:idx_name"`
	Parent      *Configure `gorm:"foreignkey:ParentID"`
	Private     bool       `gorm:"default:true" json:"private"`
	CreateTime  time.Time
	UpdateTime  time.Time
	Description string `gorm:"type:varchar(255)"`
}

func GetConfigureByID(id uint) (*Configure, error) {
	c := Configure{}
	err := db.Model(&c).Where("id=?", id).Scan(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, err
}

func GetConfigureByName(name string, pid uint) (*Configure, error) {
	c := Configure{}
	var err error
	query := db.Model(&c).Where("name = ?", name)
	if pid == 0 {
		query = query.Where("parent_id = ? OR parent_id is NULL", 0)
	} else {
		query = query.Where("parent_id = ?", pid)
	}
	err = query.First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, err
}

func CreateRoot(name string) (err error) {
	now := time.Now()
	return db.Exec("INSERT INTO `configures` (`name`, `value`, `data_type`, `parent_id`, `create_time`, `update_time`, `private`) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?)", name, "", DataTypeChild, nil, now, now, true).Error
}

func CreateNode(parent *Configure, name string, value string, dataType DataType, description string) (*Configure, error) {
	now := time.Now()

	c := &Configure{
		Name:        name,
		Value:       value,
		DataType:    dataType,
		ParentID:    parent.ID,
		Parent:      parent,
		Description: description,
	}
	c.CreateTime = now
	c.UpdateTime = now

	err := db.Create(&c).Error
	if err != nil {
		return nil, err
	}
	return c, nil
}

func GetConfigureByNames(names ...string) (*Configure, error) {
	var err error
	var c *Configure
	var pid uint
	for _, name := range names {
		c, err = GetConfigureByName(name, pid)
		if err != nil {
			return nil, err
		}
		pid = c.ID
	}
	return c, nil
}

func CreateConfigure(value, description string, dateType DataType, names ...string) (*Configure, error) {
	if len(names) != 3 {
		return nil, errors.New("names length must be 3")
	}

	// check root exists, if not, create it first
	root, err := GetConfigureByName(names[0], 0)
	if err != nil || root == nil {
		err = CreateRoot(names[0])
		if err != nil {
			return nil, err
		}
	}
	// check root again
	root, err = GetConfigureByName(names[0], 0)
	if err != nil || root == nil {
		return nil, err
	}

	// create child
	var child *Configure
	child, err = GetConfigureByName(names[1], root.ID)
	if err != nil {
		if _, err = CreateNode(root, names[1], "", DataTypeChild, description); err != nil {
			return nil, err
		}
	}
	// check child again
	child, err = GetConfigureByName(names[1], root.ID)
	if err != nil || child == nil {
		return nil, err
	}

	// check child of child
	var childChild *Configure
	childChild, err = GetConfigureByName(names[2], child.ID)
	if err != nil || childChild == nil {
		if _, err = CreateNode(child, names[2], value, dateType, description); err != nil {
			return nil, err
		}
	}
	// check child of child again
	childChild, err = GetConfigureByName(names[2], child.ID)
	return childChild, err
}

func UpdateConfigure(node Configure) error {
	return db.Model(&node).Where("id=?", node.ID).Updates(Configure{
		Value:       node.Value,
		Description: node.Description,
		UpdateTime:  time.Now(),
	}).Error
}

func GetRootConfigureRootNodes() (data []Configure, err error) {
	err = db.Model(&data).Where("parent_id = ? or parent_id is null", 0).Order("name").Find(&data).Error
	return
}

func GetSubConfigureByParentId(id uint) (data []Configure, err error) {
	err = db.Model(&data).Where("parent_id = ?", id).Order("name").Find(&data).Error
	return
}

func GetConfigureByParentId(id, subNodeID uint, keyword, orderField, orderDir string, page, pageSize int) (total int64, nodes []Configure, err error) {
	// query subNodes first
	subNodes, err := GetSubConfigureByParentId(id)
	nodeMap := make(map[uint]*Configure)

	// union node id
	ids := make([]uint, 0)
	if subNodeID == 0 {
		for _, node := range subNodes {
			ids = append(ids, node.ID)
			nodeMap[node.ID] = &node
		}
	} else {
		node, err := GetConfigureByID(subNodeID)
		if err == nil {
			nodeMap[subNodeID] = node
		}
		ids = append(ids, subNodeID)
	}

	// then query 3rd level node
	query := db.Model(&nodes).Where("parent_id in (?)", ids)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)

	if orderField != "" && orderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", orderField, orderDir))
	}
	query = query.Offset(page).Limit(pageSize)
	err = query.Find(&nodes).Error
	if err != nil {
		return 0, nil, err
	}

	// fill parent
	for i := range nodes {
		nodes[i].Parent = nodeMap[nodes[i].ParentID]
	}

	return total, nodes, nil
}

func (c *Configure) Save() error {
	return db.Save(c).Error
}
