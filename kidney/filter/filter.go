package filter

import (
	"strings"
	"unicode/utf8"

	"github.com/Yoak3n/aimin/blood/schema"
)

type Context struct {
	Query     string
	Records   []schema.ConversationRecord
	Metadata  map[string]interface{}
}

type Filter interface {
	Name() string
	Apply(ctx *Context) (*Context, error)
}

type FilterFunc struct {
	name string
	fn   func(ctx *Context) (*Context, error)
}

func NewFilterFunc(name string, fn func(ctx *Context) (*Context, error)) *FilterFunc {
	return &FilterFunc{name: name, fn: fn}
}

func (f *FilterFunc) Name() string {
	return f.name
}

func (f *FilterFunc) Apply(ctx *Context) (*Context, error) {
	return f.fn(ctx)
}

type Chain struct {
	filters []Filter
}

func NewChain(filters ...Filter) *Chain {
	return &Chain{filters: filters}
}

func (c *Chain) Add(filters ...Filter) *Chain {
	c.filters = append(c.filters, filters...)
	return c
}

func (c *Chain) Apply(ctx *Context) (*Context, error) {
	var err error
	for _, f := range c.filters {
		ctx, err = f.Apply(ctx)
		if err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

func DeduplicateFilter() Filter {
	return NewFilterFunc("deduplicate", func(ctx *Context) (*Context, error) {
		seen := make(map[string]bool)
		result := make([]schema.ConversationRecord, 0, len(ctx.Records))
		for _, r := range ctx.Records {
			if !seen[r.Id] {
				seen[r.Id] = true
				result = append(result, r)
			}
		}
		ctx.Records = result
		return ctx, nil
	})
}

func EmptyFilter() Filter {
	return NewFilterFunc("empty", func(ctx *Context) (*Context, error) {
		result := make([]schema.ConversationRecord, 0, len(ctx.Records))
		for _, r := range ctx.Records {
			if strings.TrimSpace(r.Question) != "" || strings.TrimSpace(r.Answer) != "" {
				result = append(result, r)
			}
		}
		ctx.Records = result
		return ctx, nil
	})
}

func LengthFilter(maxQuestionRunes, maxAnswerRunes int) Filter {
	return NewFilterFunc("length", func(ctx *Context) (*Context, error) {
		result := make([]schema.ConversationRecord, 0, len(ctx.Records))
		for _, r := range ctx.Records {
			qLen := utf8.RuneCountInString(strings.TrimSpace(r.Question))
			aLen := utf8.RuneCountInString(strings.TrimSpace(r.Answer))
			if (maxQuestionRunes <= 0 || qLen <= maxQuestionRunes) &&
				(maxAnswerRunes <= 0 || aLen <= maxAnswerRunes) {
				result = append(result, r)
			}
		}
		ctx.Records = result
		return ctx, nil
	})
}

func KeywordFilter(keywords []string, caseSensitive bool) Filter {
	return NewFilterFunc("keyword", func(ctx *Context) (*Context, error) {
		if len(keywords) == 0 {
			return ctx, nil
		}
		result := make([]schema.ConversationRecord, 0, len(ctx.Records))
		for _, r := range ctx.Records {
			text := r.Question + " " + r.Answer
			if !caseSensitive {
				text = strings.ToLower(text)
			}
			matched := false
			for _, kw := range keywords {
				k := kw
				if !caseSensitive {
					k = strings.ToLower(k)
				}
				if strings.Contains(text, k) {
					matched = true
					break
				}
			}
			if matched {
				result = append(result, r)
			}
		}
		ctx.Records = result
		return ctx, nil
	})
}

func LimitFilter(limit int) Filter {
	return NewFilterFunc("limit", func(ctx *Context) (*Context, error) {
		if limit > 0 && len(ctx.Records) > limit {
			ctx.Records = ctx.Records[:limit]
		}
		return ctx, nil
	})
}

func TransformFilter(transform func(schema.ConversationRecord) schema.ConversationRecord) Filter {
	return NewFilterFunc("transform", func(ctx *Context) (*Context, error) {
		for i := range ctx.Records {
			ctx.Records[i] = transform(ctx.Records[i])
		}
		return ctx, nil
	})
}

func TrimSpaceFilter() Filter {
	return TransformFilter(func(r schema.ConversationRecord) schema.ConversationRecord {
		r.Question = strings.TrimSpace(r.Question)
		r.Answer = strings.TrimSpace(r.Answer)
		r.Thoughts = strings.TrimSpace(r.Thoughts)
		r.System = strings.TrimSpace(r.System)
		return r
	})
}
