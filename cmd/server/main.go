package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ByteCrafter-rgb/kv-store/internal/server"
)

func main() {
	port := flag.Int("port", 6379, "port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	srv := server.NewServer(addr)

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down server...")
	srv.Stop()
}
