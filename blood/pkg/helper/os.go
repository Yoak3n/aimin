package helper

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/host"
)

func GetOSInfo() (string, error) {
	info, err := host.Info()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", info.Platform, info.PlatformVersion), nil
}
