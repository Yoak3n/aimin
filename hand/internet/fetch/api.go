package fetch

import (
	"context"
	"time"
)

func FetchPage(ctx context.Context, rawURL string, opt *ProcessOptions) (*PageResult, error) {
	if opt == nil {
		opt = &ProcessOptions{JSFallback: true}
	}
	return ProcessOne(ctx, rawURL, opt)
}

func FetchPageText(ctx context.Context, rawURL string, opt *ProcessOptions) (string, string, error) {
	if opt == nil {
		opt = &ProcessOptions{JSFallback: true}
	}
	r, err := ProcessOne(ctx, rawURL, opt)
	if err != nil {
		return "", "", err
	}
	return r.Text, r.FinalURL, nil
}

func FetchTextDefault(rawURL string) (string, error) {
	text, _, err := FetchTextDefaultWithFinalURL(rawURL)
	return text, err
}

func FetchTextDefaultWithFinalURL(rawURL string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	r, err := ProcessOne(ctx, rawURL, &ProcessOptions{JSFallback: true})
	if err != nil {
		return "", "", err
	}
	return r.Text, r.FinalURL, nil
}
