package controller

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"
)

func GetAccountID(c *gin.Context) (int64, error) {
	sess := sessions.Default(c)
	userIDS := sess.Get("userID")
	return strconv.ParseInt(userIDS.(string), 10, 64)
}

func GetMailbox(c *gin.Context) string {
	sess := sessions.Default(c)
	mailbox := sess.Get("mailbox")
	return mailbox.(string)
}

func CamelFieldName(s string) string {
	buf := strings.Builder{}
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			buf.WriteByte('_')
			buf.WriteRune(c + 32)
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

func FormatNumber(num float64, digits int) string {
	si := []struct {
		value  float64
		symbol string
	}{
		{value: 1, symbol: ""},
		{value: 1e3, symbol: "K"},
		{value: 1e6, symbol: "M"},
		{value: 1e9, symbol: "G"},
	}

	rx := regexp.MustCompile(`\.0+$|(\.[0-9]*[1-9])0+$`)

	var i int
	for i = len(si) - 1; i > 0; i-- {
		if num >= si[i].value {
			break
		}
	}

	// Format the number with the specified number of decimal digits
	formattedNum := strconv.FormatFloat(num/si[i].value, 'f', digits, 64)

	// Remove trailing zeros
	formattedNum = rx.ReplaceAllString(formattedNum, "${1}")

	// If the number is an integer and there is no symbol, remove the decimal point
	if formattedNum[len(formattedNum)-1] == '0' && strings.Index(formattedNum, ".") != -1 && si[i].symbol == "" {
		formattedNum = formattedNum[:len(formattedNum)-1]
	}

	return formattedNum + si[i].symbol
}

func SendMailOpenRelay(host string, sender string, receipts []string, subject string, body []byte) error {
	client, err := smtp.Dial(host + ":25")
	if err != nil {
		return err
	}
	defer client.Close()
	domain := ""
	if d := strings.SplitN(sender, "@", 2); len(d) == 2 {
		domain = d[1]
	}
	if err := client.Hello(domain); err != nil {
		return err
	}

	if err := client.Mail(sender); err != nil {
		return err
	}

	for _, receipt := range receipts {
		if err := client.Rcpt(receipt); err != nil {
			log.Println(err)
		}
	}

	// Send the email body.
	wc, err := client.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(wc, string(body))
	if err != nil {
		return err
	}

	err = wc.Close()
	if err != nil {
		resp := err.Error()
		// send mail successfully
		if strings.HasPrefix(resp, "2") {
			err = nil
		}
	}
	_ = client.Quit()
	return err
}
