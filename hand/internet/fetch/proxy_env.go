package fetch

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

func proxyForChromedp(rawURL string) string {
	if p := proxyFromEnv(); p != "" {
		return p
	}
	return systemProxyForURL(rawURL)
}

func proxyForHTTP(req *http.Request) (*url.URL, error) {
	p, err := http.ProxyFromEnvironment(req)
	if p != nil || err != nil {
		return p, err
	}
	spec := strings.TrimSpace(systemProxyForURL(req.URL.String()))
	if spec == "" {
		return nil, nil
	}
	return parseProxySpecForRequest(spec, req.URL), nil
}

func proxyFromEnv() string {
	for _, k := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "ALL_PROXY", "all_proxy"} {
		if v, ok := os.LookupEnv(k); ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return normalizeProxyServerForChrome(v)
			}
		}
	}
	return ""
}

func normalizeProxyServerForChrome(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.Contains(v, "://") {
		return v
	}
	if strings.Contains(v, "=") || strings.Contains(v, ";") {
		return v
	}
	if strings.Contains(v, " ") {
		v = strings.ReplaceAll(v, " ", "")
	}
	if strings.Contains(v, ":") {
		return "http://" + v
	}
	return v
}

func parseProxySpecForRequest(spec string, target *url.URL) *url.URL {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil
	}

	if strings.Contains(spec, "=") || strings.Contains(spec, ";") {
		chosen := chooseProxyByScheme(spec, target)
		return parseProxyURL(chosen)
	}
	return parseProxyURL(spec)
}

func chooseProxyByScheme(spec string, target *url.URL) string {
	want := ""
	if target != nil {
		want = strings.ToLower(strings.TrimSpace(target.Scheme))
	}
	parts := strings.Split(spec, ";")
	fallback := ""
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if strings.Contains(p, "=") {
			kv := strings.SplitN(p, "=", 2)
			k := strings.ToLower(strings.TrimSpace(kv[0]))
			v := ""
			if len(kv) == 2 {
				v = strings.TrimSpace(kv[1])
			}
			if v == "" {
				continue
			}
			if fallback == "" {
				fallback = v
			}
			if want != "" && k == want {
				return v
			}
			continue
		}
		if fallback == "" {
			fallback = p
		}
	}
	return fallback
}

func parseProxyURL(spec string) *url.URL {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil
	}
	if !strings.Contains(spec, "://") {
		spec = "http://" + spec
	}
	u, err := url.Parse(spec)
	if err != nil || u == nil || strings.TrimSpace(u.Host) == "" {
		return nil
	}
	return u
}
