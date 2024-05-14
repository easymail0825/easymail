package queue

import (
	"fmt"
	"log"
	"os/exec"
	"testing"
)

func TestDump(t *testing.T) {
	/*
		out = []byte(`{"queue_name": "active", "queue_id": "09EAF4422BA", "arrival_time": 1713849287, "message_size": 688, "forced_expire": false, "sender": "admin@super.com", "recipients": [{"address": "admin@super.com"}]}
			{"queue_name": "active", "queue_id": "BC2624422BB", "arrival_time": 1713849287, "message_size": 688, "forced_expire": false, "sender": "admin@super.com", "recipients": [{"address": "admin@super.com"}]}
			{"queue_name": "active", "queue_id": "D43AA4422B8", "arrival_time": 1713849285, "message_size": 688, "forced_expire": false, "sender": "admin@super.com", "recipients": [{"address": "admin@super.com"}]}
			{"queue_name": "deferred", "queue_id": "C9AD04011E", "arrival_time": 1713788783, "message_size": 599, "sender": "admin@super.com", "recipients": [{"address": "admin@super.com", "delay_reason":"connect to 192.168.1.106[192.168.1.106]:10025: Connection refused"}]}`)

	*/
	//t.Log(d)
	queues, err := Dump()
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range queues {
		t.Log(q.Name, q.Id)
	}
}

func TestCmd(t *testing.T) {
	cmd := exec.Command("sudo", "/usr/sbin/postcat", "-qh", "581284422BB")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:\n%s\n", string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("combined out:\n%s\n", string(out))

}
