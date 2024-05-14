package maillog

import (
	"bufio"
	"strings"
	"testing"
)

func TestMigrate(t *testing.T) {
	db.AutoMigrate(&MailLog{})
}

func TestInsertFromString(t *testing.T) {
	s := `Apr 21 21:00:22 racknerd-6a8470 postfix/smtpd[18851]: connect from out203-205-221-153.mail.qq.com[203.205.221.153]
Apr 21 21:00:24 racknerd-6a8470 postfix/smtpd[18851]: 12087205C7: client=out203-205-221-153.mail.qq.com[203.205.221.153]
Apr 21 21:00:24 racknerd-6a8470 postfix/cleanup[18856]: 12087205C7: message-id=<tencent_C5BF6673393163ED4E9C2F90B6A007580705@qq.com>
Apr 21 21:00:24 racknerd-6a8470 postfix/qmgr[18828]: 12087205C7: from=<easymail0825@qq.com>, size=3114, nrcpt=1 (queue active)
Apr 21 21:00:24 racknerd-6a8470 postfix/lmtp[18857]: 12087205C7: to=<admin@easypostix.com>, relay=127.0.0.1[127.0.0.1]:10025, delay=0.92, delays=0.9/0/0/0.02, dsn=2.0.0, status=sent (250 <admin@easypostix.com> mail ok)
Apr 21 21:00:24 racknerd-6a8470 postfix/qmgr[18828]: 12087205C7: removed
Apr 21 21:00:24 racknerd-6a8470 postfix/smtpd[18851]: disconnect from out203-205-221-153.mail.qq.com[203.205.221.153] ehlo=2 starttls=1 mail=1 rcpt=1 data=1 quit=1 commands=7
Apr 21 21:03:45 racknerd-6a8470 postfix/anvil[18854]: statistics: max connection rate 1/60s for (smtp:107.173.114.206) at Apr 21 20:59:13
Apr 21 21:03:45 racknerd-6a8470 postfix/anvil[18854]: statistics: max connection count 1 for (smtp:107.173.114.206) at Apr 21 20:59:13
Apr 21 21:03:45 racknerd-6a8470 postfix/anvil[18854]: statistics: max cache size 1 at Apr 21 20:59:13`

	scan := bufio.NewScanner(strings.NewReader(s))
	for scan.Scan() {
		line := scan.Text()
		mailLog, err := Parse(line)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		err = Save(mailLog)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}
}
