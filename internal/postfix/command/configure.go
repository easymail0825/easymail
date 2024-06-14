package command

import (
	"easymail/internal/model"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

type PostfixConfigure struct {
	Name  string
	Value string
}

func parseAddressAndPort(addrPort string) (string, int, error) {
	parts := strings.SplitN(addrPort, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address and port format: %s", addrPort)
	}

	ip := net.ParseIP(parts[0])
	if ip == nil {
		return "", 0, fmt.Errorf("invalid IP address: %s", parts[0])
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port number: %s", parts[1])
	}
	if port < 0 || port > 65535 {
		return "", 0, fmt.Errorf("port number out of range: %d", port)
	}

	return ip.String(), port, nil
}

// MakeServicePath make service path for postfix configure
func MakeServicePath(family, listen string) (url string, err error) {
	family = strings.ToLower(family)
	if family == "tcp" {
		ip, port, err := parseAddressAndPort(listen)
		if err != nil {
			return "", err
		}

		url = fmt.Sprintf("inet:%s:%d", ip, port)
		return url, nil
	} else if family == "unix" {
		url = fmt.Sprintf("unix:%s", listen)
		return url, nil
	}
	return "", fmt.Errorf("invalid configure, family: %s, listen: %s", family, listen)
}

func FlushPostfixConfig(data []PostfixConfigure) error {
	postConf, err := model.GetConfigureByNames("postfix", "execute", "postconf")
	if err != nil {
		return err
	}

	for _, cfg := range data {
		cmd := exec.Command("sudo", postConf.Value, "-e", cfg.Name+"="+cfg.Value)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("postconf %s=%s failed: %s", cfg.Name, cfg.Value, string(out))
		}
		log.Println("postconf ", cfg.Name, "=", cfg.Value)
	}
	return nil
}
