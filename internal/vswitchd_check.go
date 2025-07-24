package internal

import (
	"os/exec"
	"strings"
)

func IsVswitchdActive() bool {
	out, err := exec.Command("systemctl", "is-active", "vswitchd.service").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}
