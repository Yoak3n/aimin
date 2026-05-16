package filter

import (
	"testing"

	"github.com/Yoak3n/aimin/blood/schema"
)

func TestDeduplicateFilter(t *testing.T) {
	ctx := &Context{
		Query: "test",
		Records: []schema.ConversationRecord{
			{Id: "1", Question: "q1"},
			{Id: "2", Question: "q2"},
			{Id: "1", Question: "q1"},
		},
	}

	chain := NewChain(DeduplicateFilter())
	result, err := chain.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(result.Records))
	}
}

func TestEmptyFilter(t *testing.T) {
	ctx := &Context{
		Query: "test",
		Records: []schema.ConversationRecord{
			{Id: "1", Question: "q1", Answer: "a1"},
			{Id: "2", Question: "", Answer: ""},
			{Id: "3", Question: "q3", Answer: ""},
		},
	}

	chain := NewChain(EmptyFilter())
	result, err := chain.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(result.Records))
	}
}

func TestLengthFilter(t *testing.T) {
	ctx := &Context{
		Query: "test",
		Records: []schema.ConversationRecord{
			{Id: "1", Question: "short"},
			{Id: "2", Question: "this is a longer question that exceeds the limit"},
		},
	}

	chain := NewChain(LengthFilter(10, 0))
	result, err := chain.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(result.Records))
	}
}

func TestLimitFilter(t *testing.T) {
	ctx := &Context{
		Query: "test",
		Records: []schema.ConversationRecord{
			{Id: "1"}, {Id: "2"}, {Id: "3"}, {Id: "4"}, {Id: "5"},
		},
	}

	chain := NewChain(LimitFilter(3))
	result, err := chain.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(result.Records))
	}
}

func TestChainMultipleFilters(t *testing.T) {
	ctx := &Context{
		Query: "test",
		Records: []schema.ConversationRecord{
			{Id: "1", Question: "q1", Answer: "a1"},
			{Id: "2", Question: "", Answer: ""},
			{Id: "1", Question: "q1", Answer: "a1"},
			{Id: "3", Question: "q3", Answer: "a3"},
		},
	}

	chain := NewChain(
		DeduplicateFilter(),
		EmptyFilter(),
		LimitFilter(2),
	)
	result, err := chain.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(result.Records))
	}
}

func TestChainAdd(t *testing.T) {
	chain := NewChain(DeduplicateFilter())
	chain.Add(EmptyFilter(), LimitFilter(5))

	if len(chain.filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(chain.filters))
	}
}
