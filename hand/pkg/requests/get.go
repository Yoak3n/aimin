package requests

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

type Options struct {
	Headers map[string]string
	Query   map[string]string
	Timeout time.Duration
}

type Response struct {
	StatusCode int
	Body       []byte
	JSON       gjson.Result
	Headers    http.Header
	URL        string
}

func client(timeout time.Duration) *http.Client {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = proxyFunc
	if timeout <= 0 {
		return &http.Client{Transport: base}
	}
	return &http.Client{Timeout: timeout, Transport: base}
}

func proxyFunc(req *http.Request) (*url.URL, error) {
	p, err := http.ProxyFromEnvironment(req)
	if p != nil || err != nil {
		return p, err
	}
	spec := strings.TrimSpace(systemProxyString())
	if spec == "" {
		return nil, nil
	}
	return parseProxySpecForRequest(spec, req.URL), nil
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

func applyHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func buildURL(raw string, q map[string]string) (string, error) {
	if len(q) == 0 {
		return raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	qs := u.Query()
	for k, v := range q {
		qs.Set(k, v)
	}
	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func doRequest(ctx context.Context, method, rawURL string, body io.Reader, opt *Options) (*Response, error) {
	if opt == nil {
		opt = &Options{}
	}
	u, err := buildURL(rawURL, opt.Query)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	if opt.Headers != nil {
		applyHeaders(req, opt.Headers)
	}
	c := client(opt.Timeout)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	finalURL := u
	if res.Request != nil && res.Request.URL != nil {
		if t := strings.TrimSpace(res.Request.URL.String()); t != "" {
			finalURL = t
		}
	}
	return &Response{
		StatusCode: res.StatusCode,
		Body:       b,
		JSON:       gjson.ParseBytes(b),
		Headers:    res.Header,
		URL:        finalURL,
	}, nil
}

func Get(ctx context.Context, rawURL string, opt *Options) (*Response, error) {
	if opt == nil {
		opt = &Options{}
	}
	if opt.Headers == nil {
		opt.Headers = map[string]string{}
	}
	if _, ok := opt.Headers["Accept"]; !ok {
		opt.Headers["Accept"] = "application/json, text/plain;q=0.9, */*;q=0.8"
	}
	return doRequest(ctx, http.MethodGet, rawURL, nil, opt)
}
