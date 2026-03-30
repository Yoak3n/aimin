package bilibili

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/hand/pkg/requests"
	"github.com/tidwall/gjson"
)

const (
	SearchURLWbi    = "https://api.bilibili.com/x/web-interface/wbi/search/type"
	SearchURLLegacy = "https://api.bilibili.com/x/web-interface/search/type"
	NavURL          = "https://api.bilibili.com/x/web-interface/nav"
	HomeURL         = "https://www.bilibili.com"
)

type Video struct {
	AID      int64
	BVID     string
	Title    string
	Author   string
	MID      int64
	Pic      string
	Play     int64
	Danmaku  int64
	Duration string
	PubDate  int64
}

type SearchResult struct {
	Code      int
	Message   string
	Seid      string
	Page      int
	PageSize  int
	NumResult int
	NumPages  int
	Videos    []Video
	Raw       gjson.Result
}

type SearchParams struct {
	Keyword    string
	Page       int
	Order      string
	Cookie     string
	UserAgent  string
	SearchType string
	Timeout    time.Duration
}

type Topic struct {
	Name    string
	KeyWord []string
	Videos  []Video
	Err     error
}

func NewTopic(name string, topic []string) *Topic {
	t := &Topic{
		Name:    name,
		KeyWord: append([]string(nil), topic...),
	}
	t.fetchVideos()
	return t
}

func (t *Topic) fetchVideos() {
	kw := strings.Join(t.KeyWord, ",")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := SearchVideos(ctx, SearchParams{
		Keyword: kw,
		Page:    1,
		Order:   "totalrank",
	})
	if err != nil {
		t.Err = err
		return
	}
	t.Videos = res.Videos
	t.Err = nil
}

func SearchVideos(ctx context.Context, p SearchParams) (*SearchResult, error) {
	p.Keyword = strings.TrimSpace(p.Keyword)
	if p.Keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if strings.TrimSpace(p.Order) == "" {
		p.Order = "totalrank"
	}
	if strings.TrimSpace(p.SearchType) == "" {
		p.SearchType = "video"
	}
	if strings.TrimSpace(p.UserAgent) == "" {
		p.UserAgent = defaultUserAgent()
	}
	if p.Timeout <= 0 {
		p.Timeout = 10 * time.Second
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cookie, _ := ensureCookie(ctx, p.Cookie, p.UserAgent, p.Timeout)

	params := url.Values{}
	params.Set("search_type", p.SearchType)
	params.Set("keyword", p.Keyword)
	params.Set("page", fmt.Sprintf("%d", p.Page))
	params.Set("order", p.Order)

	var resp *requests.Response
	var err error
	resp, err = getSearchWbi(ctx, params, cookie, p.UserAgent, p.Timeout)
	if err != nil {
		resp, err = getSearchLegacy(ctx, params, cookie, p.UserAgent, p.Timeout)
		if err != nil {
			return nil, err
		}
	}

	code := int(resp.JSON.Get("code").Int())
	msg := resp.JSON.Get("message").String()
	if code != 0 {
		return &SearchResult{Code: code, Message: msg, Raw: resp.JSON}, fmt.Errorf("bilibili search failed: code=%d message=%s", code, msg)
	}

	out := &SearchResult{
		Code:      code,
		Message:   msg,
		Seid:      resp.JSON.Get("data.seid").String(),
		Page:      int(resp.JSON.Get("data.page").Int()),
		PageSize:  int(resp.JSON.Get("data.pagesize").Int()),
		NumResult: int(resp.JSON.Get("data.numResults").Int()),
		NumPages:  int(resp.JSON.Get("data.numPages").Int()),
		Videos:    nil,
		Raw:       resp.JSON,
	}

	results := resp.JSON.Get("data.result")
	if !results.Exists() || !results.IsArray() {
		return out, nil
	}

	videos := make([]Video, 0, len(results.Array()))
	for _, item := range results.Array() {
		videos = append(videos, Video{
			AID:      item.Get("aid").Int(),
			BVID:     item.Get("bvid").String(),
			Title:    cleanTitle(item.Get("title").String()),
			Author:   item.Get("author").String(),
			MID:      item.Get("mid").Int(),
			Pic:      item.Get("pic").String(),
			Play:     item.Get("play").Int(),
			Danmaku:  item.Get("danmaku").Int(),
			Duration: item.Get("duration").String(),
			PubDate:  item.Get("pubdate").Int(),
		})
	}
	out.Videos = videos
	return out, nil
}

func getSearchLegacy(ctx context.Context, params url.Values, cookie, ua string, timeout time.Duration) (*requests.Response, error) {
	q := map[string]string{}
	for k := range params {
		q[k] = params.Get(k)
	}
	headers := baseHeaders(cookie, ua)
	return requests.Get(ctx, SearchURLLegacy, &requests.Options{
		Query:   q,
		Headers: headers,
		Timeout: timeout,
	})
}

func getSearchWbi(ctx context.Context, params url.Values, cookie, ua string, timeout time.Duration) (*requests.Response, error) {
	imgKey, subKey, err := getWbiKeys(ctx, cookie, ua, timeout)
	if err != nil {
		return nil, err
	}

	mk := mixinKey(imgKey + subKey)
	query := signedWbiQuery(params, mk)
	fullURL := SearchURLWbi + "?" + query
	headers := baseHeaders(cookie, ua)
	return requests.Get(ctx, fullURL, &requests.Options{
		Headers: headers,
		Timeout: timeout,
	})
}

func getWbiKeys(ctx context.Context, cookie, ua string, timeout time.Duration) (string, string, error) {
	headers := baseHeaders(cookie, ua)
	resp, err := requests.Get(ctx, NavURL, &requests.Options{
		Headers: headers,
		Timeout: timeout,
	})
	if err != nil {
		return "", "", err
	}
	code := int(resp.JSON.Get("code").Int())
	if code != 0 {
		msg := resp.JSON.Get("message").String()
		return "", "", fmt.Errorf("get wbi keys failed: code=%d message=%s", code, msg)
	}

	imgURL := resp.JSON.Get("data.wbi_img.img_url").String()
	subURL := resp.JSON.Get("data.wbi_img.sub_url").String()
	imgKey := keyFromWbiURL(imgURL)
	subKey := keyFromWbiURL(subURL)
	if imgKey == "" || subKey == "" {
		return "", "", fmt.Errorf("get wbi keys failed: invalid wbi_img")
	}
	return imgKey, subKey, nil
}

func keyFromWbiURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err == nil && u.Path != "" {
		raw = u.Path
	}
	base := path.Base(raw)
	if base == "." || base == "/" {
		return ""
	}
	ext := path.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func signedWbiQuery(params url.Values, mixinKey string) string {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	params = cloneValues(params)
	params.Set("wts", ts)

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := sanitizeWbiValue(params.Get(k))
		parts = append(parts, wbiEscape(k)+"="+wbiEscape(v))
	}
	query := strings.Join(parts, "&")

	sum := md5.Sum([]byte(query + mixinKey))
	wRid := hex.EncodeToString(sum[:])
	return query + "&w_rid=" + wRid
}

func wbiEscape(s string) string {
	esc := url.QueryEscape(s)
	esc = strings.ReplaceAll(esc, "+", "%20")
	return esc
}

func sanitizeWbiValue(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '!', '\'', '(', ')', '*':
			return -1
		default:
			return r
		}
	}, s)
}

func cloneValues(v url.Values) url.Values {
	out := url.Values{}
	for k, vv := range v {
		out[k] = append([]string(nil), vv...)
	}
	return out
}

func mixinKey(origin string) string {
	tab := []int{
		46, 47, 18, 2, 53, 8, 23, 32,
		15, 50, 10, 31, 58, 3, 45, 35,
		27, 43, 5, 49, 33, 9, 42, 19,
		29, 28, 14, 39, 12, 38, 41, 13,
		37, 48, 7, 16, 24, 55, 40, 61,
		26, 17, 0, 1, 60, 51, 30, 4,
		22, 25, 54, 21, 56, 59, 6, 63,
		57, 62, 11, 36, 20, 34, 44, 52,
	}
	buf := make([]byte, 0, len(tab))
	for _, i := range tab {
		if i >= 0 && i < len(origin) {
			buf = append(buf, origin[i])
		}
	}
	if len(buf) > 32 {
		buf = buf[:32]
	}
	return string(buf)
}

func ensureCookie(ctx context.Context, cookie, ua string, timeout time.Duration) (string, error) {
	cookie = strings.TrimSpace(cookie)
	if cookie != "" {
		return cookie, nil
	}
	resp, err := requests.Get(ctx, HomeURL, &requests.Options{
		Headers: map[string]string{
			"User-Agent": ua,
		},
		Timeout: timeout,
	})
	if err != nil {
		return "", err
	}
	return cookieFromHeaders(resp.Headers), nil
}

func cookieFromHeaders(h http.Header) string {
	set := h.Values("Set-Cookie")
	if len(set) == 0 {
		return ""
	}
	parts := make([]string, 0, len(set))
	for _, v := range set {
		seg := strings.SplitN(v, ";", 2)[0]
		seg = strings.TrimSpace(seg)
		if seg != "" {
			parts = append(parts, seg)
		}
	}
	return strings.Join(parts, "; ")
}

func baseHeaders(cookie, ua string) map[string]string {
	h := map[string]string{
		"User-Agent": ua,
		"Referer":    HomeURL + "/",
	}
	if strings.TrimSpace(cookie) != "" {
		h["Cookie"] = cookie
	}
	return h
}

func defaultUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
}

func cleanTitle(s string) string {
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "<em class=\"keyword\">", "")
	s = strings.ReplaceAll(s, "</em>", "")
	s = strings.ReplaceAll(s, "<em class='keyword'>", "")
	return strings.TrimSpace(s)
}
