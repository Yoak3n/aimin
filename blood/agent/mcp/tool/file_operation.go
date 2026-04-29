package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yoak3n/aimin/blood/config"
)

func FileOperation(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "args is empty"
	}
	args := parseArgsN(p, 3)
	op := strings.ToLower(strings.TrimSpace(firstNonEmpty(args["op"], args["action"], args["_0"])))
	path := strings.TrimSpace(firstNonEmpty(args["path"], args["_1"]))
	content := firstNonEmpty(args["content"], args["_2"])
	if op == "" {
		return fmt.Sprintf("args %s is invalid", p)
	}
	switch op {
	case "read":
		if path == "" {
			return fmt.Sprintf("args %s is invalid", p)
		}
		abs, err := resolveFileOpPath(path)
		if err != nil {
			return fmt.Sprintf("access denied: %s", err.Error())
		}
		return ReadFile(abs)
	case "write":
		if path == "" || strings.TrimSpace(content) == "" {
			return fmt.Sprintf("args %s is invalid", p)
		}
		abs, err := resolveFileOpPath(path)
		if err != nil {
			return fmt.Sprintf("access denied: %s", err.Error())
		}
		return WriteFile(abs, content)
	case "append":
		if path == "" || strings.TrimSpace(content) == "" {
			return fmt.Sprintf("args %s is invalid", p)
		}
		abs, err := resolveFileOpPath(path)
		if err != nil {
			return fmt.Sprintf("access denied: %s", err.Error())
		}
		return AppendFile(abs, content)
	default:
		return fmt.Sprintf("file operation %s not found", op)
	}
}

// func resolveWorkspacePath(p string) (string, error) {
// 	if p == "" {
// 		return "", fmt.Errorf("path is empty")
// 	}

// 	workspaceRoot := strings.TrimSpace(config.GlobalConfiguration().Workspace.Path)
// 	if workspaceRoot == "" {
// 		return "", fmt.Errorf("workspace path is empty")
// 	}

// 	p = filepath.FromSlash(p)
// 	if filepath.IsAbs(p) {
// 		return filepath.Clean(p), nil
// 	}

// 	p = strings.TrimLeft(p, `\/`)
// 	abs := filepath.Clean(filepath.Join(workspaceRoot, p))

// 	rel, err := filepath.Rel(workspaceRoot, abs)
// 	if err != nil {
// 		return "", err
// 	}
// 	rel = filepath.Clean(rel)
// 	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
// 		return "", fmt.Errorf("path escapes workspace")
// 	}
// 	return abs, nil
// }

func AppendFile(path, content string) string {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("append file failed with error: %s", err.Error())
	}
	return "append file success"
}

func ReadFile(path string) string {
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("read file failed with error: %s", err.Error())
	}
	return string(buf)
}

func WriteFile(path, content string) string {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("write file failed with error: %s", err.Error())
	}
	return "write file success"
}

func resolveFileOpPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("path is empty")
	}
	cfg := config.GlobalConfiguration()
	ws := ""
	if cfg != nil && cfg.Workspace != nil {
		ws = strings.TrimSpace(cfg.Workspace.Path)
	}
	mode := ""
	var deny []string
	if cfg != nil && cfg.Workspace != nil {
		mode = strings.ToLower(strings.TrimSpace(cfg.Workspace.AccessMode))
		deny = append([]string(nil), cfg.Workspace.DenyPaths...)
	}
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(config.DefaultWorkspace().AccessMode))
	}

	p = filepath.FromSlash(p)
	abs := ""
	if filepath.IsAbs(p) {
		abs = filepath.Clean(p)
	} else if ws != "" {
		p = strings.TrimLeft(p, `\/`)
		abs = filepath.Clean(filepath.Join(ws, p))
	} else {
		a, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		abs = filepath.Clean(a)
	}

	if strings.EqualFold(mode, "workspace_only") {
		if ws == "" {
			return "", fmt.Errorf("workspace path is empty")
		}
		r, err := filepath.Rel(ws, abs)
		if err != nil {
			return "", err
		}
		r = filepath.Clean(r)
		if r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("path escapes workspace")
		}
		return abs, nil
	}

	if strings.EqualFold(mode, "blacklist") {
		if len(deny) == 0 {
			deny = append([]string(nil), config.DefaultWorkspace().DenyPaths...)
		}
		if match, rule := isDeniedByGlobList(abs, deny); match {
			return "", fmt.Errorf("path matches deny rule: %s", rule)
		}
		return abs, nil
	}

	return abs, nil
}

func isDeniedByGlobList(absPath string, globs []string) (bool, string) {
	p := strings.ToLower(filepath.ToSlash(filepath.Clean(absPath)))
	tmp := strings.ToLower(filepath.ToSlash(filepath.Clean(os.TempDir())))
	if tmp != "" {
		if p == tmp || strings.HasPrefix(p, tmp+"/") {
			return false, ""
		}
	}
	for _, raw := range globs {
		g := strings.TrimSpace(raw)
		if g == "" {
			continue
		}
		g = strings.ToLower(filepath.ToSlash(strings.TrimSpace(g)))
		m, err := newGlobMatcher(g)
		if err != nil {
			continue
		}
		if m.Match(p) {
			return true, raw
		}
	}
	return false, ""
}
