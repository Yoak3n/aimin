//go:build !windows

package fetch

func systemProxyForURL(rawURL string) string {
	return ""
}
