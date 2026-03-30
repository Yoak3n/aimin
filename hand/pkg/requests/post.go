package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func Post(ctx context.Context, rawURL string, body []byte, contentType string, opt *Options) (*Response, error) {
	if opt == nil {
		opt = &Options{}
	}
	if opt.Headers == nil {
		opt.Headers = map[string]string{}
	}
	if _, ok := opt.Headers["Accept"]; !ok {
		opt.Headers["Accept"] = "application/json, text/plain;q=0.9, */*;q=0.8"
	}
	if contentType != "" {
		opt.Headers["Content-Type"] = contentType
	}
	return doRequest(ctx, http.MethodPost, rawURL, bytes.NewReader(body), opt)
}

func PostJSON(ctx context.Context, rawURL string, payload any, opt *Options) (*Response, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return Post(ctx, rawURL, b, "application/json; charset=utf-8", opt)
}

func PostForm(ctx context.Context, rawURL string, values url.Values, opt *Options) (*Response, error) {
	b := []byte(values.Encode())
	return Post(ctx, rawURL, b, "application/x-www-form-urlencoded", opt)
}
