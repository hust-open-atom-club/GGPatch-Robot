package mail

import (
	"fmt"
	"strings"
)

// Headers captures the original email metadata for the reply.
type Headers struct {
	FromName  string
	FromAddr  string
	MessageID string
	CCList    []string
}

// PatchExtract extracts the patch body from the raw email, filters out
// quoted-list footers, and writes a .patch file.
// Returns the patch filename and metadata for the reply.
func PatchExtract(raw RawEmail, whitelist func(addr string) bool) (patchName string, body string, h Headers, ok bool) {
	h.FromName = raw.FromName
	h.FromAddr = raw.FromAddr
	h.MessageID = raw.MessageID

	// Check all recipients against whitelist; ignore if any fail.
	for _, addr := range append(raw.Addresses.To, raw.Addresses.Cc...) {
		if !whitelist(addr) {
			return "", "", Headers{}, false
		}
		h.CCList = append(h.CCList, addr)
	}

	text := string(raw.Body)
	if strings.Contains(text, "Reviewed-by:") {
		return "", "", Headers{}, false
	}

	lines := strings.Split(text, "\n")
	var changedPaths []string
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git a") {
			sub := line[13:]
			idx := strings.Index(sub, " ")
			changedPaths = append(changedPaths, sub[:idx])
		}
	}
	if len(changedPaths) == 0 {
		return "", "", Headers{}, false
	}

	// Strip the mailing-list footer.
	if idx := strings.Index(text, "You received this message because"); idx != -1 {
		text = text[:idx-5]
	}

	name := strings.ReplaceAll(raw.FromName, " ", "")
	patchName = fmt.Sprintf("%s_%s.patch", name, raw.Date)

	return patchName, text, h, true
}
