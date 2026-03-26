package sessiontrace

import "testing"

func TestTrunc(t *testing.T) {
	if got := Trunc("abc", 2); got != "ab" {
		t.Fatalf("expected ab, got %q", got)
	}
	if got := Trunc("ab", 2); got != "ab" {
		t.Fatalf("expected ab, got %q", got)
	}
	if got := Trunc("ab", 0); got != "ab" {
		t.Fatalf("expected ab, got %q", got)
	}
}

func TestMaskEmail(t *testing.T) {
	if got := MaskEmail(""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if got := MaskEmail("a@b.com"); got != "***@b.com" {
		t.Fatalf("unexpected: %q", got)
	}
	if got := MaskEmail("abcd@b.com"); got != "ab***@b.com" {
		t.Fatalf("unexpected: %q", got)
	}
}

