package requests

import (
	"context"
	"io"
	"net/http"
	"net/url"
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
	if timeout <= 0 {
		return &http.Client{}
	}
	return &http.Client{Timeout: timeout}
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
	return &Response{
		StatusCode: res.StatusCode,
		Body:       b,
		JSON:       gjson.ParseBytes(b),
		Headers:    res.Header,
		URL:        u,
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
