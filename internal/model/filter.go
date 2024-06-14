package model

import (
	context "context"
	"easymail/internal/database"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type FilterCategory uint8

const (
	FilterCategoryAll FilterCategory = iota
	FilterCategoryUnknown
	FilterCategoryHam
	FilterCategorySpam
)

type FilterAction uint8

const (
	FilterActionAccept FilterAction = 1 + iota
	FilterActionTrash
	FilterActionDefer
	FilterActionReject
	FilterActionDiscard
	FilterActionQuarantine
)

type FilterStage uint8

const (
	FilterStageConnect FilterStage = iota
	FilterStageHelo
	FilterStageMailFrom
	FilterStageRcptTo
	FilterStageHeader
	FilterStageData
)

type FilterRule struct {
	ID            int64        `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Priority      int64        `gorm:"type:int(11);default(0)" json:"priority"`
	Description   string       `gorm:"type:varchar(255)" json:"description"`
	Action        FilterAction `gorm:"type:int(8);default(0)" json:"action"`
	ClientIP      string       `gorm:"type:varchar(1024)" json:"client_ip"`
	Sender        string       `gorm:"type:varchar(255)" json:"sender"`
	HeaderFrom    string       `gorm:"type:varchar(255)" json:"header_from"`
	Nick          string       `gorm:"type:varchar(255)" json:"nick"`
	Rcpt          string       `gorm:"type:varchar(255)" json:"rcpt"`
	Size          string       `gorm:"type:varchar(255)" json:"size"`
	Mailer        string       `gorm:"type:varchar(255)" json:"mailer"`
	Subject       string       `gorm:"type:varchar(255)" json:"subject"`
	Text          string       `gorm:"type:varchar(1024)" json:"text"`
	Html          string       `gorm:"type:varchar(1024)" json:"html"`
	TextHash      string       `gorm:"type:varchar(255)" json:"text_hash"`
	AttachName    string       `gorm:"type:varchar(255)" json:"attach_name"`
	AttachHash    string       `gorm:"type:varchar(255)" json:"attach_hash"`
	AttachMd5     string       `gorm:"type:varchar(255)" json:"attach_md5"`
	AttachContent string       `gorm:"type:varchar(255)" json:"attach_content"`
	URL           string       `gorm:"type:varchar(255)" json:"url"`
	Assembly      string       `gorm:"type:varchar(4098)" json:"assembly"`
	AccountID     int          `gorm:"type:int(11);default(0)" json:"account_id"`
	Status        uint8        `gorm:"type:int(8);default(0)" json:"status"` //0=inactive;1=active;2=deleted
	CreateTime    time.Time    `json:"create_time"`
	UpdateTime    time.Time    `json:"update_time"`
	DeleteTime    time.Time    `json:"delete_time"`
}

type FilterLog struct {
	ID         int64        `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	QueueID    string       `gorm:"type:varchar(255)" json:"queue_id"`
	ClientIP   string       `gorm:"type:varchar(64)" json:"client_ip"`
	Sender     string       `gorm:"type:varchar(255)" json:"sender"`
	Nick       string       `gorm:"type:varchar(255)" json:"nick"`
	Rcpt       string       `gorm:"type:varchar(1024)" json:"rcpt"`
	Size       string       `gorm:"type:varchar(255)" json:"size"`
	Mailer     string       `gorm:"type:varchar(255)" json:"mailer"`
	Subject    string       `gorm:"type:varchar(255)" json:"subject"`
	Feature    string       `gorm:"type:mediumtext" json:"feature"`
	Action     FilterAction `gorm:"type:int(8);default(0)" json:"action"`
	CreateTime time.Time
}

type FilterField struct {
	ID          int64       `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Name        string      `gorm:"type:varchar(255)" json:"name"`
	Description string      `gorm:"type:varchar(255)" json:"description"`
	AccountID   int         `gorm:"type:int(11);default(0)" json:"account_id"`
	CreateTime  time.Time   `json:"create_time"`
	UpdateTime  time.Time   `json:"update_time"`
	DeleteTime  time.Time   `json:"delete_time"`
	Status      uint8       `gorm:"type:int(8);default(0)" json:"status"` //0=inactive;1=active;2=deleted
	CanMetric   bool        `gorm:"type:tinyint(1);default(0)" json:"can_metric"`
	Stage       FilterStage `gorm:"type:int(8);default(0)" json:"stage"`
}

type MetricOperation uint8

const (
	MetricOperationCount MetricOperation = iota
	MetricOperationCollect
)

type MetricUnit uint8

const (
	MetricUnitMinute MetricUnit = iota
	MetricUnitHour
	MetricUnitDay
	MetricUnitWeek
	MetricUnitMonth
	MetricUnitYear
)

type FilterMetric struct {
	ID               int64     `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Name             string    `gorm:"type:varchar(255);index:idx_name,unique" json:"name"` // feature name used in filter rule
	Description      string    `gorm:"type:varchar(255)" json:"description"`
	AccountID        int64     `gorm:"type:int(11);default(0)" json:"account_id"`
	CreateTime       time.Time `json:"createTime"`
	UpdateTime       time.Time `json:"updateTime"`
	DeleteTime       time.Time `json:"deleteTime"`
	Status           uint8     `gorm:"type:int(8);default(0)" json:"status"` //0=inactive;1=active;2=deleted
	PrimaryField     FilterField
	PrimaryFieldID   int64 `json:"primary_field_id"`
	SecondaryField   FilterField
	SecondaryFieldID int64           `json:"secondary_field_id"`
	Operation        MetricOperation `gorm:"type:int(8);default(0)" json:"operation"`
	Category         FilterCategory  `gorm:"type:int(8);default(-1)" json:"category"` // -1 means ignore category
	Unit             MetricUnit      `gorm:"type:int(8);default(0)" json:"unit"`      // time unit
	Interval         int             `gorm:"type:int(11);default(0)" json:"interval"` // time interval
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
	case "hasPrefix":
		return fmt.Sprintf(".HasPrefix(\"%s\")", d[1]), nil
	case "hasSuffix":
		return fmt.Sprintf(".HasSuffix(\"%s\")", d[1]), nil
	case "equals":
		return fmt.Sprintf("==\"%s\"", d[1]), nil
	case "notEquals":
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
	case "notEqual":
		return fmt.Sprintf("!=%s", d[1]), nil

	}
	return "", errors.New("invalid rule define")
}

func (r FilterRule) Convert2DRL() (drl string, err error) {
	sb := strings.Builder{}
	sb.WriteString(
		fmt.Sprintf("rule rule_%d \"%s\" salience %d {\n",
			r.ID,
			strings.Replace(r.Description, "\"", "", -1),
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
	sb.WriteString(fmt.Sprintf("\t\tresult.RuleID=%d;\n", r.ID))
	sb.WriteString(fmt.Sprintf("\t\tresult.Action=%d;\n", r.Action))
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

func GetFilterField(orderField, orderDir string, page, pageSize int) (total int64, fields []FilterField, err error) {
	query := db.Model(&fields)
	query.Count(&total)

	if orderField != "" && orderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", orderField, orderDir))
	}
	query = query.Offset(page).Limit(pageSize)
	err = query.Find(&fields).Error
	if err != nil {
		return 0, nil, err
	}
	return
}

func GetFilterMetricByStage(stage FilterStage) (metrics []FilterMetric, err error) {
	metrics = make([]FilterMetric, 0)

	// query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("filter_metric_%d", stage)

	ctx := context.Background()
	cache, err := rdb.Get(ctx, key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(cache), &metrics)
		return metrics, err
	}

	query := db.Model(&metrics).Preload("PrimaryField").Preload("SecondaryField").Where("filter_metrics.status=?", 1)
	query = query.Joins("left join filter_fields on (filter_fields.id = filter_metrics.primary_field_id OR filter_fields.id = filter_metrics.secondary_field_id)").
		Where("filter_fields.stage=?", stage).
		Where("filter_fields.status=?", 1).
		Where("filter_fields.can_metric", 1)
	err = query.Find(&metrics).Error
	if err != nil {
		return nil, err
	}
	// cache
	if d, err := json.Marshal(metrics); err == nil {
		_ = rdb.Set(ctx, key, d, time.Minute*5).Err()
	}
	return
}

func GetFilterMetric(req IndexFilterMetricRequest) (total int64, data []IndexFilterMetricResponse, err error) {
	metrics := make([]FilterMetric, 0)
	query := db.Model(&metrics)
	query.Count(&total)

	if req.OrderField != "" && req.OrderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", req.OrderField, req.OrderDir))
	}
	query = query.Offset(req.Start).Limit(req.Length)
	err = query.Find(&metrics).Error
	if err != nil {
		return 0, nil, err
	}

	// query all fields
	fields, err := GetAllFilterField()
	if err != nil {
		return 0, nil, err
	}
	fieldMap := make(map[int64]string)
	for _, f := range fields {
		fieldMap[f.ID] = f.Name
	}
	// 将metrics转换为IndexMetricResponse
	for _, m := range metrics {
		o := IndexFilterMetricResponse{}
		o.ID = m.ID
		o.Name = m.Name
		o.Description = m.Description
		o.AccountID = m.AccountID
		o.CreateTime = m.CreateTime
		o.UpdateTime = m.UpdateTime
		o.Status = m.Status
		o.PrimaryFieldID = m.PrimaryFieldID
		o.SecondaryFieldID = m.SecondaryFieldID
		o.Operation = m.Operation
		o.Category = m.Category
		o.Unit = m.Unit
		o.Interval = m.Interval
		o.PrimaryFieldName = fieldMap[o.PrimaryFieldID]
		o.SecondaryFieldName = fieldMap[o.SecondaryFieldID]
		data = append(data, o)
	}
	return
}

func GetFilterMetricByID(id int64) (metric FilterMetric, err error) {
	err = db.Model(&metric).Where("id=?", id).First(&metric).Error
	return
}

func GetAllFilterField() (fields []FilterField, err error) {
	err = db.Model(&fields).Where("status=?", 1).Where("can_metric=?", 1).Order("name asc").Find(&fields).Error
	return
}

func CreateFilterMetric(req CreateFilterMetricRequest) error {
	var metric FilterMetric
	// req.ID>0 means update
	if req.ID > 0 {
		err := db.Model(&metric).Where("id=?", req.ID).First(&metric).Error
		if err != nil && metric.ID == 0 {
			return errors.New("metric not exists")
		}
		metric.Name = req.Name
		metric.Description = req.Description
		metric.AccountID = req.AccountID
		metric.UpdateTime = time.Now()
		metric.PrimaryFieldID = req.PrimaryField
		metric.SecondaryFieldID = req.SecondaryField
		metric.Operation = req.Operation
		metric.Category = req.Category
		metric.Unit = req.Unit
		metric.Interval = req.Interval
		return db.Model(&metric).Where("id=?", req.ID).Updates(metric).Error
	}

	err := db.Model(&metric).Where("name=?", req.Name).First(&metric).Error
	if err == nil && metric.ID > 0 {
		return errors.New("metric name already exists")
	}

	refreshFilterMetricCache()

	metric.Name = req.Name
	metric.Description = req.Description
	metric.AccountID = req.AccountID
	metric.CreateTime = time.Now()
	metric.UpdateTime = time.Now()
	metric.Status = 0
	metric.PrimaryFieldID = req.PrimaryField
	metric.SecondaryFieldID = req.SecondaryField
	metric.Operation = req.Operation
	metric.Category = req.Category
	metric.Unit = req.Unit
	metric.Interval = req.Interval

	return db.Model(&metric).Create(&metric).Error
}

func ToggleFilterMetric(id int64) error {
	var metric FilterMetric
	err := db.Model(&metric).Where("id=?", id).First(&metric).Error
	if err != nil && metric.ID == 0 {
		return errors.New("metric not exists")
	}
	if metric.Status == 0 {
		metric.Status = 1
	} else if metric.Status == 1 {
		metric.Status = 0
	} else {
		return errors.New("metric status error")
	}
	refreshFilterMetricCache()
	return db.Model(&metric).Where("id=?", id).Updates(map[string]interface{}{
		"status":      metric.Status,
		"update_time": time.Now(),
	}).Error
}

func DeleteFilterMetric(id int64) error {
	var metric FilterMetric
	err := db.Model(&metric).Where("id=?", id).First(&metric).Error
	if err != nil && metric.ID == 0 {
		return errors.New("metric not exists")
	}
	refreshFilterMetricCache()

	return db.Model(&metric).Where("id=?", id).Updates(FilterMetric{
		Status:     2,
		DeleteTime: time.Now(),
	}).Error
}

func refreshFilterMetricCache() {
	rdb := database.GetRedisClient()
	// range FilterStage
	for i := 0; i <= 5; i++ {
		key := fmt.Sprintf("filter_metric_%d", i)
		rdb.Del(context.Background(), key)
	}
}

func EditFilterMetric(req CreateFilterMetricRequest) error {
	var metric FilterMetric
	err := db.Model(&metric).Where("id=?", req.ID).First(&metric).Error
	if err != nil && metric.ID == 0 {
		return errors.New("metric not exists")
	}

	refreshFilterMetricCache()

	metric.Name = req.Name
	metric.Description = req.Description
	metric.PrimaryFieldID = req.PrimaryField
	metric.SecondaryFieldID = req.SecondaryField
	metric.Operation = req.Operation
	metric.Category = req.Category
	metric.Unit = req.Unit
	metric.Interval = req.Interval
	metric.UpdateTime = time.Now()
	return db.Save(&metric).Error

}

func GetFilterFieldByStage(stage FilterStage) ([]FilterField, error) {
	var fields []FilterField

	// query cache first
	rdb := database.GetRedisClient()
	key := fmt.Sprintf("filter_field_%d", stage)

	ctx := context.Background()
	cache, err := rdb.Get(ctx, key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(cache), &fields)
		return fields, err
	}

	err = db.Model(&fields).Where("status=? AND stage=?", 1, stage).Order("name asc").Find(&fields).Error
	if err == nil {
		// cache
		cache, err := json.Marshal(fields)
		if err == nil {
			rdb.Set(ctx, key, cache, time.Duration(1)*time.Hour)
		}
	}
	return fields, err
}

func truncateTimeByMinute(t time.Time, threshold int) string {
	minutes := t.Minute()
	remainder := minutes % threshold
	if remainder != 0 {
		t = t.Add(-time.Duration(remainder) * time.Minute).Add(time.Duration(threshold) * time.Minute)
	}
	return t.Truncate(time.Duration(threshold) * time.Minute).Format("200601021504")
}

func truncateTimeByHour(t time.Time, threshold int) string {
	hours := t.Hour()
	remainder := hours % threshold
	if remainder != 0 {
		t = t.Add(-time.Duration(remainder) * time.Hour).Add(time.Duration(threshold) * time.Hour)
	}
	return t.Truncate(time.Duration(threshold) * time.Hour).Format("2006010215")
}

func truncateTimeByDay(t time.Time, threshold int) string {
	days := t.Day()
	remainder := days % threshold
	if remainder != 0 {
		t = t.AddDate(0, 0, -remainder).AddDate(0, 0, threshold)
	}
	return t.Truncate(time.Duration(threshold) * time.Hour * 24).Format("20060102")
}

func truncateTimeByWeek(t time.Time, threshold int) string {
	_, weeks := t.ISOWeek()
	remainder := weeks % threshold
	if remainder != 0 {
		t = t.AddDate(0, 0, -remainder*7).AddDate(0, 0, threshold*7)

	}
	nt := t.Truncate(time.Duration(threshold) * time.Hour * 24 * 7)
	year, week := nt.ISOWeek()
	return fmt.Sprintf("%d%02d", year, week)
}

func truncateTimeByMonth(t time.Time, threshold int) string {
	months := int(t.Month())
	remainder := months % threshold
	if remainder != 0 {
		t = t.AddDate(0, -remainder, 0).AddDate(0, threshold, 0)
	}
	return t.Truncate(time.Duration(threshold) * time.Hour * 24).Format("200601")
}

func truncateTimeByYear(t time.Time, threshold int) string {
	years := int(t.Month())
	remainder := years % threshold
	if remainder != 0 {
		t = t.AddDate(-remainder, 0, 0).AddDate(threshold, 0, 0)
	}
	return t.Truncate(time.Duration(threshold) * time.Hour * 24).Format("2006")
}

func (m FilterMetric) MakeFilterMetricKey() string {
	switch m.Unit {
	case MetricUnitMinute:
		return fmt.Sprintf("%s:%s", m.Name, truncateTimeByMinute(time.Now(), m.Interval))
	case MetricUnitHour:
		return fmt.Sprintf("%s:%s", m.Name, truncateTimeByHour(time.Now(), m.Interval))
	case MetricUnitDay:
		return fmt.Sprintf("%s:%s", m.Name, truncateTimeByDay(time.Now(), m.Interval))
	case MetricUnitWeek:
		return fmt.Sprintf("%s:%s", m.Name, truncateTimeByWeek(time.Now(), m.Interval))
	case MetricUnitMonth:
		return fmt.Sprintf("%s:%s", m.Name, truncateTimeByMonth(time.Now(), m.Interval))
	case MetricUnitYear:
		return fmt.Sprintf("%s:%s", m.Name, truncateTimeByYear(time.Now(), m.Interval))
	}
	return ""
}

func (m FilterMetric) MakeFilterMetricTimeout() time.Duration {
	switch m.Unit {
	case MetricUnitMinute:
		return time.Duration(m.Interval) * time.Minute
	case MetricUnitHour:
		return time.Duration(m.Interval) * time.Hour
	case MetricUnitDay:
		return time.Duration(m.Interval) * 24 * time.Hour
	case MetricUnitWeek:
		return time.Duration(m.Interval) * 7 * 24 * time.Hour
	case MetricUnitMonth:
		return time.Duration(m.Interval) * 30 * 24 * time.Hour
	case MetricUnitYear:
		return time.Duration(m.Interval) * 365 * 24 * time.Hour
	}
	return time.Duration(5) * time.Minute
}
