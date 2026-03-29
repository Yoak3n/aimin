package tool

import (
	"fmt"
	"strings"
)

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseArgs(s string) map[string]string {
	out := make(map[string]string)
	s = strings.TrimSpace(s)
	if s == "" {
		return out
	}

	parts := splitTopLevelCommas(s)
	pos := 0
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if k, v, ok := strings.Cut(p, "="); ok {
			k = strings.ToLower(strings.TrimSpace(k))
			v = strings.TrimSpace(v)
			v = strings.Trim(v, `"'`)
			out[k] = v
			continue
		}
		out[fmt.Sprintf("_%d", pos)] = strings.Trim(p, `"'`)
		pos++
	}
	return out
}

func splitTopLevelCommas(s string) []string {
	var out []string
	var b strings.Builder
	inSingle := false
	inDouble := false
	escape := false

	for _, r := range s {
		if escape {
			b.WriteRune(r)
			escape = false
			continue
		}
		if r == '\\' {
			escape = true
			b.WriteRune(r)
			continue
		}
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			b.WriteRune(r)
			continue
		}
		if r == '"' && !inSingle {
			inDouble = !inDouble
			b.WriteRune(r)
			continue
		}
		if r == ',' && !inSingle && !inDouble {
			out = append(out, b.String())
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	out = append(out, b.String())
	return out
}
