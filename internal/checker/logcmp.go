package checker

import "strings"

// LogDiff holds the categorized differences between before and after logs.
type LogDiff struct {
	UnsolvedWarnings []string
	UnsolvedErrors   []string
	NewWarnings      []string
	NewErrors        []string
}

type spf struct {
	line string
	full string
}

// Logcmp compares pre-patch and post-patch checker output, classifying
// warnings and errors into unsolved (still present) and new (introduced).
func Logcmp(pre, after, warnKeyword, errKeyword string) LogDiff {
	preLines := splitLines(pre)
	afterLines := splitLines(after)

	makeSpf := func(lines []string, keyword string) []spf {
		var out []spf
		for _, l := range lines {
			if !strings.Contains(l, keyword) {
				continue
			}
			end1 := strings.Index(l, ":")
			linePre := l[:end1]
			start2 := strings.Index(l, " ")
			var end2 int
			if strings.Contains(l, "on lines:") {
				end2 = strings.LastIndex(l, ":")
			} else {
				end2 = len(l)
			}
			lineAfter := l[start2+1 : end2]
			out = append(out, spf{line: linePre + lineAfter, full: l})
		}
		return out
	}

	preWarnSpf := makeSpf(preLines, warnKeyword)
	preErrSpf := makeSpf(preLines, errKeyword)
	afterWarnSpf := makeSpf(afterLines, warnKeyword)
	afterErrSpf := makeSpf(afterLines, errKeyword)

	preWarnSet := spfSet(preWarnSpf)
	preErrSet := spfSet(preErrSpf)

	var diff LogDiff

	for _, a := range afterWarnSpf {
		if preWarnSet[a.line] {
			diff.UnsolvedWarnings = append(diff.UnsolvedWarnings, a.full)
		} else {
			diff.NewWarnings = append(diff.NewWarnings, a.full)
		}
	}
	for _, a := range afterErrSpf {
		if preErrSet[a.line] {
			diff.UnsolvedErrors = append(diff.UnsolvedErrors, a.full)
		} else {
			diff.NewErrors = append(diff.NewErrors, a.full)
		}
	}
	return diff
}

func spfSet(items []spf) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[it.line] = true
	}
	return m
}
