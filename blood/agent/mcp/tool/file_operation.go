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
		return ReadFile(path)
	case "write":
		if path == "" || strings.TrimSpace(content) == "" {
			return fmt.Sprintf("args %s is invalid", p)
		}
		return WriteFile(path, content)
	case "append":
		if path == "" || strings.TrimSpace(content) == "" {
			return fmt.Sprintf("args %s is invalid", p)
		}
		return AppendFile(path, content)
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
