package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrInvalidEmail  = errors.New("invalid email format, expected user@domain")
	ErrUnknownDomain = errors.New("unsupported email domain")
)

type Config struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Procs       int      `json:"procs"`
	Interval    int      `json:"interval"`
	WhiteLists  []string `json:"whiteLists"`
	MailingList string   `json:"mailingList"`
}

type SMTP struct {
	Server   string
	Port     int
	Username string
	Password string
}

type IMAP struct {
	Server   string
	Port     int
	Username string
	Password string
}

type MailInfo struct {
	SMTP        SMTP
	IMAP        IMAP
	WhiteLists  []string
	Procs       int
	Interval    int
	MailingList string
}

type domainConfig struct {
	SMTP string
	Port int
	IMAP string
}

var domains = map[string]domainConfig{
	"126.com":     {SMTP: "smtp.126.com", Port: 25, IMAP: "imap.126.com"},
	"hust.edu.cn": {SMTP: "mail.hust.edu.cn", Port: 465, IMAP: "mail.hust.edu.cn"},
}

func Load(path string) (*MailInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if cfg.Procs == 0 {
		cfg.Procs = 20
	}
	if cfg.Interval == 0 {
		cfg.Interval = 20
	}
	if cfg.MailingList == "" {
		cfg.MailingList = "kernel_testing_robot@googlegroups.com"
	}

	parts := strings.Split(cfg.Username, "@")
	if len(parts) != 2 {
		return nil, ErrInvalidEmail
	}
	dc, ok := domains[parts[1]]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownDomain, parts[1])
	}

	return &MailInfo{
		SMTP: SMTP{
			Server:   dc.SMTP,
			Port:     dc.Port,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		IMAP: IMAP{
			Server:   dc.IMAP,
			Port:     993,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		WhiteLists:  cfg.WhiteLists,
		Procs:       cfg.Procs,
		Interval:    cfg.Interval,
		MailingList: cfg.MailingList,
	}, nil
}
