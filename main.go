package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	forum "github.com/alijnmerchant21/forum-updated/abci"
	db "github.com/cometbft/cometbft-db"
	abciserver "github.com/cometbft/cometbft/abci/server"
)

func main() {
	// Create a new database
	db, err := db.NewGoLevelDB("forum-db", ".")
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Create a new ForumApp instance
	dbPath := "forum-db"
	app, err := forum.NewForumApp(dbPath)
	if err != nil {
		log.Fatalf("failed to create ForumApp instance: %v", err)
	}

	// Start the ABCI server
	socketAddr := "tcp://127.0.0.1:26658"
	server := abciserver.NewSocketServer(socketAddr, app)
	if err := server.Start(); err != nil {
		log.Fatalf("failed to start ABCI server: %v", err)
	}
	defer server.Stop()

	// Wait for a termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Gracefully stop the ABCI server
	if err := server.Stop(); err != nil {
		log.Fatalf("failed to stop ABCI server: %v", err)
	}

	fmt.Println("Forum application stopped")
}
