package sync

import (
	"easymail/internal/model"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
)

/*
SynchronizeVirtualDomain
@Description: tell postfix what virtual domains to apply
@Date: 2024/4/23 11:04 AM
*/
func SynchronizeVirtualDomain() error {
	c, err := model.GetConfigureByNames("postfix", "configure", "virtual_mailbox_domains")
	if err != nil {
		return err
	}
	if c.DataType != model.DataTypeString {
		return errors.New("wrong data type")
	}
	dist := c.Value

	domains, err := model.FindAllValidateDomain()
	if err != nil {
		return err
	}
	fd, err := os.OpenFile(dist, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func(fd *os.File) {
		err := fd.Close()
		if err != nil {
			log.Println(err)
		}
	}(fd)
	for _, domain := range domains {
		_, err2 := fd.WriteString(fmt.Sprintf("%s OK\n", domain.Name))
		if err2 != nil {
			return err2
		}
	}

	// run postmap to make hash
	c, err = model.GetConfigureByNames("postfix", "execute", "postmap")
	if err != nil {
		return err
	}
	postmap := c.Value
	cmd := exec.Command("sudo", postmap, dist)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("postmap run failed with %s\n", err)
		return err
	}
	log.Printf("postmap run result:\n%s\n", string(out))

	// last, should reload postfix
	c, err = model.GetConfigureByNames("postfix", "execute", "postfix")
	if err != nil {
		return err
	}

	cmd = exec.Command("sudo", c.Value, "reload")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	log.Printf("postfix reload result:\n%s\n", string(out))
	return nil
}
