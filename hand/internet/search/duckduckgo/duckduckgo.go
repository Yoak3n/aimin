package duckduckgo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/hand/pkg/requests"
	"github.com/tidwall/gjson"
)

const InstantAnswerURL = "https://api.duckduckgo.com/"

type Options struct {
	NoRedirect bool
	NoHTML     *bool
	SafeSearch int
	Timeout    time.Duration
	UserAgent  string
}

type ResultItem struct {
	Text     string
	FirstURL string
}

type SearchResult struct {
	Heading       string
	AbstractText  string
	AbstractURL   string
	Answer        string
	AnswerType    string
	Definition    string
	DefinitionURL string
	Results       []ResultItem
	RelatedTopics []ResultItem
	Raw           gjson.Result
}

func Search(ctx context.Context, query string, opt *Options) (*SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if opt == nil {
		opt = &Options{}
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 10 * time.Second
	}

	q := map[string]string{
		"q":      query,
		"format": "json",
	}
	if opt.NoRedirect {
		q["no_redirect"] = "1"
	}
	noHTML := true
	if opt.NoHTML != nil {
		noHTML = *opt.NoHTML
	}
	if noHTML {
		q["no_html"] = "1"
	}
	if opt.SafeSearch != 0 {
		q["safe_search"] = fmt.Sprintf("%d", opt.SafeSearch)
	}

	headers := map[string]string{}
	if strings.TrimSpace(opt.UserAgent) != "" {
		headers["User-Agent"] = opt.UserAgent
	}

	resp, err := requests.Get(ctx, InstantAnswerURL, &requests.Options{
		Query:   q,
		Headers: headers,
		Timeout: opt.Timeout,
	})
	if err != nil {
		return nil, err
	}

	raw := resp.JSON
	out := &SearchResult{
		Heading:       raw.Get("Heading").String(),
		AbstractText:  raw.Get("AbstractText").String(),
		AbstractURL:   raw.Get("AbstractURL").String(),
		Answer:        raw.Get("Answer").String(),
		AnswerType:    raw.Get("AnswerType").String(),
		Definition:    raw.Get("Definition").String(),
		DefinitionURL: raw.Get("DefinitionURL").String(),
		Results:       parseItems(raw.Get("Results")),
		RelatedTopics: parseRelatedTopics(raw.Get("RelatedTopics")),
		Raw:           raw,
	}
	return out, nil
}

func parseItems(arr gjson.Result) []ResultItem {
	if !arr.Exists() || !arr.IsArray() {
		return nil
	}
	out := make([]ResultItem, 0, len(arr.Array()))
	for _, it := range arr.Array() {
		text := strings.TrimSpace(it.Get("Text").String())
		url := strings.TrimSpace(it.Get("FirstURL").String())
		if text == "" && url == "" {
			continue
		}
		out = append(out, ResultItem{Text: text, FirstURL: url})
	}
	return out
}

func parseRelatedTopics(rt gjson.Result) []ResultItem {
	if !rt.Exists() || !rt.IsArray() {
		return nil
	}
	out := make([]ResultItem, 0, len(rt.Array()))
	for _, it := range rt.Array() {
		if it.Get("Topics").IsArray() {
			out = append(out, parseItems(it.Get("Topics"))...)
			continue
		}
		text := strings.TrimSpace(it.Get("Text").String())
		url := strings.TrimSpace(it.Get("FirstURL").String())
		if text == "" && url == "" {
			continue
		}
		out = append(out, ResultItem{Text: text, FirstURL: url})
	}
	return out
}
