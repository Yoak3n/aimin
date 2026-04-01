package fetch

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-shiori/go-readability"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	"github.com/ledongthuc/pdf"

	"github.com/Yoak3n/aimin/hand/pkg/requests"
)

type FetchOptions struct {
	Timeout        time.Duration
	ConnectTimeout time.Duration
	UserAgent      string
}

type FetchResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
	Headers     http.Header
	FinalURL    string
}

func defaultFetchOptions() FetchOptions {
	return FetchOptions{
		Timeout:        20 * time.Second,
		ConnectTimeout: 10 * time.Second,
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/123.0.0.0 Safari/537.36",
	}
}

func Fetch(ctx context.Context, rawURL string, opt *FetchOptions) (*FetchResult, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("url is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := defaultFetchOptions()
	if opt != nil {
		if opt.Timeout > 0 {
			cfg.Timeout = opt.Timeout
		}
		if opt.ConnectTimeout > 0 {
			cfg.ConnectTimeout = opt.ConnectTimeout
		}
		if strings.TrimSpace(opt.UserAgent) != "" {
			cfg.UserAgent = opt.UserAgent
		}
	}

	resp, err := requests.Get(ctx, rawURL, &requests.Options{
		Headers: map[string]string{
			"User-Agent": cfg.UserAgent,
			"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		},
		Timeout: cfg.Timeout,
	})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	return &FetchResult{
		StatusCode:  resp.StatusCode,
		ContentType: strings.ToLower(resp.Headers.Get("Content-Type")),
		Body:        resp.Body,
		Headers:     resp.Headers.Clone(),
		FinalURL:    strings.TrimSpace(resp.URL),
	}, nil
}

func FetchHTML(ctx context.Context, rawURL string, opt *FetchOptions) (string, string, error) {
	res, err := Fetch(ctx, rawURL, opt)
	if err != nil {
		return "", "", err
	}
	if !looksLikeHTML(res.ContentType, res.FinalURL) {
		return "", "", fmt.Errorf("not html content, content-type=%q", res.ContentType)
	}
	return decodeHTMLBytes(res.Body, res.ContentType), res.FinalURL, nil
}

func ExtractText(htmlText string) (string, error) {
	node, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return "", err
	}

	removeTags(node, map[string]bool{
		"script":   true,
		"style":    true,
		"noscript": true,
		"svg":      true,
		"canvas":   true,
	})
	removeTags(node, map[string]bool{
		"nav":    true,
		"footer": true,
		"header": true,
		"aside":  true,
	})

	main := findFirstElement(node, "article")
	if main == nil {
		main = findFirstElement(node, "main")
	}
	if main == nil {
		main = findFirstElement(node, "body")
	}
	if main == nil {
		main = node
	}

	var lines []string
	collectText(main, &lines)

	text := strings.Join(lines, "\n")
	return postProcessExtractedText(text), nil
}

func ExtractReadabilityText(htmlText string, baseURL string) (string, string, error) {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		base = nil
	}
	article, err := readability.FromReader(strings.NewReader(htmlText), base)
	if err != nil {
		return "", "", err
	}
	title := strings.TrimSpace(article.Title)
	text := strings.TrimSpace(article.TextContent)
	if text == "" {
		text, _ = ExtractText(article.Content)
	}
	text = postProcessExtractedText(text)
	return title, text, nil
}

func ExtractMarkdown(htmlText string, baseURL string) (string, error) {
	node, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return "", err
	}

	removeTags(node, map[string]bool{
		"script":   true,
		"style":    true,
		"noscript": true,
		"svg":      true,
		"canvas":   true,
	})
	removeTags(node, map[string]bool{
		"nav":    true,
		"footer": true,
		"header": true,
		"aside":  true,
	})

	main := findFirstElement(node, "article")
	if main == nil {
		main = findFirstElement(node, "main")
	}
	if main == nil {
		main = findFirstElement(node, "body")
	}
	if main == nil {
		main = node
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, main); err != nil {
		return "", err
	}

	_ = baseURL
	converter := md.NewConverter(md.DomainFromURL(baseURL), true, nil)
	out, err := converter.ConvertString(buf.String())
	if err != nil {
		return "", err
	}
	return postProcessMarkdown(out), nil
}

func SafeFilenameFromURL(rawURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	host := strings.ReplaceAll(u.Host, ":", "_")
	if host == "" {
		host = "unknown_host"
	}

	p := strings.Trim(u.Path, "/")
	if p == "" {
		p = "index"
	}
	p = strings.ReplaceAll(p, "/", "_")
	return fmt.Sprintf("%s__%s.txt", host, p), nil
}

func SaveTextFile(path string, finalURL string, text string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if strings.TrimSpace(finalURL) != "" {
		if _, err := f.WriteString("URL: " + finalURL + "\n\n"); err != nil {
			return err
		}
	}
	if _, err := f.WriteString(text + "\n"); err != nil {
		return err
	}
	return nil
}

func FetchText(ctx context.Context, rawURL string, opt *FetchOptions) (string, string, error) {
	htmlText, finalURL, err := FetchHTML(ctx, rawURL, opt)
	if err != nil {
		return "", "", err
	}
	text, err := ExtractText(htmlText)
	if err != nil {
		return "", "", err
	}
	return text, finalURL, nil
}

type PageResult struct {
	URL         string `json:"url"`
	FinalURL    string `json:"final_url"`
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type"`
	Title       string `json:"title,omitempty"`
	IsPDF       bool   `json:"is_pdf"`
	UsedJS      bool   `json:"used_js"`
	Saved       struct {
		Raw      string `json:"raw"`
		Text     string `json:"text"`
		Meta     string `json:"meta"`
		Rendered string `json:"rendered,omitempty"`
	} `json:"saved"`
	SavedBase string `json:"-"`
	Text      string `json:"-"`
}

type ProcessOptions struct {
	OutDir       string
	JSFallback   bool
	JSTimeout    time.Duration
	PDFMaxPages  int
	LowTextChars int
}

func defaultProcessOptions() ProcessOptions {
	return ProcessOptions{
		OutDir:       ".",
		JSFallback:   false,
		JSTimeout:    30 * time.Second,
		PDFMaxPages:  0,
		LowTextChars: 200,
	}
}

func ProcessOne(ctx context.Context, rawURL string, opt *ProcessOptions) (*PageResult, error) {
	cfg := defaultProcessOptions()
	if opt != nil {
		cfg.OutDir = strings.TrimSpace(opt.OutDir)
		cfg.JSFallback = opt.JSFallback
		if opt.JSTimeout > 0 {
			cfg.JSTimeout = opt.JSTimeout
		}
		if opt.PDFMaxPages > 0 {
			cfg.PDFMaxPages = opt.PDFMaxPages
		}
		if opt.LowTextChars > 0 {
			cfg.LowTextChars = opt.LowTextChars
		}
	}

	saveFiles := strings.TrimSpace(cfg.OutDir) != ""
	if saveFiles {
		if err := os.MkdirAll(cfg.OutDir, 0755); err != nil {
			return nil, err
		}
	}

	usedJSFromStart := false
	res, err := Fetch(ctx, rawURL, nil)
	if cfg.JSFallback && isGoogleSearchURL(rawURL) {
		rendered, rerr := RenderWithChromedp(ctx, rawURL, &RenderOptions{Timeout: cfg.JSTimeout, Scroll: true})
		if rerr == nil && strings.TrimSpace(rendered) != "" {
			usedJSFromStart = true
			res = &FetchResult{
				StatusCode:  200,
				ContentType: "text/html; charset=utf-8",
				Body:        []byte(rendered),
				Headers:     http.Header{},
				FinalURL:    rawURL,
			}
			err = nil
		} else if rerr != nil {
			err = rerr
		} else {
			err = fmt.Errorf("empty rendered html")
		}
	}
	if err != nil && cfg.JSFallback {
		rendered, rerr := RenderWithChromedp(ctx, rawURL, &RenderOptions{Timeout: cfg.JSTimeout, Scroll: true})
		if rerr == nil && strings.TrimSpace(rendered) != "" {
			usedJSFromStart = true
			res = &FetchResult{
				StatusCode:  200,
				ContentType: "text/html; charset=utf-8",
				Body:        []byte(rendered),
				Headers:     http.Header{},
				FinalURL:    rawURL,
			}
			err = nil
		} else if rerr != nil {
			return nil, fmt.Errorf("%w (js fallback failed: %v)", err, rerr)
		} else {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("empty fetch result")
	}

	finalURL := res.FinalURL
	isPDF := looksLikePDF(res.ContentType, finalURL)
	base := ""
	metaPath := ""
	rawPath := ""
	textPath := ""
	if saveFiles {
		base = filepath.Join(cfg.OutDir, urlHash(finalURL))
		metaPath = base + ".json"
		rawPath = base + ".html"
		if isPDF {
			rawPath = base + ".pdf"
		}
		textPath = base + ".txt"

		if err := os.WriteFile(rawPath, res.Body, 0644); err != nil {
			return nil, err
		}
		if usedJSFromStart && base != "" {
			_ = os.WriteFile(base+".rendered.html", res.Body, 0644)
		}
	}

	usedJS := usedJSFromStart
	title := ""
	text := ""

	if isPDF {
		text, err = pdfToText(res.Body, cfg.PDFMaxPages)
		if err != nil {
			return nil, err
		}
	} else {
		htmlText := decodeHTMLBytes(res.Body, res.ContentType)
		title = ExtractTitle(htmlText)
		if tTitle, tText, rErr := ExtractReadabilityText(htmlText, finalURL); rErr == nil && strings.TrimSpace(tText) != "" {
			title = pickNonEmpty(tTitle, title)
			text = tText
		} else {
			text = extractBestText(htmlText, finalURL)
		}

		if cfg.JSFallback && !usedJSFromStart && looksLikeLowContent(text, cfg.LowTextChars) {
			rendered, rerr := RenderWithChromedp(ctx, finalURL, &RenderOptions{Timeout: cfg.JSTimeout, Scroll: true})
			if rerr == nil && strings.TrimSpace(rendered) != "" {
				usedJS = true
				if saveFiles && base != "" {
					_ = os.WriteFile(base+".rendered.html", []byte(rendered), 0644)
				}
				if tTitle, tText, rErr2 := ExtractReadabilityText(rendered, finalURL); rErr2 == nil && strings.TrimSpace(tText) != "" {
					title = pickNonEmpty(tTitle, title)
					text = tText
				} else {
					title2 := ExtractTitle(rendered)
					text2 := extractBestText(rendered, finalURL)
					if strings.TrimSpace(text2) != "" {
						title = pickNonEmpty(title2, title)
						text = text2
					}
				}
			}
		}
	}

	if saveFiles {
		if err := SaveTextFile(textPath, finalURL, text); err != nil {
			return nil, err
		}
	}

	out := &PageResult{
		URL:         rawURL,
		FinalURL:    finalURL,
		StatusCode:  res.StatusCode,
		ContentType: res.ContentType,
		Title:       title,
		IsPDF:       isPDF,
		UsedJS:      usedJS,
		SavedBase:   base,
		Text:        text,
	}
	if saveFiles {
		out.Saved.Raw = filepath.Base(rawPath)
		out.Saved.Text = filepath.Base(textPath)
		out.Saved.Meta = filepath.Base(metaPath)
		if usedJS {
			out.Saved.Rendered = filepath.Base(base + ".rendered.html")
		}

		buf, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(metaPath, buf, 0644); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func isGoogleSearchURL(rawURL string) bool {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u == nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(u.Host))
	if host == "" {
		return false
	}
	if host != "www.google.com" && host != "google.com" {
		return false
	}
	path := strings.TrimSpace(u.Path)
	if path != "/search" {
		return false
	}
	return true
}

type CrawlOptions struct {
	OutDir           string
	MaxPages         int
	Depth            int
	Delay            time.Duration
	SameDomainOnly   bool
	IncludePattern   string
	ExcludePattern   string
	FollowPagination bool
	JSFallback       bool
	JSTimeout        time.Duration
	PDFMaxPages      int
}

func defaultCrawlOptions() CrawlOptions {
	return CrawlOptions{
		OutDir:           ".",
		MaxPages:         1,
		Depth:            0,
		Delay:            0,
		SameDomainOnly:   false,
		IncludePattern:   "",
		ExcludePattern:   "",
		FollowPagination: false,
		JSFallback:       false,
		JSTimeout:        30 * time.Second,
		PDFMaxPages:      0,
	}
}

func Crawl(ctx context.Context, startURL string, opt *CrawlOptions) ([]*PageResult, error) {
	cfg := defaultCrawlOptions()
	if opt != nil {
		if strings.TrimSpace(opt.OutDir) != "" {
			cfg.OutDir = opt.OutDir
		}
		if opt.MaxPages > 0 {
			cfg.MaxPages = opt.MaxPages
		}
		if opt.Depth >= 0 {
			cfg.Depth = opt.Depth
		}
		if opt.Delay > 0 {
			cfg.Delay = opt.Delay
		}
		cfg.SameDomainOnly = opt.SameDomainOnly
		cfg.IncludePattern = opt.IncludePattern
		cfg.ExcludePattern = opt.ExcludePattern
		cfg.FollowPagination = opt.FollowPagination
		cfg.JSFallback = opt.JSFallback
		if opt.JSTimeout > 0 {
			cfg.JSTimeout = opt.JSTimeout
		}
		if opt.PDFMaxPages > 0 {
			cfg.PDFMaxPages = opt.PDFMaxPages
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}
	if err := os.MkdirAll(cfg.OutDir, 0755); err != nil {
		return nil, err
	}

	var includeRe *regexp.Regexp
	var excludeRe *regexp.Regexp
	var err error
	if strings.TrimSpace(cfg.IncludePattern) != "" {
		includeRe, err = regexp.Compile(cfg.IncludePattern)
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(cfg.ExcludePattern) != "" {
		excludeRe, err = regexp.Compile(cfg.ExcludePattern)
		if err != nil {
			return nil, err
		}
	}

	type item struct {
		u string
		d int
	}
	queue := []item{{u: startURL, d: 0}}
	seen := map[string]bool{}
	results := make([]*PageResult, 0, cfg.MaxPages)

	for len(queue) > 0 && len(results) < cfg.MaxPages {
		it := queue[0]
		queue = queue[1:]

		u := strings.TrimSpace(it.u)
		if u == "" {
			continue
		}
		if seen[u] {
			continue
		}
		seen[u] = true

		if cfg.SameDomainOnly && !sameDomain(startURL, u) {
			continue
		}
		if !shouldInclude(u, includeRe, excludeRe) {
			continue
		}

		r, err := ProcessOne(ctx, u, &ProcessOptions{
			OutDir:      cfg.OutDir,
			JSFallback:  cfg.JSFallback,
			JSTimeout:   cfg.JSTimeout,
			PDFMaxPages: cfg.PDFMaxPages,
		})
		if err != nil {
			failBase := filepath.Join(cfg.OutDir, urlHash(u))
			_ = os.WriteFile(failBase+".json", []byte(fmt.Sprintf(`{"url":%q,"error":%q}`+"\n", u, err.Error())), 0644)
			continue
		}
		results = append(results, r)

		if !r.IsPDF && it.d <= cfg.Depth {
			rawHTML, _ := os.ReadFile(r.SavedBase + ".html")
			if len(rawHTML) > 0 {
				htmlText := decodeHTMLBytes(rawHTML, r.ContentType)

				if cfg.FollowPagination {
					if next := findNextLink(htmlText, r.FinalURL); next != "" && !seen[next] {
						queue = append([]item{{u: next, d: it.d}}, queue...)
					}
				}

				if it.d < cfg.Depth {
					links := extractLinks(htmlText, r.FinalURL)
					for _, lk := range links {
						if !seen[lk] {
							queue = append(queue, item{u: lk, d: it.d + 1})
						}
					}
				}
			}
		}

		if cfg.Delay > 0 {
			select {
			case <-time.After(cfg.Delay):
			case <-ctx.Done():
				return results, ctx.Err()
			}
		}
	}

	return results, nil
}

func removeTags(root *html.Node, tags map[string]bool) {
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			if c.Type == html.ElementNode && tags[strings.ToLower(c.Data)] {
				n.RemoveChild(c)
			} else {
				walk(c)
			}
			c = next
		}
	}
	walk(root)
}

func findFirstElement(root *html.Node, tag string) *html.Node {
	tag = strings.ToLower(tag)
	var found *html.Node
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if found != nil {
			return
		}
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == tag {
			found = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
			if found != nil {
				return
			}
		}
	}
	walk(root)
	return found
}

func collectText(n *html.Node, out *[]string) {
	if n.Type == html.TextNode {
		s := strings.TrimSpace(n.Data)
		if s != "" {
			*out = append(*out, s)
		}
		return
	}

	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "br", "p", "div", "section", "article", "header", "footer", "li", "ul", "ol", "h1", "h2", "h3", "h4", "h5", "h6", "tr", "td", "th", "pre", "blockquote":
			*out = append(*out, "\n")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, out)
	}

	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "p", "div", "section", "article", "li", "pre", "blockquote":
			*out = append(*out, "\n")
		}
	}
}

var reMultiNL = regexp.MustCompile(`\n{3,}`)

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = reMultiNL.ReplaceAllString(s, "\n\n")

	lines := strings.Split(s, "\n")
	trimmed := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed = append(trimmed, strings.TrimSpace(line))
	}
	return strings.Join(trimmed, "\n")
}

func postProcessMarkdown(s string) string {
	s = normalizeNewlines(s)
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		t := strings.TrimRightFunc(line, unicode.IsSpace)
		if strings.TrimSpace(t) == "" {
			if !blank {
				out = append(out, "")
				blank = true
			}
			continue
		}
		blank = false
		out = append(out, t)
	}
	return strings.TrimSpace(normalizeNewlines(strings.Join(out, "\n")))
}

func postProcessExtractedText(s string) string {
	s = normalizeNewlines(s)
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	lines := strings.Split(s, "\n")
	nonEmpty := make([]string, 0, len(lines))
	for _, line := range lines {
		if t := strings.TrimSpace(line); t != "" {
			nonEmpty = append(nonEmpty, t)
		}
	}

	trimLeadingShort := false
	{
		n := 0
		small := 0
		for _, line := range nonEmpty {
			n++
			if runeLen(line) <= 3 {
				small++
			}
			if n >= 15 {
				break
			}
		}
		if n >= 10 && small >= 8 {
			trimLeadingShort = true
		}
	}

	out := make([]string, 0, len(lines))
	started := false
	prev := ""
	blank := false

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			if started && !blank {
				out = append(out, "")
				blank = true
			}
			continue
		}

		if trimLeadingShort && !started && runeLen(t) <= 3 {
			continue
		}

		if isJunkLine(t) {
			continue
		}

		if t == prev {
			continue
		}
		prev = t
		started = true
		blank = false
		out = append(out, t)
	}

	return strings.TrimSpace(normalizeNewlines(strings.Join(out, "\n")))
}

func runeLen(s string) int {
	return len([]rune(s))
}

func isJunkLine(s string) bool {
	rs := []rune(strings.TrimSpace(s))
	if len(rs) <= 1 {
		return true
	}

	alnum := 0
	digits := 0
	other := 0
	for _, r := range rs {
		if unicode.IsSpace(r) {
			continue
		}
		if unicode.IsLetter(r) {
			alnum++
			continue
		}
		if unicode.IsDigit(r) {
			alnum++
			digits++
			continue
		}
		other++
	}

	if alnum == 0 {
		return true
	}

	if digits == alnum && digits <= 2 {
		return true
	}

	if other > alnum*3 && len(rs) < 10 {
		return true
	}

	return false
}

func extractBestText(htmlText string, baseURL string) string {
	mdText, mdErr := ExtractMarkdown(htmlText, baseURL)
	if mdErr == nil && strings.TrimSpace(mdText) != "" {
		return mdText
	}
	plain, _ := ExtractText(htmlText)
	return plain
}

func firstNLines(s string, n int) []string {
	if n <= 0 {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return lines
	}
	return lines[:n]
}

func looksLikeHTML(contentType string, finalURL string) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml+xml") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(finalURL), ".html") || strings.HasSuffix(strings.ToLower(finalURL), "/")
}

func looksLikePDF(contentType string, finalURL string) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "application/pdf") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(finalURL), ".pdf")
}

func decodeHTMLBytes(b []byte, contentType string) string {
	utf8Text := string(b)
	head := b
	if len(head) > 8192 {
		head = head[:8192]
	}
	headLower := bytes.ToLower(head)
	if bytes.Contains(headLower, []byte("charset=utf-8")) || strings.Contains(strings.ToLower(contentType), "utf-8") {
		return utf8Text
	}
	r, err := charset.NewReader(bytes.NewReader(b), contentType)
	if err != nil {
		return utf8Text
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return utf8Text
	}
	decoded := string(out)
	if replacementCount(utf8Text) < 50 && mojibakeScore(decoded) > mojibakeScore(utf8Text)+8 {
		return utf8Text
	}
	return decoded
}

func replacementCount(s string) int {
	n := 0
	for _, r := range s {
		if r == '\uFFFD' {
			n++
		}
	}
	return n
}

func mojibakeScore(s string) int {
	score := 0
	for _, r := range s {
		switch r {
		case '锛', '銆', '鈥', '鎬', '鐨', '鍛', '绗', '鍏', '鍚', '鍧', '浼', '浠', '鎴', '绐', '璇', '鍙', '鍒', '涔', '浜', '绠', '绾':
			score++
		}
	}
	return score
}

func urlHash(rawURL string) string {
	h := sha1.Sum([]byte(strings.TrimSpace(rawURL)))
	return hex.EncodeToString(h[:])
}

func looksLikeLowContent(text string, minChars int) bool {
	return len([]rune(strings.TrimSpace(text))) < minChars
}

func pickNonEmpty(a string, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func normalizeURL(baseURL string, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	l := strings.ToLower(href)
	if strings.HasPrefix(l, "javascript:") || strings.HasPrefix(l, "mailto:") || strings.HasPrefix(l, "tel:") {
		return ""
	}
	b, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	u = b.ResolveReference(u)
	u.Fragment = ""
	return u.String()
}

func extractLinks(htmlText string, baseURL string) []string {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil
	}
	out := make([]string, 0, 32)
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "a" {
			for _, a := range n.Attr {
				if strings.ToLower(a.Key) == "href" {
					if u := normalizeURL(baseURL, a.Val); u != "" {
						out = append(out, u)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return out
}

func findNextLink(htmlText string, baseURL string) string {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return ""
	}
	var next string
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if next != "" {
			return
		}
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			if tag == "a" || tag == "link" {
				rel := ""
				href := ""
				class := ""
				for _, a := range n.Attr {
					k := strings.ToLower(a.Key)
					switch k {
					case "rel":
						rel = strings.ToLower(strings.TrimSpace(a.Val))
					case "href":
						href = strings.TrimSpace(a.Val)
					case "class":
						class = strings.ToLower(strings.TrimSpace(a.Val))
					}
				}
				if href != "" && (strings.Contains(rel, "next") || strings.Contains(class, "next")) {
					next = normalizeURL(baseURL, href)
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
			if next != "" {
				return
			}
		}
	}
	walk(doc)
	return next
}

func sameDomain(a string, b string) bool {
	ua, err := url.Parse(strings.TrimSpace(a))
	if err != nil || ua.Host == "" {
		return false
	}
	ub, err := url.Parse(strings.TrimSpace(b))
	if err != nil || ub.Host == "" {
		return false
	}
	return strings.EqualFold(ua.Hostname(), ub.Hostname())
}

func shouldInclude(u string, include *regexp.Regexp, exclude *regexp.Regexp) bool {
	if include != nil && !include.MatchString(u) {
		return false
	}
	if exclude != nil && exclude.MatchString(u) {
		return false
	}
	return true
}

func ExtractTitle(htmlText string) string {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return ""
	}
	var title string
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if title != "" {
			return
		}
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				title = strings.TrimSpace(n.FirstChild.Data)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
			if title != "" {
				return
			}
		}
	}
	walk(doc)
	return title
}

func pdfToText(b []byte, maxPages int) (string, error) {
	r := bytes.NewReader(b)
	reader, err := pdf.NewReader(r, int64(len(b)))
	if err != nil {
		return "", err
	}
	n := reader.NumPage()
	if maxPages > 0 && maxPages < n {
		n = maxPages
	}
	var sb strings.Builder
	for i := 1; i <= n; i++ {
		page := reader.Page(i)
		txt, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		if i > 1 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(txt)
	}
	return strings.TrimSpace(normalizeNewlines(sb.String())), nil
}
