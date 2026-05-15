package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileOperation_Write_LargeContentWithCommasAndNewlines_NotTruncated(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	content := "line1\nline2,with,commas\nline3\n"
	ctx := NewMcpContext()
	ctx.SetPayload("write," + target + "," + content)

	res := FileOperation(ctx)
	if res != "write file success" {
		t.Fatalf("unexpected result: %s", res)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if string(got) != content {
		t.Fatalf("content mismatch:\n---want---\n%s\n---got---\n%s", content, string(got))
	}
}

func TestFileOperation_Write_ContentAsKeyValue_AllowsCommas(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	content := "a,b,c"
	ctx := NewMcpContext()
	ctx.SetPayload("op=write,path=" + target + ",content=\"" + content + "\"")

	res := FileOperation(ctx)
	if res != "write file success" {
		t.Fatalf("unexpected result: %s", res)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if string(got) != content {
		t.Fatalf("content mismatch: want=%q got=%q", content, string(got))
	}
}
