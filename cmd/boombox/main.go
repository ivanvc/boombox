package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ivanvc/boombox/internal/config"
	"github.com/ivanvc/boombox/internal/server"
	k8s "github.com/ivanvc/boombox/internal/services/kubernetes"

	"github.com/charmbracelet/log"
)

func main() {
	cfg := config.LoadConfig()
	log.SetLevel(log.ParseLevel(cfg.LogLevel))
	client := k8s.LoadClient(cfg.Namespace)

	s := server.New(cfg, client)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Infof("Starting SSH server on %s", cfg.Listen)
	listenChan := make(chan error, 1)

	go func() {
		defer close(listenChan)
		defer close(done)
		listenChan <- s.ListenAndServe()
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	if err := <-listenChan; err != nil {
		log.Fatal(err)
	}
}
