package collector

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckServiceActive returns an error if the service is not active
func CheckServiceActive(serviceName string) error {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check service status: %v", err)
	}

	status := strings.TrimSpace(string(output))
	if status != "active" {
		return fmt.Errorf("service %s is not active (status: %s)", serviceName, status)
	}

	return nil
}
