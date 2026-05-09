package mail

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strconv"
)

type smtpSender struct {
	server   string
	port     int
	username string
	password string
}

func NewSMTPSender(server string, port int, username, password string) *smtpSender {
	return &smtpSender{
		server:   server,
		port:     port,
		username: username,
		password: password,
	}
}

func (s *smtpSender) Send(ctx context.Context, to []string, subject, body string) error {
	addr := s.server + ":" + strconv.Itoa(s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.server)
	msg := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)
	slog.Debug("sending email", "to", to, "addr", addr)
	if err := smtp.SendMail(addr, auth, s.username, to, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
