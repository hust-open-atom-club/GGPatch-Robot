package mail

import (
	"context"
)

// RawEmail holds parsed email data extracted from the IMAP stream.
type RawEmail struct {
	FromName  string
	FromAddr  string
	Subject   string
	MessageID string
	Date      string // RFC3339 format

	Addresses struct {
		To []string
		Cc []string
	}

	Header map[string][]string
	Body   []byte
}

// Receiver fetches new patch emails.
// Define interface at the call site (engine) — this file documents the
// return type used across the package boundary.
type Receiver interface {
	Receive(ctx context.Context) ([]RawEmail, error)
}
