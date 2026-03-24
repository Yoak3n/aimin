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
	ps := strings.Split(p, ",")
	if len(ps) >= 2 {
		op := strings.ToLower(strings.TrimSpace(ps[0]))
		switch op {
		case "read":
			return ReadFile(ps[1])
		case "write":
			if len(ps) != 3 {
				return fmt.Sprintf("args %s is invalid", p)
			}
			return WriteFile(ps[1], ps[2])
		case "append":
			if len(ps) != 3 {
				return fmt.Sprintf("args %s is invalid", p)
			}
			return AppendFile(ps[1], ps[2])
		default:
			return fmt.Sprintf("file operation %s not found", ps[0])
		}
	} else {
		return fmt.Sprintf("args %s is invalid", p)
	}
}

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
