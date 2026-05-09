package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ggpatch-robot/internal/checker"
	"ggpatch-robot/internal/config"
	"ggpatch-robot/internal/engine"
	"ggpatch-robot/internal/kernel"
	"ggpatch-robot/internal/mail"
)

func main() {
	configPath := flag.String("config", "", "path to config.json")
	flag.Parse()
	if *configPath == "" {
		slog.Error("no config file specified")
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	mgr := kernel.NewManager(cfg.Procs)

	recv := mail.NewIMAPReceiver(cfg.IMAP.Server, cfg.IMAP.Port, cfg.IMAP.Username, cfg.IMAP.Password)
	send := mail.NewSMTPSender(cfg.SMTP.Server, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)

	cocci := checker.NewCoccicheck(mgr)
	cppcheck := checker.NewCppcheck(mgr)

	eng := engine.New(cfg, recv, send, mgr, cocci, cppcheck)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := eng.Run(ctx); err != nil {
		slog.Error("engine stopped", "error", err)
		os.Exit(1)
	}
}
