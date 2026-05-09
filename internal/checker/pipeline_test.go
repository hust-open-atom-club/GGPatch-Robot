package checker

import (
	"reflect"
	"testing"
)

func TestParseChangedPaths(t *testing.T) {
	body := `diff --git a/fs/ext4/inode.c b/fs/ext4/inode.c
--- a/fs/ext4/inode.c
+++ b/fs/ext4/inode.c
@@ -1234,6 +1234,7 @@
diff --git a/fs/ext4/super.c b/fs/ext4/super.c
--- a/fs/ext4/super.c
+++ b/fs/ext4/super.c`

	paths := ParseChangedPaths(body)
	expected := []string{"fs/ext4/inode.c", "fs/ext4/super.c"}
	if !reflect.DeepEqual(paths, expected) {
		t.Errorf("got %v, want %v", paths, expected)
	}
}

func TestParseChangedPathsEmpty(t *testing.T) {
	paths := ParseChangedPaths("just a commit message\nno diffs here\n")
	if len(paths) != 0 {
		t.Errorf("expected empty, got %v", paths)
	}
}

func TestBuildReport(t *testing.T) {
	paths := []string{"fs/ext4/inode.c"}
	logMsg := "Fix null deref in ext4"
	results := []Result{
		{Name: "CheckPatchPl", Passed: true, Detail: "*** CheckPatch PASS ***\n"},
		{Name: "ApplyCheck", Passed: true, Detail: "linux-next"},
		{Name: "BuildCheck", Passed: true, Detail: "*** BuildCheck PASS ***\n"},
	}

	report := BuildReport(paths, logMsg, "linux-next", results)
	if report == "" {
		t.Error("expected non-empty report")
	}
	if len(splitLines(report)) < 3 {
		t.Error("expected report with multiple lines")
	}
}
