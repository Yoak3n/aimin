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
	return parseArgsN(s, 0)
}

func parseArgsN(s string, maxPositional int) map[string]string {
	out := make(map[string]string)
	s = strings.Trim(s, " \t\r")
	if s == "" {
		return out
	}

	parts := splitTopLevelCommasN(s, maxPositional)
	pos := 0
	for _, p := range parts {
		soft := strings.Trim(p, " \t\r")
		if soft == "" {
			continue
		}
		if k, v, ok := strings.Cut(soft, "="); ok {
			k = strings.ToLower(strings.TrimSpace(k))
			v = strings.Trim(v, " \t\r")
			v = strings.Trim(v, `"'`)
			out[k] = v
			continue
		}
		v := strings.Trim(soft, `"'`)
		out[fmt.Sprintf("_%d", pos)] = v
		pos++
	}
	return out
}

func splitTopLevelCommasN(s string, maxParts int) []string {
	var out []string
	var b strings.Builder
	inSingle := false
	inDouble := false
	escape := false
	partCount := 0

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
			if maxParts > 0 && partCount >= maxParts-1 {
				b.WriteRune(r)
				continue
			}
			out = append(out, b.String())
			b.Reset()
			partCount++
			continue
		}
		b.WriteRune(r)
	}
	out = append(out, b.String())
	return out
}
