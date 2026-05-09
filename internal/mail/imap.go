package mail

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type imapReceiver struct {
	server   string
	port     int
	username string
	password string
}

func NewIMAPReceiver(server string, port int, username, password string) *imapReceiver {
	return &imapReceiver{
		server:   server,
		port:     port,
		username: username,
		password: password,
	}
}

func (r *imapReceiver) Receive(ctx context.Context) ([]RawEmail, error) {
	addr := fmt.Sprintf("%s:%d", r.server, r.port)
	slog.Debug("connecting to IMAP", "addr", addr)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("dial IMAP: %w", err)
	}
	defer c.Logout()

	idClient := id.NewClient(c)
	idClient.ID(id.ID{
		id.FieldName:    "ggpatch-robot",
		id.FieldVersion: "2.0.0",
	})

	if err := c.Login(r.username, r.password); err != nil {
		return nil, fmt.Errorf("imap login: %w", err)
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("select INBOX: %w", err)
	}
	if mbox.Recent == 0 {
		return nil, nil
	}

	from := mbox.Messages - mbox.Recent + 1
	to := mbox.Messages

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}

	messages := make(chan *imap.Message, to-from+2)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	var emails []RawEmail
	for msg := range messages {
		if msg.Envelope == nil || !strings.HasPrefix(msg.Envelope.Subject, "[PATCH") {
			continue
		}

		section, err := imap.ParseBodySectionName("BODY[]")
		if err != nil {
			slog.Warn("parse body section failed", "error", err)
			continue
		}
		r := msg.GetBody(section)
		if r == nil {
			continue
		}

		mr, err := mail.CreateReader(r)
		if err != nil {
			slog.Warn("create mail reader failed", "error", err)
			continue
		}

		raw, err := readRaw(mr)
		if err != nil {
			slog.Warn("read raw email failed", "error", err)
			continue
		}
		emails = append(emails, raw)
	}
	if err := <-done; err != nil {
		return nil, fmt.Errorf("imap fetch: %w", err)
	}

	return emails, nil
}

func readRaw(mr *mail.Reader) (RawEmail, error) {
	header := mr.Header
	var raw RawEmail

	if from, err := header.AddressList("From"); err == nil && len(from) > 0 {
		raw.FromName = from[0].Name
		if raw.FromName == "" {
			idx := strings.Index(from[0].Address, "@")
			raw.FromName = from[0].Address[:idx]
		}
		raw.FromAddr = from[0].Address
	}
	if subject, err := header.Subject(); err == nil {
		raw.Subject = subject
	}
	if msgid, err := header.MessageID(); err == nil {
		raw.MessageID = msgid
	}
	if date, err := header.Date(); err == nil {
		raw.Date = date.Format("20060102150405")
	}
	if cc, err := header.AddressList("Cc"); err == nil {
		for _, a := range cc {
			raw.Addresses.Cc = append(raw.Addresses.Cc, a.Address)
		}
	}
	if to, err := header.AddressList("To"); err == nil {
		for _, a := range to {
			raw.Addresses.To = append(raw.Addresses.To, a.Address)
		}
	}

	raw.Header = make(map[string][]string)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Debug("skip mail part", "error", err)
			continue
		}
		switch p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := io.ReadAll(p.Body)
			raw.Body = b
		}
	}
	return raw, nil
}
