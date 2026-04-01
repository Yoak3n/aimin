//go:build !windows

package requests

func systemProxyString() string {
	return ""
}
