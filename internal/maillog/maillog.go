package maillog

import (
	_ "easymail/internal/database"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var TopFilterRegexp = regexp.MustCompile("^([A-Z][a-z]{2} \\d{1,2} \\d{2}:\\d{2}:\\d{2}) (\\w+) (postfix/\\w+\\[\\d+\\]): (.*)")
var HostRegexp = regexp.MustCompile("(\\S+)\\[(\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})\\]")
var timeFormat = "Jan 02 15:04:05 2006"

type MailLog struct {
	ID         int64 `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	CreateTime time.Time
	LogTime    time.Time `gorm:"index:idx_log_time" json:"logTime"`
	State      string    `gorm:"type:varchar(36)" json:"state"`
	QueueID    string    `gorm:"type:varchar(36);index:idx_queue" json:"queueID"`
	Hostname   string    `gorm:"type:varchar(64);index:idx_host" json:"hostname"`
	IP         string    `gorm:"type:varchar(64);index:idx_ip" json:"ip"`
	From       string    `gorm:"type:varchar(255);index:idx_from" json:"from"`
	To         string    `gorm:"type:varchar(255);index:idx_to" json:"to"`
	SessionID  string    `gorm:"type:varchar(64)" json:"sessionID"`
	Process    string    `gorm:"type:varchar(64)" json:"process"`
	Message    string    `gorm:"type:text" json:"message"`
}

// FormatMailbox
// remove prefix character < and suffix character > from mailbox
func FormatMailbox(m string) string {
	// make string buffer
	var sb = &strings.Builder{}
	bs := []byte(m)
	for _, c := range bs {
		if c == '<' || c == '>' {
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

func Parse(line string) (mailLog *MailLog, err error) {
	mailLog = &MailLog{}
	found := TopFilterRegexp.FindStringSubmatch(line)
	if len(found) != 5 {
		return nil, errors.New("not enough fields")
	}

	now := time.Now()
	logTime, err := time.Parse(timeFormat, found[1]+" "+now.Format("2006"))
	if err != nil {
		return nil, err
	}
	mailLog.LogTime = logTime
	mailLog.CreateTime = now
	mailLog.SessionID = found[2]
	mailLog.Process = found[3]
	mailLog.Message = found[4]

	if strings.HasPrefix(mailLog.Process, "postfix/anvil") {
		mailLog.State = "statistics"
	} else if strings.HasPrefix(mailLog.Message, "connect from ") {
		mailLog.State = "connect"
		// parse ip and domain
		hosts := HostRegexp.FindStringSubmatch(mailLog.Message)
		if len(hosts) == 3 {
			mailLog.Hostname = hosts[1]
			mailLog.IP = hosts[2]
		}
	} else if strings.HasPrefix(mailLog.Message, "disconnect from ") {
		mailLog.State = "disconnect"
		hosts := HostRegexp.FindStringSubmatch(mailLog.Message)
		if len(hosts) == 3 {
			mailLog.Hostname = hosts[1]
			mailLog.IP = hosts[2]
		}
	} else if strings.Index(mailLog.Message, ": ") > 0 && strings.Index(mailLog.Message, ": ") < 16 {
		parts := strings.SplitN(mailLog.Message, ": ", 2)
		if len(parts) == 2 {
			mailLog.QueueID = parts[0]
		}
		for _, p := range strings.Split(parts[1], ", ") {
			if strings.Index(p, "=") > 0 {
				d := strings.SplitN(p, "=", 2)
				if len(d) == 2 {
					if d[0] == "from" {
						mailLog.From = FormatMailbox(d[1])
					}
					if d[0] == "to" {
						mailLog.To = FormatMailbox(d[1])
					}
					if d[0] == "status" {
						x := strings.SplitN(d[1], " ", 2)
						if len(x) == 2 {
							mailLog.State = x[0]
						}
					}
				}
			}
		}
	}
	return mailLog, nil
}

func Save(mailLog *MailLog) error {
	return db.Model(mailLog).Create(mailLog).Error
}

func Index(startTime, endTime time.Time, searchField int, keyword, orderField, orderDir string, page, pageSize int) (int64, []MailLog, error) {
	logs := make([]MailLog, 0)
	query := db.Model(&logs)

	if !startTime.IsZero() {
		query = query.Where("log_time >=?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("log_time <=?", endTime.Add(24*time.Hour))
	}
	if keyword != "" {
		switch searchField {
		case 0:
			query = query.Where("queue_id =?", keyword)
			break
		case 1:
			query = query.Where("hostname like ?", "%"+keyword+"%")
			break
		case 2:
			query = query.Where("ip like ?", "%"+keyword+"%")
		case 3:
			query = query.Where("`from` like ?", "%"+keyword+"%")
		case 4:
			query = query.Where("`to` like ?", "%"+keyword+"%")
		}
	}

	var total int64
	query.Count(&total)

	if orderField != "" && orderDir != "" {
		query = query.Order(fmt.Sprintf("%s %s", orderField, orderDir))
	}
	query = query.Offset(page).Limit(pageSize)
	err := query.Find(&logs).Error
	if err != nil {
		return 0, nil, err
	}
	return total, logs, nil
}
