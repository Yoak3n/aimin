package mcp

import "testing"

func TestExecute_ShellCommand_FailedExitCodeReturnsError(t *testing.T) {
	h := NewMcpHUB()
	h.RegisterTool(ShellCommandTool())

	_, err := h.Execute("ShellCommand(windows,exit /b 7)")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestExecute_ShellCommand_SuccessReturnsNoError(t *testing.T) {
	h := NewMcpHUB()
	h.RegisterTool(ShellCommandTool())

	_, err := h.Execute("ShellCommand(windows,exit /b 0)")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
