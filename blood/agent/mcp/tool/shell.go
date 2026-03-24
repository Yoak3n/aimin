package tool

import (
	"fmt"
	"os/exec"
	"strings"
)

func ShellCommand(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "args is empty"
	}
	ps := strings.SplitN(p, ",", 2)
	if len(ps) != 2 {
		return fmt.Sprintf("invalid args format for ShellCommand: %s", p)
	}

	osType := strings.ToLower(strings.TrimSpace(ps[0]))
	commandStr := strings.TrimSpace(ps[1])

	var cmd *exec.Cmd

	switch osType {
	case "windows":
		cmd = exec.Command("cmd", "/C", commandStr)
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", commandStr)
	default:
		return fmt.Sprintf("unsupported os type: %s", osType)
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// 如果不是 ExitError，说明是其他系统错误（如找不到命令）
			return fmt.Sprintf("command execution failed: %s\nOutput: %s", err, string(output))
		}
		return fmt.Sprintf("command execution finished with error\nExit Code: %d\nOutput: %s", exitCode, string(output))
	}

	return fmt.Sprintf("command execution success\nExit Code: %d\nOutput: %s", exitCode, string(output))
}
