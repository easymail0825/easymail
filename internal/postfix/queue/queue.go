package queue

import (
	"easymail/internal/model"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

/*
run postqueue -j to dump queue information
{"queue_name": "deferred", "queue_id": "C9AD04011E", "arrival_time": 1713788783, "message_size": 599,
"sender": "admin@super.com", "recipients": [{"address": "admin@super.com", "delay_reason":
"connect to 192.168.1.106[192.168.1.106]:10025: Connection refused"}]}
*/

type Queue struct {
	Name        string `json:"queue_name"`
	Id          string `json:"queue_id"`
	ArrivalTime int64  `json:"arrival_time"`
	MessageSize int    `json:"message_size"`
	Sender      string `json:"sender"`
	Recipients  []struct {
		Address string `json:"address"`
	} `json:"recipients"`
}

func Dump() ([]Queue, error) {
	queues := make([]Queue, 0)
	c, err := model.GetConfigureByNames("postfix", "execute", "postqueue")
	if err != nil {
		return queues, err
	}
	postqueue := c.Value

	// run postqueue to dump queue detail as json format
	cmd := exec.Command(postqueue, "-j")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return queues, err
	}

	// append , to the end of each line
	str := string(out)
	str = strings.TrimSpace(str)
	bs := strings.Builder{}
	bs.WriteString("[")
	d := strings.Split(str, "\n")
	l := len(d)
	for i, q := range d {
		bs.WriteString(q)
		if i != l-1 {
			bs.WriteString(",\n")
		}
	}
	bs.WriteString("]")
	err = json.Unmarshal([]byte(bs.String()), &queues)
	if err != nil {
		fmt.Println(bs.String())
		return queues, err
	}

	return queues, nil
}

func View(queueID string) (string, error) {
	c, err := model.GetConfigureByNames("postfix", "execute", "postcat")
	if err != nil {
		return "", err
	}
	// need sudo and no password
	var out []byte
	cmd := exec.Command("sudo", c.Value, "-qh", queueID)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Flush(queueID string) (string, error) {
	c, err := model.GetConfigureByNames("postfix", "execute", "postsuper")
	if err != nil {
		return "", err
	}
	// need sudo and no password
	var out []byte
	out, err = exec.Command("sudo", c.Value, "-f", queueID).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Delete(queueID string) (string, error) {
	c, err := model.GetConfigureByNames("postfix", "execute", "postsuper")
	if err != nil {
		return "", err
	}
	// need sudo and no password
	var out []byte
	out, err = exec.Command("sudo", c.Value, "-d", queueID).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
