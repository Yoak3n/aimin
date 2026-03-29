package helper

import "testing"

func TestParseFunctionCall_MultilineArgs(t *testing.T) {
	in := "FileOperation(Write,IDENTITY.md,# IDENTITY.md - 我是谁\n\n- a\n- b\n)"
	name, args, err := ParseFunctionCall(in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if name != "FileOperation" {
		t.Fatalf("expected name FileOperation, got %s", name)
	}
	if args == "" {
		t.Fatalf("expected args not empty")
	}
}

func TestParseFunctionCall_MissingClosingParen(t *testing.T) {
	in := "FileOperation(Write,IDENTITY.md,# IDENTITY.md - 我是谁\n\n- a\n- b"
	name, args, err := ParseFunctionCall(in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if name != "FileOperation" {
		t.Fatalf("expected name FileOperation, got %s", name)
	}
	if args == "" {
		t.Fatalf("expected args not empty")
	}
}

func TestParseFunctionCall_LeadingJunkAndTrailingText(t *testing.T) {
	in := "。\n请执行：FileOperation(Read,default_workspace/SOUL.md)\n谢谢"
	name, args, err := ParseFunctionCall(in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if name != "FileOperation" {
		t.Fatalf("expected name FileOperation, got %s", name)
	}
	if args == "" {
		t.Fatalf("expected args not empty")
	}
}

func TestExtractContentByTag_UnclosedTag(t *testing.T) {
	in := "<action>FileOperation(Read,a/b)\n"
	got := ExtractContentByTag(in, "action")
	if got == "" {
		t.Fatalf("expected content, got empty")
	}
}
