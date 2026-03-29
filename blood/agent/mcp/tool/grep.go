package tool

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func Grep(ctx *Context) string {
	p := strings.TrimSpace(ctx.GetPayload())
	if p == "" {
		return "args is empty"
	}

	args := parseArgs(p)
	pattern := strings.TrimSpace(firstNonEmpty(args["pattern"], args["_0"]))
	if pattern == "" {
		return fmt.Sprintf("invalid args format for Grep: %s", p)
	}

	root := "."
	if v := strings.TrimSpace(firstNonEmpty(args["root"], args["_1"])); v != "" {
		root = v
	}

	fileGlob := strings.TrimSpace(firstNonEmpty(args["file_glob"], args["glob"], args["_2"]))

	maxMatches := 200
	if v := strings.TrimSpace(firstNonEmpty(args["max_matches"], args["limit"], args["_3"])); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return fmt.Sprintf("invalid max_matches: %s", v)
		}
		maxMatches = n
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Sprintf("invalid regex: %s", err.Error())
	}

	var glob globMatcher
	if fileGlob != "" {
		g, err := newGlobMatcher(fileGlob)
		if err != nil {
			return fmt.Sprintf("invalid file_glob: %s", err.Error())
		}
		glob = g
	}

	results, err := grepInDir(root, re, glob, maxMatches)
	if err != nil {
		return fmt.Sprintf("grep failed: %s", err.Error())
	}
	if len(results) == 0 {
		return "no matches"
	}
	return strings.Join(results, "\n")
}

func grepInDir(root string, re *regexp.Regexp, fileGlob globMatcher, maxMatches int) ([]string, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root is not a directory: %s", root)
	}

	var out []string
	matchCount := 0

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
		relSlash := filepath.ToSlash(rel)
		if fileGlob != nil && !fileGlob.Match(relSlash) {
			return nil
		}

		f, err := os.Open(fullPath)
		if err != nil {
			return nil
		}
		defer f.Close()

		if isBinaryFile(f) {
			return nil
		}

		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)

		lineNo := 0
		for sc.Scan() {
			lineNo++
			line := sc.Text()
			if re.MatchString(line) {
				out = append(out, fmt.Sprintf("%s:%d:%s", fullPath, lineNo, line))
				matchCount++
				if matchCount >= maxMatches {
					return fs.SkipAll
				}
			}
		}
		return nil
	})
	if err != nil && err != fs.SkipAll {
		return nil, err
	}

	return out, nil
}

func isBinaryFile(r io.ReadSeeker) bool {
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return true
	}
	defer r.Seek(0, io.SeekStart)

	buf := make([]byte, 8000)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return true
	}
	return bytes.IndexByte(buf[:n], 0) >= 0
}
