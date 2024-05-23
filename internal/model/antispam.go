package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type AntispamAction uint8

const (
	AntispamActionAccept AntispamAction = iota
	AntispamActionTrash
	AntispamActionDefer
	AntispamActionReject
	AntispamActionDiscard
	AntispamActionQuarantine
)

type FilterRule struct {
	ID               int64          `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Priority         int64          `gorm:"type:int(11);default(0)" json:"priority"`
	Describe         string         `gorm:"type:varchar(255)" json:"describe"`
	Action           AntispamAction `gorm:"type:int(8);default(0)" json:"action"`
	ClientIP         string         `gorm:"type:varchar(1024)" json:"client_ip"`
	Sender           string         `gorm:"type:varchar(255)" json:"sender"`
	Nick             string         `gorm:"type:varchar(255)" json:"nick"`
	Rcpt             string         `gorm:"type:varchar(255)" json:"rcpt"`
	Size             string         `gorm:"type:varchar(255)" json:"size"`
	Mailer           string         `gorm:"type:varchar(255)" json:"mailer"`
	Subject          string         `gorm:"type:varchar(255)" json:"subject"`
	Text             string         `gorm:"type:varchar(1024)" json:"text"`
	Html             string         `gorm:"type:varchar(1024)" json:"html"`
	TextHash         string         `gorm:"type:varchar(255)" json:"text_hash"`
	AttachName       string         `gorm:"type:varchar(255)" json:"attach_name"`
	AttachHash       string         `gorm:"type:varchar(255)" json:"attach_hash"`
	AttachContent    string         `gorm:"type:varchar(255)" json:"attach_content"`
	FileNameInAttach string         `gorm:"type:varchar(255)" json:"file_name_in_attach"`
	URL              string         `gorm:"type:varchar(255)" json:"url"`
	Assembly         string         `gorm:"type:varchar(4098)" json:"assembly"`
	AccountID        int            `gorm:"type:int(11);default(0)" json:"account_id"`
	Status           uint8          `gorm:"type:int(8);default(0)" json:"status"` //0=inactive;1=active;2=deleted
	CreateTime       time.Time      `json:"create_time"`
	UpdateTime       time.Time      `json:"update_time"`
	DeleteTime       time.Time      `json:"delete_time"`
}

type FilterLog struct {
	ID         int64          `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	QueueID    string         `gorm:"type:varchar(255)" json:"queue_id"`
	ClientIP   string         `gorm:"type:varchar(64)" json:"client_ip"`
	Sender     string         `gorm:"type:varchar(255)" json:"sender"`
	Nick       string         `gorm:"type:varchar(255)" json:"nick"`
	Rcpt       string         `gorm:"type:varchar(1024)" json:"rcpt"`
	Size       string         `gorm:"type:varchar(255)" json:"size"`
	Mailer     string         `gorm:"type:varchar(255)" json:"mailer"`
	Subject    string         `gorm:"type:varchar(255)" json:"subject"`
	Feature    string         `gorm:"type:mediumtext" json:"feature"`
	Action     AntispamAction `gorm:"type:int(8);default(0)" json:"action"`
	CreateTime time.Time
}

func formatRuleDefine(s string) (string, error) {
	d := strings.SplitN(s, "::", 2)
	if len(d) != 2 {
		return "", errors.New("invalid rule define")
	}
	operator := strings.ToLower(d[0])
	switch operator {
	case "contains":
		return fmt.Sprintf(".Contains(\"%s\")", d[1]), nil
	case "hasprefix":
		return fmt.Sprintf(".HasPrefix(\"%s\")", d[1]), nil
	case "hashsuffix":
		return fmt.Sprintf(".HasSuffix(\"%s\")", d[1]), nil
	case "equals":
		return fmt.Sprintf("==\"%s\"", d[1]), nil
	case "notequals":
		return fmt.Sprintf("!=\"%s\"", d[1]), nil
	case "gt":
		return fmt.Sprintf(">%s", d[1]), nil
	case "lt":
		return fmt.Sprintf("<%s", d[1]), nil
	case "gte":
		return fmt.Sprintf(">=%s", d[1]), nil
	case "lte":
		return fmt.Sprintf("<=%s", d[1]), nil
	case "equal":
		return fmt.Sprintf("==%s", d[1]), nil
	case "notequal":
		return fmt.Sprintf("!=%s", d[1]), nil

	}
	return "", errors.New("invalid rule define")
}

func (r FilterRule) Convert2DRL() (drl string, err error) {
	sb := strings.Builder{}
	sb.WriteString(
		fmt.Sprintf("rule rule_%d \"%s\" salience %d {\n",
			r.ID,
			strings.Replace(r.Describe, "\"", "", -1),
			r.Priority,
		),
	)
	sb.WriteString(fmt.Sprintf("\twhen\n"))
	condition := make([]string, 0)
	//The field has been defined, and it must be: operator::operator factor.
	// it must define only one time
	if r.ClientIP != "" {
		if t, err := formatRuleDefine(r.ClientIP); err == nil {
			condition = append(condition, fmt.Sprintf("feature.client_ip%s", t))
		}
	}
	if r.Assembly != "" {
		aList := strings.Split(r.Assembly, ";;")
		for _, a := range aList {
			condition = append(condition, fmt.Sprintf("feature.%s", a))
		}
	}
	if len(condition) == 0 {
		return "", errors.New("rule condition is empty")
	}
	sb.WriteString(fmt.Sprintf("\t\t%s\n", strings.Join(condition, " && ")))
	sb.WriteString(fmt.Sprintf("\tthen\n"))
	sb.WriteString(fmt.Sprintf("\t\tantispam.RuleID=%d;\n", r.ID))
	sb.WriteString(fmt.Sprintf("\t\tantispam.Action=%d;\n", r.Action))
	sb.WriteString(fmt.Sprintf("\t\tRetract(\"rule_%d\");\n", r.ID))
	sb.WriteString(fmt.Sprintf("\t\tComplete();\n}\n"))
	return sb.String(), nil
}

func GetFilterRules() (rules []FilterRule, err error) {
	err = db.Model(&rules).Where("status=?", 1).Order("priority desc").Find(&rules).Error
	return
}

func GetLastTimeOfRule() (last time.Time, err error) {
	// 从filter_rules表中，只选择出update_time，然后赋值last
	err = db.Model(&FilterRule{}).Select("update_time").Where("status=?", 1).
		Order("update_time desc").Limit(1).Scan(&last).Error
	return
}
