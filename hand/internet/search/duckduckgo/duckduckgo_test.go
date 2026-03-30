package duckduckgo

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSearch_ReturnsSomething(t *testing.T) {
	if strings.TrimSpace(os.Getenv("DUCKDUCKGO_RUN")) != "1" {
		t.Skip("set DUCKDUCKGO_RUN=1 to run this integration test")
	}

	query := strings.TrimSpace(os.Getenv("DUCKDUCKGO_QUERY"))
	if query == "" {
		query = "golang"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := Search(ctx, query, &Options{Timeout: 25 * time.Second})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if strings.TrimSpace(res.Heading) == "" &&
		strings.TrimSpace(res.AbstractText) == "" &&
		strings.TrimSpace(res.Answer) == "" &&
		len(res.Results) == 0 &&
		len(res.RelatedTopics) == 0 {
		t.Fatalf("empty response: heading=%q abstract=%q answer=%q results=%d related=%d", res.Heading, res.AbstractText, res.Answer, len(res.Results), len(res.RelatedTopics))
	}

	t.Logf("heading=%q", res.Heading)
	t.Logf("abstract=%q", res.AbstractText)
	t.Logf("answer=%q answerType=%q", res.Answer, res.AnswerType)
	t.Logf("results=%d related=%d", len(res.Results), len(res.RelatedTopics))

	n := 5
	if len(res.Results) < n {
		n = len(res.Results)
	}
	for i := 0; i < n; i++ {
		t.Logf("[%d] url=%s text=%s", i+1, res.Results[i].FirstURL, res.Results[i].Text)
	}

	if len(res.Results) == 0 {
		n = 5
		if len(res.RelatedTopics) < n {
			n = len(res.RelatedTopics)
		}
		for i := 0; i < n; i++ {
			t.Logf("[related %d] url=%s text=%s", i+1, res.RelatedTopics[i].FirstURL, res.RelatedTopics[i].Text)
		}
	}
}
