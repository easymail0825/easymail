package preprocessing

import "strings"

// GetDomain get domain part from the email address
func GetDomain(email string) string {
	d := strings.SplitN(email, "@", 2)
	if len(d) == 2 {
		return d[1]
	}
	return ""
}
