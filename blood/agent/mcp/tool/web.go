package tool

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	handfetch "github.com/Yoak3n/aimin/hand/internet/fetch"
	"github.com/Yoak3n/aimin/hand/internet/search/duckduckgo"
)

func Web(ctx *Context) string {
	raw := strings.TrimSpace(ctx.GetPayload())
	if raw == "" {
		return "ERROR: args is empty"
	}
	args := parseArgsN(raw, 0)

	action := strings.ToLower(strings.TrimSpace(firstNonEmpty(args["action"], args["_0"])))
	if action == "" {
		v := strings.TrimSpace(firstNonEmpty(args["url"], args["query"], args["_0"]))
		if looksLikeURL(v) {
			action = "fetch"
			args["url"] = v
		} else {
			action = "search"
			args["query"] = v
		}
	} else if !isSupportedWebAction(action) {
		if looksLikeURL(action) {
			args["url"] = action
			action = "fetch"
		} else {
			args["query"] = action
			action = "search"
		}
	}

	switch action {
	case "fetch", "read":
		return webFetch(ctx, args)
	case "search":
		return webSearch(ctx, args)
	default:
		return "ERROR: unsupported action: " + action
	}
}

func isSupportedWebAction(action string) bool {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "fetch", "read", "search":
		return true
	default:
		return false
	}
}

func webFetch(ctx *Context, args map[string]string) string {
	rawURL := strings.TrimSpace(firstNonEmpty(args["url"], args["_1"]))
	if rawURL == "" {
		return "ERROR: missing url"
	}

	timeout := parseDurationSeconds(firstNonEmpty(args["timeout_s"], args["timeout"], args["t"]), 20*time.Second)
	js := parseBool(firstNonEmpty(args["js"], args["render"], args["javascript"]), true)
	pdfMaxPages := parseInt(firstNonEmpty(args["pdf_max_pages"], args["pdf_max"], args["pdf_pages"]), 0)

	opts := &handfetch.ProcessOptions{
		OutDir:      strings.TrimSpace(firstNonEmpty(args["out_dir"], args["out"])),
		JSFallback:  js,
		JSTimeout:   timeout,
		PDFMaxPages: pdfMaxPages,
	}

	cctx := ctx.Ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		cctx, cancel = withTimeout(cctx, timeout)
		defer cancel()
	}

	res, err := handfetch.FetchPage(cctx, rawURL, opts)
	if err != nil {
		return "ERROR: " + err.Error()
	}

	sb := strings.Builder{}
	sb.WriteString("<web_fetch>\n")
	fmt.Fprintf(&sb, "<url>%s</url>\n", xmlTextCompact(rawURL, 500))
	fmt.Fprintf(&sb, "<final_url>%s</final_url>\n", xmlTextCompact(res.FinalURL, 500))
	if strings.TrimSpace(res.Title) != "" {
		fmt.Fprintf(&sb, "<title>%s</title>\n", xmlTextCompact(res.Title, 500))
	}
	if res.IsPDF {
		sb.WriteString("<is_pdf>true</is_pdf>\n")
	} else {
		sb.WriteString("<is_pdf>false</is_pdf>\n")
	}
	if res.UsedJS {
		sb.WriteString("<used_js>true</used_js>\n")
	} else {
		sb.WriteString("<used_js>false</used_js>\n")
	}
	sb.WriteString("<content><![CDATA[")
	sb.WriteString(sanitizeCDATA(res.Text))
	sb.WriteString("]]></content>\n")
	sb.WriteString("</web_fetch>")
	return sb.String()
}

func webSearch(ctx *Context, args map[string]string) string {
	query := strings.TrimSpace(firstNonEmpty(args["query"], args["_1"]))
	if query == "" {
		return "ERROR: missing query"
	}

	limit := parseInt(firstNonEmpty(args["limit"], args["n"]), 5)
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	timeout := parseDurationSeconds(firstNonEmpty(args["timeout_s"], args["timeout"], args["t"]), 15*time.Second)

	cctx := ctx.Ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		cctx, cancel = withTimeout(cctx, timeout)
		defer cancel()
	}

	res, err := duckduckgo.Search(cctx, query, &duckduckgo.Options{Timeout: timeout})
	if err != nil {
		return "ERROR: " + err.Error()
	}

	items := make([]duckduckgo.ResultItem, 0, len(res.Results)+len(res.RelatedTopics))
	items = append(items, res.Results...)
	items = append(items, res.RelatedTopics...)
	if len(items) > limit {
		items = items[:limit]
	}

	sb := strings.Builder{}
	sb.WriteString("<web_search>\n")
	fmt.Fprintf(&sb, "<query>%s</query>\n", xmlTextCompact(query, 400))
	if strings.TrimSpace(res.Heading) != "" {
		fmt.Fprintf(&sb, "<heading>%s</heading>\n", xmlTextCompact(res.Heading, 400))
	}
	if strings.TrimSpace(res.AbstractText) != "" {
		sb.WriteString("<abstract><![CDATA[")
		sb.WriteString(sanitizeCDATA(res.AbstractText))
		sb.WriteString("]]></abstract>\n")
	}
	sb.WriteString("<results>\n")
	for _, it := range items {
		sb.WriteString("<item>\n")
		fmt.Fprintf(&sb, "<url>%s</url>\n", xmlTextCompact(it.FirstURL, 800))
		sb.WriteString("<snippet><![CDATA[")
		sb.WriteString(sanitizeCDATA(it.Text))
		sb.WriteString("]]></snippet>\n")
		sb.WriteString("</item>\n")
	}
	sb.WriteString("</results>\n</web_search>")
	return sb.String()
}

func parseBool(s string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(s))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return def
	}
}

func parseInt(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseDurationSeconds(s string, def time.Duration) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	if strings.ContainsAny(s, "hms") {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	if f <= 0 {
		return def
	}
	return time.Duration(float64(time.Second) * f)
}

func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, d)
}

func looksLikeURL(s string) bool {
	ls := strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(ls, "http://") || strings.HasPrefix(ls, "https://")
}

func xmlTextCompact(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if max > 0 && len([]rune(s)) > max {
		rs := []rune(s)
		s = string(rs[:max]) + "..."
	}
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func sanitizeCDATA(s string) string {
	return strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
}
