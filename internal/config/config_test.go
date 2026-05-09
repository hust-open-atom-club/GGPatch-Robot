package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"username": "test@126.com", "password": "secret"}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Procs != 20 {
		t.Errorf("expected procs=20, got %d", info.Procs)
	}
	if info.Interval != 20 {
		t.Errorf("expected interval=20, got %d", info.Interval)
	}
	if info.MailingList != "kernel_testing_robot@googlegroups.com" {
		t.Errorf("expected default mailing list, got %q", info.MailingList)
	}
	if info.SMTP.Server != "smtp.126.com" {
		t.Errorf("expected smtp.126.com, got %q", info.SMTP.Server)
	}
	if info.IMAP.Port != 993 {
		t.Errorf("expected IMAP port 993, got %d", info.IMAP.Port)
	}
}

func TestLoadInvalidEmail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"username": "invalid", "password": "secret"}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestLoadUnknownDomain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"username": "test@unknown.com", "password": "secret"}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown domain")
	}
}

func TestLoadCustomValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"username": "test@hust.edu.cn",
		"password": "pw",
		"procs": 8,
		"interval": 5,
		"mailingList": "custom@list.com",
		"whiteLists": ["hust.edu.cn"]
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Procs != 8 {
		t.Errorf("procs: got %d", info.Procs)
	}
	if info.Interval != 5 {
		t.Errorf("interval: got %d", info.Interval)
	}
	if info.MailingList != "custom@list.com" {
		t.Errorf("mailingList: got %q", info.MailingList)
	}
	if info.SMTP.Server != "mail.hust.edu.cn" {
		t.Errorf("smtp server: got %q", info.SMTP.Server)
	}
}
