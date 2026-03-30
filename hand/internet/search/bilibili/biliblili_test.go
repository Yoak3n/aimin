package bilibili

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
)

func TestSearchVideos_ReturnsVideoList(t *testing.T) {
	cookie := strings.TrimSpace(os.Getenv("BILIBILI_COOKIE"))
	if cookie == "" {
		cfg := config.GlobalConfiguration()
		if cfg != nil && cfg.Internet != nil {
			cookie = strings.TrimSpace(cfg.Internet.BilibiliCookie)
		}
	}
	if cookie == "" {
		t.Skip("missing bilibili cookie: set env BILIBILI_COOKIE or config.json internet.bilibili_cookie")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := SearchVideos(ctx, SearchParams{
		Keyword: "洛天依",
		Page:    1,
		Order:   "totalrank",
		Cookie:  cookie,
	})
	if err != nil {
		if res != nil && res.Code == -412 {
			t.Skip("bilibili returned -412; cookie likely insufficient (need buvid3 and wbi-related validation)")
		}
		t.Fatalf("search failed: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if len(res.Videos) == 0 {
		t.Fatalf("empty videos; code=%d message=%s", res.Code, res.Message)
	}

	n := 5
	if len(res.Videos) < n {
		n = len(res.Videos)
	}
	for i := 0; i < n; i++ {
		v := res.Videos[i]
		t.Logf("[%d] bvid=%s aid=%d author=%s title=%s", i+1, v.BVID, v.AID, v.Author, v.Title)
	}

	v := res.Videos[0]
	if v.BVID == "" && v.AID == 0 {
		t.Fatalf("invalid first video: %#v", v)
	}
}
