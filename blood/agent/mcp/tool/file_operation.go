package tool

import (
	"fmt"
	"os"
	"strings"
)

func FileOperation(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "args is empty"
	}
	ps := strings.SplitN(p, ",", 3)
	if len(ps) >= 2 {
		op := strings.ToLower(strings.TrimSpace(ps[0]))
		// path, err := resolveWorkspacePath(strings.TrimSpace(ps[1]))
		// if err != nil {
		// 	return fmt.Sprintf("resolve path failed: %s", err.Error())
		// }
		switch op {
		case "read":
			return ReadFile(strings.TrimSpace(ps[1]))
		case "write":
			if len(ps) < 3 {
				return fmt.Sprintf("args %s is invalid", p)
			}
			return WriteFile(strings.TrimSpace(ps[1]), ps[2])
		case "append":
			if len(ps) < 3 {
				return fmt.Sprintf("args %s is invalid", p)
			}
			return AppendFile(strings.TrimSpace(ps[1]), ps[2])
		default:
			return fmt.Sprintf("file operation %s not found", ps[0])
		}
	} else {
		return fmt.Sprintf("args %s is invalid", p)
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
