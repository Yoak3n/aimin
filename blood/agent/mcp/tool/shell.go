package tool

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func ShellCommand(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "ERROR: args is empty"
	}
	args := parseArgsN(p, 2)
	osType := strings.ToLower(strings.TrimSpace(firstNonEmpty(args["os_type"], args["os"], args["_0"])))
	commandStr := strings.TrimSpace(firstNonEmpty(args["command"], args["_1"]))
	if osType == "" || commandStr == "" {
		return fmt.Sprintf("ERROR: invalid args format for ShellCommand: %s", p)
	}

	switch osType {
	case "windows":
		commandStr = sanitizeWindowsCommand(commandStr)
		commandStr = normalizeWindowsCommand(commandStr)

		out, exitCode, err := runWindowsCmd(commandStr)
		if err != nil {
			if fallback, ok := windowsFallback(commandStr); ok {
				return fallback
			}
			if exitCode == 0 {
				return fmt.Sprintf("ERROR: command execution failed: %s\nOutput: %s", err, out)
			}
			return fmt.Sprintf("ERROR: command execution finished with error\nExit Code: %d\nOutput: %s", exitCode, out)
		}
		return fmt.Sprintf("command execution success\nExit Code: %d\nOutput: %s", exitCode, out)
	case "linux", "darwin":
		cmd := exec.Command("sh", "-c", commandStr)
		output, err := cmd.CombinedOutput()
		out := string(output)
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				return fmt.Sprintf("ERROR: command execution failed: %s\nOutput: %s", err, out)
			}
			return fmt.Sprintf("ERROR: command execution finished with error\nExit Code: %d\nOutput: %s", exitCode, out)
		}
		return fmt.Sprintf("command execution success\nExit Code: %d\nOutput: %s", exitCode, out)
	default:
		return fmt.Sprintf("ERROR: unsupported os type: %s", osType)
	}
}

func runWindowsCmd(commandStr string) (string, int, error) {
	cmdLine := "chcp 65001>nul & " + commandStr
	cmd := exec.Command("cmd", "/D", "/C", cmdLine)
	output, err := cmd.CombinedOutput()
	out := decodeWindowsOutput(output)
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		return out, exitCode, err
	}
	return out, exitCode, nil
}

func sanitizeWindowsCommand(s string) string {
	s = strings.TrimSpace(s)
	r := strings.NewReplacer(
		"\uFEFF", "",
		"\u200B", "",
		"\u200C", "",
		"\u200D", "",
		"\u2060", "",
		"\u00A0", " ",
		"“", `"`,
		"”", `"`,
		"„", `"`,
		"‟", `"`,
		"’", "'",
		"‘", "'",
		"‛", "'",
	)
	return strings.TrimSpace(r.Replace(s))
}

func normalizeWindowsCommand(s string) string {
	ls := strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(ls, "dir ") ||
		strings.HasPrefix(ls, "del ") ||
		strings.HasPrefix(ls, "erase ") ||
		strings.HasPrefix(ls, "type ") ||
		strings.HasPrefix(ls, "copy ") ||
		strings.HasPrefix(ls, "move ") ||
		strings.HasPrefix(ls, "cd ") ||
		strings.HasPrefix(ls, "rmdir ") ||
		strings.HasPrefix(ls, "rd ") ||
		strings.HasPrefix(ls, "mkdir ") {
		if strings.Contains(s, "/") && !strings.Contains(strings.ToLower(s), "http://") && !strings.Contains(strings.ToLower(s), "https://") {
			s = strings.ReplaceAll(s, "/", `\`)
		}
	}
	return s
}

func decodeWindowsOutput(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	if utf8.Valid(b) {
		return string(b)
	}
	r := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(r)
	if err == nil && utf8.Valid(decoded) {
		return string(decoded)
	}
	return string(b)
}

func windowsFallback(commandStr string) (string, bool) {
	target, ok := parseWindowsDelTarget(commandStr)
	if ok {
		ps := fmt.Sprintf(`Remove-Item -LiteralPath '%s' -Force -ErrorAction Stop`, escapePSSingleQuoted(target))
		out, exitCode, err := runWindowsPowerShell(ps)
		if err != nil {
			return fmt.Sprintf("ERROR: command execution finished with error\nExit Code: %d\nOutput: %s", exitCode, out), true
		}
		return fmt.Sprintf("command execution success\nExit Code: %d\nOutput: %s", exitCode, out), true
	}
	return "", false
}

func runWindowsPowerShell(script string) (string, int, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	output, err := cmd.CombinedOutput()
	out := decodeWindowsOutput(output)
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		return out, exitCode, err
	}
	return out, exitCode, nil
}

var delTargetRe = regexp.MustCompile(`(?i)^\s*(?:del|erase)\s+(?:/[^\s]+\s+)*("([^"]+)"|'([^']+)'|([^\s]+))`)

func parseWindowsDelTarget(cmd string) (string, bool) {
	m := delTargetRe.FindStringSubmatch(cmd)
	if len(m) < 5 {
		return "", false
	}
	for _, g := range []int{2, 3, 4} {
		v := strings.TrimSpace(m[g])
		if v != "" {
			return v, true
		}
	}
	return "", false
}

func escapePSSingleQuoted(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
