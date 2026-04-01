//go:build windows

package fetch

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

func systemProxyForURL(rawURL string) string {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()

	enabled, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil || enabled == 0 {
		return ""
	}
	ps, _, err := k.GetStringValue("ProxyServer")
	if err != nil {
		return ""
	}
	ps = strings.TrimSpace(ps)
	if ps == "" {
		return ""
	}
	return normalizeProxyServerForChrome(ps)
}
