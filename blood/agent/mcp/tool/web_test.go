package tool

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeb_Fetch_ReturnsExtractedContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html>
  <head><title>X</title></head>
  <body>
    <main>
      <h1>Hello</h1>
      <p>World</p>
    </main>
  </body>
</html>`))
	}))
	defer srv.Close()

	ctx := NewMcpContext()
	ctx.SetPayload("fetch,url=" + srv.URL + ",js=false")
	out := Web(ctx)
	if strings.HasPrefix(out, "ERROR:") {
		t.Fatalf("unexpected error: %s", out)
	}
	if !strings.Contains(out, "<web_fetch>") {
		t.Fatalf("missing tag: %s", out)
	}
	if !strings.Contains(out, "World") {
		t.Fatalf("expected extracted content, got: %s", out)
	}
}

func TestWeb_Fetch_InferByURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><main>OK</main></body></html>`))
	}))
	defer srv.Close()

	ctx := NewMcpContext()
	ctx.SetPayload("fetch,url=" + srv.URL + ",js=false")
	out := Web(ctx)
	if strings.HasPrefix(out, "ERROR:") {
		t.Fatalf("unexpected error: %s", out)
	}
	if !strings.Contains(out, "OK") {
		t.Fatalf("expected extracted content, got: %s", out)
	}
}

func TestWeb_Search_MissingQuery(t *testing.T) {
	ctx := NewMcpContext()
	ctx.SetPayload("search")
	out := Web(ctx)
	if !strings.HasPrefix(out, "ERROR:") {
		t.Fatalf("expected error, got: %s", out)
	}
}
