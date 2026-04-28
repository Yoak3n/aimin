package tool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/Yoak3n/aimin/hand/sandbox"
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

		detachProvided := strings.TrimSpace(firstNonEmpty(args["detach"], args["background"], args["bg"])) != ""
		detach := false
		if detachProvided {
			detach = parseBool(strings.TrimSpace(firstNonEmpty(args["detach"], args["background"], args["bg"])), false)
		} else {
			detach = shouldAutoDetachWindows(commandStr)
		}

		timeout := parseDurationSeconds(firstNonEmpty(args["timeout_s"], args["timeout"], args["t"]), 0)
		if timeout <= 0 && isLikelyAgentBrowserCommand(commandStr) {
			timeout = 90 * time.Second
		}
		base := ctx.Ctx
		if base == nil {
			base = context.Background()
		}

		if detach {
			out, pid, sbID, exitCode, err := runWindowsDetachedStart(ctx, base, commandStr, timeout)
			if err != nil {
				if fallback, ok := windowsFallback(commandStr); ok {
					return fallback
				}
				if exitCode == 0 {
					return fmt.Sprintf("ERROR: command execution failed: %s\nOutput: %s", err, out)
				}
				return fmt.Sprintf("ERROR: command execution finished with error\nExit Code: %d\nOutput: %s", exitCode, out)
			}
			if pid > 0 && sbID != "" {
				return fmt.Sprintf("command execution started (detached)\nSandbox ID: %s\nPID: %d\nOutput: %s", sbID, pid, out)
			}
			return fmt.Sprintf("command execution started (detached)\nOutput: %s", out)
		}

		cctx, cancel := withTimeoutOrCancel(base, timeout)
		defer cancel()

		out, exitCode, err := runWindowsCmdSandboxed(ctx, cctx, commandStr, cancel)
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
		timeout := parseDurationSeconds(firstNonEmpty(args["timeout_s"], args["timeout"], args["t"]), 0)
		base := ctx.Ctx
		if base == nil {
			base = context.Background()
		}
		cctx, cancel := withTimeoutOrCancel(base, timeout)
		defer cancel()

		out, exitCode, err := runUnixCmdSandboxed(ctx, cctx, commandStr, cancel)
		if err != nil {
			if exitCode == 0 {
				return fmt.Sprintf("ERROR: command execution failed: %s\nOutput: %s", err, out)
			}
			return fmt.Sprintf("ERROR: command execution finished with error\nExit Code: %d\nOutput: %s", exitCode, out)
		}
		return fmt.Sprintf("command execution success\nExit Code: %d\nOutput: %s", exitCode, out)
	default:
		return fmt.Sprintf("ERROR: unsupported os type: %s", osType)
	}
}

func emitProgress(ctx *Context, s string) {
	if ctx == nil || ctx.OnProgress == nil {
		return
	}
	ctx.OnProgress(s)
}

func withTimeoutOrCancel(base context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if base == nil {
		base = context.Background()
	}
	if timeout > 0 {
		return context.WithTimeout(base, timeout)
	}
	return context.WithCancel(base)
}

func runWindowsCmdSandboxed(ctx *Context, cctx context.Context, commandStr string, cancel context.CancelFunc) (string, int, error) {
	cmdLine := "chcp 65001>nul & " + commandStr
	cmd := exec.CommandContext(cctx, "cmd", "/D", "/C", cmdLine)

	sbID := ""
	if ctx != nil && ctx.Sandbox != nil {
		sbID = ctx.Sandbox.NewSandboxID()
	}
	pid := 0
	registered := false
	out, exitCode, _, err := runCmdWithProgress(
		cctx,
		cmd,
		func(p int) {
			pid = p
			if sbID == "" || ctx == nil || ctx.Sandbox == nil || p <= 0 {
				return
			}
			ctx.Sandbox.Register(&sandbox.Proc{
				ID:         sbID,
				RunID:      ctx.RunID,
				ToolCallID: ctx.ToolCallID,
				Action:     ctx.Action,
				PID:        p,
				StartedAt:  time.Now(),
				Cancel:     cancel,
			})
			registered = true
		},
		func(elapsed time.Duration) {
			if sbID != "" {
				emitProgress(ctx, fmt.Sprintf("running... sandbox_id=%s pid=%d elapsed=%s", sbID, pid, elapsed.Truncate(time.Second)))
				return
			}
			emitProgress(ctx, fmt.Sprintf("running... pid=%d elapsed=%s", pid, elapsed.Truncate(time.Second)))
		},
	)
	if registered {
		defer ctx.Sandbox.Unregister(sbID)
	}
	decoded := decodeWindowsOutput(out)
	if sbID != "" && pid > 0 {
		decoded = fmt.Sprintf("Sandbox ID: %s\nPID: %d\n%s", sbID, pid, decoded)
	}
	return decoded, exitCode, err
}

func runUnixCmdSandboxed(ctx *Context, cctx context.Context, commandStr string, cancel context.CancelFunc) (string, int, error) {
	cmd := exec.CommandContext(cctx, "sh", "-c", commandStr)

	sbID := ""
	if ctx != nil && ctx.Sandbox != nil {
		sbID = ctx.Sandbox.NewSandboxID()
	}
	pid := 0
	registered := false
	out, exitCode, _, err := runCmdWithProgress(
		cctx,
		cmd,
		func(p int) {
			pid = p
			if sbID == "" || ctx == nil || ctx.Sandbox == nil || p <= 0 {
				return
			}
			ctx.Sandbox.Register(&sandbox.Proc{
				ID:         sbID,
				RunID:      ctx.RunID,
				ToolCallID: ctx.ToolCallID,
				Action:     ctx.Action,
				PID:        p,
				StartedAt:  time.Now(),
				Cancel:     cancel,
			})
			registered = true
		},
		func(elapsed time.Duration) {
			if sbID != "" {
				emitProgress(ctx, fmt.Sprintf("running... sandbox_id=%s pid=%d elapsed=%s", sbID, pid, elapsed.Truncate(time.Second)))
				return
			}
			emitProgress(ctx, fmt.Sprintf("running... pid=%d elapsed=%s", pid, elapsed.Truncate(time.Second)))
		},
	)
	if registered {
		defer ctx.Sandbox.Unregister(sbID)
	}
	s := string(out)
	if sbID != "" && pid > 0 {
		s = fmt.Sprintf("Sandbox ID: %s\nPID: %d\n%s", sbID, pid, s)
	}
	return s, exitCode, err
}

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (w *lockedBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (w *lockedBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]byte(nil), w.b.Bytes()...)
}

func runCmdWithProgress(ctx context.Context, cmd *exec.Cmd, onStart func(pid int), onTick func(elapsed time.Duration)) ([]byte, int, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, 0, 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, 0, 0, err
	}

	var buf lockedBuffer
	if err := cmd.Start(); err != nil {
		return nil, 0, 0, err
	}

	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	if onStart != nil {
		onStart(pid)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&buf, stdout)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&buf, stderr)
	}()

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	start := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var waitErr error
	for {
		select {
		case waitErr = <-waitCh:
			goto done
		case <-ticker.C:
			if onTick != nil {
				onTick(time.Since(start))
			}
		case <-ctx.Done():
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			waitErr = <-waitCh
			goto done
		}
	}

done:
	wg.Wait()
	out := buf.Bytes()

	exitCode := 0
	if waitErr != nil {
		if exitError, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		return out, exitCode, pid, waitErr
	}

	return out, exitCode, pid, nil
}

func runWindowsDetachedStart(ctx *Context, base context.Context, commandStr string, timeout time.Duration) (string, int, string, int, error) {
	if base == nil {
		base = context.Background()
	}
	cctx, cancel := withTimeoutOrCancel(base, timeout)
	defer cancel()

	if isSimpleWindowsExec(commandStr) {
		file, args := splitWindowsExec(commandStr)
		if file != "" {
			out, pid, sbID, exitCode, err := runWindowsStartProcessDetached(ctx, cctx, file, args)
			if err == nil || out != "" || pid > 0 {
				return out, pid, sbID, exitCode, err
			}
		}
	}

	cmdLine := "chcp 65001>nul & start \"\" " + commandStr
	cmd := exec.CommandContext(cctx, "cmd", "/D", "/C", cmdLine)
	output, err := cmd.CombinedOutput()
	out := decodeWindowsOutput(output)
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		return out, 0, "", exitCode, err
	}
	return out, 0, "", 0, nil
}

func isSimpleWindowsExec(commandStr string) bool {
	s := strings.TrimSpace(commandStr)
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, "|&<>") {
		return false
	}
	ls := strings.ToLower(s)
	if strings.HasPrefix(ls, "cmd ") || strings.HasPrefix(ls, "powershell ") || strings.HasPrefix(ls, "pwsh ") {
		return false
	}
	return true
}

func splitWindowsExec(commandStr string) (string, []string) {
	toks := tokenizeWindowsCommand(commandStr)
	if len(toks) == 0 {
		return "", nil
	}
	return toks[0], toks[1:]
}

func tokenizeWindowsCommand(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	var b strings.Builder
	inDouble := false
	escape := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escape {
			b.WriteByte(ch)
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inDouble = !inDouble
			continue
		}
		if !inDouble && (ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n') {
			if b.Len() > 0 {
				out = append(out, b.String())
				b.Reset()
			}
			continue
		}
		b.WriteByte(ch)
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

func runWindowsStartProcessDetached(ctx *Context, cctx context.Context, file string, args []string) (string, int, string, int, error) {
	sbID := ""
	if ctx != nil && ctx.Sandbox != nil {
		sbID = ctx.Sandbox.NewSandboxID()
	}
	argExpr := "@()"
	if len(args) > 0 {
		parts := make([]string, 0, len(args))
		for _, a := range args {
			parts = append(parts, "'"+escapePSSingleQuoted(a)+"'")
		}
		argExpr = "@(" + strings.Join(parts, ",") + ")"
	}

	ps := fmt.Sprintf("$p=Start-Process -FilePath '%s' -ArgumentList %s -PassThru; Write-Output $p.Id", escapePSSingleQuoted(file), argExpr)
	out, exitCode, err := runWindowsPowerShellCtx(cctx, ps)
	if err != nil {
		return out, 0, "", exitCode, err
	}

	pid := 0
	for _, f := range strings.Fields(out) {
		n, e := strconv.Atoi(strings.TrimSpace(f))
		if e == nil && n > 0 {
			pid = n
			break
		}
	}
	if pid > 0 && sbID != "" && ctx != nil && ctx.Sandbox != nil {
		ctx.Sandbox.Register(&sandbox.Proc{
			ID:         sbID,
			RunID:      ctx.RunID,
			ToolCallID: ctx.ToolCallID,
			Action:     ctx.Action,
			PID:        pid,
			StartedAt:  time.Now(),
		})
	}
	return out, pid, sbID, 0, nil
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
	return runWindowsPowerShellCtx(context.Background(), script)
}

func runWindowsPowerShellCtx(ctx context.Context, script string) (string, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", script)
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

func shouldAutoDetachWindows(commandStr string) bool {
	ls := strings.ToLower(strings.TrimSpace(commandStr))
	if strings.HasPrefix(ls, "start ") {
		return false
	}
	if strings.Contains(ls, "--remote-debugging-port=") {
		if strings.Contains(ls, "chrome") || strings.Contains(ls, "msedge") || strings.Contains(ls, "edge") {
			return true
		}
	}
	if strings.HasPrefix(ls, "chrome.exe") || strings.HasPrefix(ls, "msedge.exe") || strings.HasPrefix(ls, "edge.exe") {
		return true
	}
	return false
}

func isLikelyAgentBrowserCommand(commandStr string) bool {
	ls := strings.ToLower(strings.TrimSpace(commandStr))
	return strings.HasPrefix(ls, "agent-browser ") || ls == "agent-browser"
}
