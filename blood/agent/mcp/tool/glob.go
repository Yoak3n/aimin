package tool

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func Glob(ctx *Context) string {
	p := strings.TrimSpace(ctx.GetPayload())
	if p == "" {
		return "args is empty"
	}

	ps := strings.SplitN(p, ",", 2)
	pattern := strings.TrimSpace(ps[0])
	if pattern == "" {
		return fmt.Sprintf("invalid args format for Glob: %s", p)
	}

	root := "."
	if len(ps) == 2 {
		if v := strings.TrimSpace(ps[1]); v != "" {
			root = v
		}
	}

	results, err := globInDir(root, pattern, 2000)
	if err != nil {
		return fmt.Sprintf("glob failed: %s", err.Error())
	}
	if len(results) == 0 {
		return "no matches"
	}
	return strings.Join(results, "\n")
}

func globInDir(root, pattern string, limit int) ([]string, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root is not a directory: %s", root)
	}

	matcher, err := newGlobMatcher(pattern)
	if err != nil {
		return nil, err
	}

	var out []string
	err = filepath.WalkDir(root, func(fullPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".idea" || name == ".vscode" {
				if fullPath != root {
					return filepath.SkipDir
				}
			}
			return nil
		}

		rel, err := filepath.Rel(root, fullPath)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if matcher.Match(rel) {
			out = append(out, fullPath)
			if limit > 0 && len(out) >= limit {
				return fs.SkipAll
			}
		}
		return nil
	})
	if err != nil && err != fs.SkipAll {
		return nil, err
	}

	sort.Strings(out)
	return out, nil
}

type globMatcher interface {
	Match(relSlashPath string) bool
}

func newGlobMatcher(pattern string) (globMatcher, error) {
	p := filepath.ToSlash(strings.TrimSpace(pattern))
	if p == "" {
		return nil, fmt.Errorf("empty glob pattern")
	}

	if strings.Contains(p, "**") {
		re, err := compileDoubleStarRegex(p)
		if err != nil {
			return nil, err
		}
		return &regexGlobMatcher{re: re}, nil
	}

	return &pathMatchGlobMatcher{pattern: p}, nil
}

type regexGlobMatcher struct {
	re *regexp.Regexp
}

func (m *regexGlobMatcher) Match(relSlashPath string) bool {
	return m.re.MatchString(relSlashPath)
}

type pathMatchGlobMatcher struct {
	pattern string
}

func (m *pathMatchGlobMatcher) Match(relSlashPath string) bool {
	ok, err := path.Match(m.pattern, relSlashPath)
	if err != nil {
		return false
	}
	return ok
}

func compileDoubleStarRegex(patternSlash string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(patternSlash); i++ {
		ch := patternSlash[i]
		switch ch {
		case '*':
			if i+1 < len(patternSlash) && patternSlash[i+1] == '*' {
				if i+2 < len(patternSlash) && patternSlash[i+2] == '/' {
					b.WriteString("(?:.*/)?")
					i += 2
				} else {
					b.WriteString(".*")
					i++
				}
			} else {
				b.WriteString(`[^/]*`)
			}
		case '?':
			b.WriteString(`[^/]`)
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '[', ']', '\\':
			b.WriteByte('\\')
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}
