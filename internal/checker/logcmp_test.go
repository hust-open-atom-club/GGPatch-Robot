package checker

import "testing"

func TestLogcmpNoChange(t *testing.T) {
	pre := "warning: unused var\nwarning: unused func\nerror: null deref\n"
	after := "warning: unused var\nwarning: unused func\nerror: null deref\n"

	diff := Logcmp(pre, after, "warning", "error")
	if len(diff.UnsolvedWarnings) != 2 {
		t.Errorf("expected 2 unsolved warnings, got %d", len(diff.UnsolvedWarnings))
	}
	if len(diff.UnsolvedErrors) != 1 {
		t.Errorf("expected 1 unsolved error, got %d", len(diff.UnsolvedErrors))
	}
	if len(diff.NewWarnings) > 0 || len(diff.NewErrors) > 0 {
		t.Errorf("expected no new issues, got warnings=%d errors=%d", len(diff.NewWarnings), len(diff.NewErrors))
	}
}

func TestLogcmpNewWarning(t *testing.T) {
	pre := "warning: unused var\n"
	after := "warning: unused var\nwarning: new thing\n"

	diff := Logcmp(pre, after, "warning", "error")
	if len(diff.UnsolvedWarnings) != 1 {
		t.Errorf("expected 1 unsolved warning, got %d", len(diff.UnsolvedWarnings))
	}
	if len(diff.NewWarnings) != 1 {
		t.Errorf("expected 1 new warning, got %d", len(diff.NewWarnings))
	}
}

func TestLogcmpFixedError(t *testing.T) {
	pre := "error: mem leak\nerror: null deref\n"
	after := "error: null deref\n"

	diff := Logcmp(pre, after, "warn", "error")
	if len(diff.UnsolvedErrors) != 1 {
		t.Errorf("expected 1 unsolved error, got %d", len(diff.UnsolvedErrors))
	}
	// The fixed error doesn't appear in 'after' at all, so it's not counted
}

func TestLogcmpMixedKeywords(t *testing.T) {
	pre := "WARNING: deprecated\nerror: crash\n"
	after := "WARNING: deprecated\nerror: crash\nWARNING: new dep\n"

	diff := Logcmp(pre, after, "WARNING", "error")
	if len(diff.UnsolvedWarnings) != 1 {
		t.Errorf("expected 1 unsolved warning, got %d", len(diff.UnsolvedWarnings))
	}
	if len(diff.NewWarnings) != 1 {
		t.Errorf("expected 1 new warning, got %d", len(diff.NewWarnings))
	}
	if len(diff.UnsolvedErrors) != 1 {
		t.Errorf("expected 1 unsolved error, got %d", len(diff.UnsolvedErrors))
	}
}

func TestLogcmpNoKeywords(t *testing.T) {
	pre := "just some text\nnothing here\n"
	after := "just some text\nnothing here\n"

	diff := Logcmp(pre, after, "warning", "error")
	if len(diff.UnsolvedWarnings) > 0 || len(diff.UnsolvedErrors) > 0 ||
		len(diff.NewWarnings) > 0 || len(diff.NewErrors) > 0 {
		t.Error("expected empty diff for text with no keywords")
	}
}
