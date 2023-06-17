package server

import "github.com/charmbracelet/ssh"

func authHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}
