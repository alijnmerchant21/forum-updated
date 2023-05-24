package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	"github.com/spf13/viper"

	forum "github.com/alijnmerchant21/forum-updated/abci"
	db "github.com/cometbft/cometbft-db"
	cfg "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/privval"
)

var homeDir string

func init() {
	flag.StringVar(&homeDir, "cmt-home", "", "Path to the CometBFT config directory (if empty, uses $HOME/.cometbft)")
}

func main() {
	flag.Parse()
	if homeDir == "" {
		homeDir = os.ExpandEnv("$HOME/.cometbft")
	}

	config := cfg.DefaultConfig()
	config.SetRoot(homeDir)
	viper.SetConfigFile(fmt.Sprintf("%s/%s", homeDir, "config.toml"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	db, err := db.NewGoLevelDB(filepath.Join(homeDir, "forum-db"), ".")
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	dbPath := "forum-db"
	app, err := forum.NewForumApp(dbPath)

	if err != nil {
		log.Fatalf("failed to create ForumApp instance: %v", err)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel)

	if err != nil {
		log.Fatalf("failed to read genesis doc: %v", err)
	}

	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		log.Fatalf("failed to load node key: %v", err)
	}

	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	node, err := nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger,
	)

	if err != nil {
		log.Fatalf("failed to create CometBFT node: %v", err)
	}

	if err := node.Start(); err != nil {
		log.Fatalf("failed to start CometBFT node: %v", err)
	}
	defer func() {
		node.Stop()
		node.Wait()
	}()

	httpAddr := "127.0.0.1:8080"
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		// Extract the public key from the request URL
		pubkey := r.URL.Query().Get("pubkey")
		if pubkey == "" {
			http.Error(w, "missing pubkey parameter", http.StatusBadRequest)
			return
		}

		// Retrieve the messages for the given public key
		pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkey)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decode pubkey: %v", err), http.StatusBadRequest)
			return
		}
		messages, err := app.DB.GetMessagesByPubKey(pubkeyBytes)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get messages: %v", err), http.StatusInternalServerError)
			return
		}

		// Marshal the messages to JSON and send as the response
		respBytes, err := json.Marshal(messages)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal messages: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	})

	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	if err := node.Stop(); err != nil {
		log.Fatalf("failed to stop CometBFT node: %v", err)
	}

	fmt.Println("Forum application stopped")
}
