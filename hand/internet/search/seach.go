package search

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/hand/internet/fetch"
	"github.com/Yoak3n/aimin/hand/internet/search/bilibili"
	"github.com/Yoak3n/aimin/hand/internet/search/duckduckgo"
)

type Provider string

const (
	ProviderDuckDuckGo Provider = "duckduckgo"
	ProviderBilibili   Provider = "bilibili"
	ProviderBingSERP   Provider = "bing_serp"
)

type Result struct {
	Provider Provider
	Title    string
	URL      string
	Snippet  string
}

type Options struct {
	Limit   int
	Timeout time.Duration

	BilibiliCookie string
	BilibiliUA     string

	PreferBrowser bool
}

const defaultUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) " +
	"Chrome/123.0.0.0 Safari/537.36"

func Search(ctx context.Context, query string, opt *Options) ([]Result, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("query is required")
	}
	cfg := Options{
		Limit:         10,
		Timeout:       15 * time.Second,
		PreferBrowser: true,
	}
	if opt != nil {
		if opt.Limit > 0 {
			cfg.Limit = opt.Limit
		}
		if opt.Timeout > 0 {
			cfg.Timeout = opt.Timeout
		}
		cfg.BilibiliCookie = strings.TrimSpace(opt.BilibiliCookie)
		cfg.BilibiliUA = strings.TrimSpace(opt.BilibiliUA)
		cfg.PreferBrowser = opt.PreferBrowser
	}
	if cfg.Limit > 20 {
		cfg.Limit = 20
	}
	if cfg.BilibiliCookie != "" && cfg.BilibiliUA == "" {
		cfg.BilibiliUA = defaultUA
	}

	cctx := ctx
	if cctx == nil {
		cctx = context.Background()
	}
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		cctx, cancel = context.WithTimeout(cctx, cfg.Timeout)
		defer cancel()
	}

	if out, err := tryDuckDuckGo(cctx, q, cfg.Limit); err == nil && len(out) > 0 {
		return out, nil
	}

	if out, err := tryBingSERP(cctx, q, cfg); err == nil && len(out) > 0 {
		return out, nil
	}

	if strings.TrimSpace(cfg.BilibiliCookie) != "" {
		if out, err := tryBilibili(cctx, q, cfg); err == nil && len(out) > 0 {
			return out, nil
		}
	}

	return nil, fmt.Errorf("no results")
}

func tryDuckDuckGo(ctx context.Context, query string, limit int) ([]Result, error) {
	res, err := duckduckgo.Search(ctx, query, &duckduckgo.Options{Timeout: 12 * time.Second})
	if err != nil {
		return nil, err
	}
	items := make([]duckduckgo.ResultItem, 0, len(res.Results)+len(res.RelatedTopics))
	items = append(items, res.Results...)
	items = append(items, res.RelatedTopics...)
	if len(items) == 0 {
		return nil, fmt.Errorf("empty result")
	}
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]Result, 0, len(items))
	for _, it := range items {
		s := strings.TrimSpace(it.Text)
		u := strings.TrimSpace(it.FirstURL)
		if s == "" && u == "" {
			continue
		}
		title := compactLine(s, 140)
		out = append(out, Result{
			Provider: ProviderDuckDuckGo,
			Title:    title,
			URL:      u,
			Snippet:  s,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty result")
	}
	return out, nil
}

func tryBilibili(ctx context.Context, query string, cfg Options) ([]Result, error) {
	res, err := bilibili.SearchVideos(ctx, bilibili.SearchParams{
		Keyword:   query,
		Page:      1,
		Order:     "totalrank",
		Cookie:    cfg.BilibiliCookie,
		UserAgent: cfg.BilibiliUA,
		Timeout:   minDuration(cfg.Timeout, 12*time.Second),
	})
	if err != nil {
		return nil, err
	}
	if res == nil || len(res.Videos) == 0 {
		return nil, fmt.Errorf("empty result")
	}
	n := cfg.Limit
	if len(res.Videos) < n {
		n = len(res.Videos)
	}
	out := make([]Result, 0, n)
	for i := 0; i < n; i++ {
		v := res.Videos[i]
		title := strings.TrimSpace(v.Title)
		if title == "" {
			continue
		}
		videoURL := ""
		if strings.TrimSpace(v.BVID) != "" {
			videoURL = "https://www.bilibili.com/video/" + strings.TrimSpace(v.BVID)
		}
		snippet := strings.TrimSpace(v.Author)
		if v.Play > 0 {
			if snippet != "" {
				snippet += " "
			}
			snippet += fmt.Sprintf("play=%d", v.Play)
		}
		out = append(out, Result{
			Provider: ProviderBilibili,
			Title:    title,
			URL:      videoURL,
			Snippet:  snippet,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty result")
	}
	return out, nil
}

func tryBingSERP(ctx context.Context, query string, cfg Options) ([]Result, error) {
	rawURL := "https://www.bing.com/search?q=" + url.QueryEscape(strings.TrimSpace(query))
	timeout := minDuration(cfg.Timeout, 18*time.Second)
	jsTimeout := minDuration(cfg.Timeout, 30*time.Second)

	var htmlText string
	var err error
	if cfg.PreferBrowser {
		htmlText, err = fetch.RenderWithChromedp(ctx, rawURL, &fetch.RenderOptions{Timeout: jsTimeout, Scroll: true})
		if err != nil || strings.TrimSpace(htmlText) == "" {
			htmlText = ""
		}
	}
	if htmlText == "" {
		htmlText, _, err = fetch.FetchHTML(ctx, rawURL, &fetch.FetchOptions{Timeout: timeout})
		if err != nil {
			return nil, err
		}
	}

	items := parseBingHTML(htmlText, cfg.Limit)
	if len(items) == 0 {
		return nil, fmt.Errorf("empty result")
	}
	out := make([]Result, 0, len(items))
	for _, it := range items {
		out = append(out, Result{
			Provider: ProviderBingSERP,
			Title:    it.Title,
			URL:      it.URL,
			Snippet:  it.Snippet,
		})
	}
	return out, nil
}

type bingItem struct {
	Title   string
	URL     string
	Snippet string
}

func parseBingHTML(htmlText string, limit int) []bingItem {
	blockRe := regexp.MustCompile(`(?is)<li[^>]*class="b_algo"[^>]*>.*?</li>`)
	blocks := blockRe.FindAllString(htmlText, limit)
	if len(blocks) == 0 {
		return nil
	}

	linkRe := regexp.MustCompile(`(?is)<h2[^>]*>.*?<a[^>]*href="([^"]+)"[^>]*>(.*?)</a>.*?</h2>`)
	snippetRe := regexp.MustCompile(`(?is)<div[^>]*class="b_caption"[^>]*>.*?<p[^>]*>(.*?)</p>`)
	fallbackSnippetRe := regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)

	out := make([]bingItem, 0, len(blocks))
	for _, blk := range blocks {
		m := linkRe.FindStringSubmatch(blk)
		if len(m) < 3 {
			continue
		}
		raw := cleanBingText(html.UnescapeString(m[1]))
		raw = normalizeBingResultURL(raw)
		title := cleanBingText(m[2])
		if title == "" || raw == "" {
			continue
		}
		snippet := ""
		if sm := snippetRe.FindStringSubmatch(blk); len(sm) >= 2 {
			snippet = cleanBingText(sm[1])
		} else if sm := fallbackSnippetRe.FindStringSubmatch(blk); len(sm) >= 2 {
			snippet = cleanBingText(sm[1])
		}
		out = append(out, bingItem{Title: title, URL: raw, Snippet: snippet})
	}
	return out
}

func cleanBingText(s string) string {
	re := regexp.MustCompile(`(?s)<[^>]+>`)
	s = re.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

func normalizeBingResultURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		return raw
	}
	host := strings.ToLower(strings.TrimSpace(u.Host))
	if strings.HasSuffix(host, "bing.com") && strings.HasPrefix(u.Path, "/ck/a") {
		if decoded := decodeBingCkRedirect(u); decoded != "" {
			return decoded
		}
	}
	return raw
}

func decodeBingCkRedirect(u *url.URL) string {
	if u == nil {
		return ""
	}
	enc := strings.TrimPrefix(strings.TrimSpace(u.Query().Get("u")), "a1")
	if enc == "" {
		return ""
	}
	b, err := base64.RawStdEncoding.DecodeString(enc)
	if err != nil || len(b) == 0 {
		return ""
	}
	out := strings.TrimSpace(string(b))
	out = strings.ReplaceAll(out, "\u0000", "")
	out = strings.TrimSpace(out)
	if out == "" {
		return ""
	}
	if _, err := url.Parse(out); err != nil {
		return ""
	}
	return out
}

func compactLine(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if maxRunes > 0 {
		rs := []rune(s)
		if len(rs) > maxRunes {
			s = string(rs[:maxRunes]) + "..."
		}
	}
	return s
}

func minDuration(a time.Duration, b time.Duration) time.Duration {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}
