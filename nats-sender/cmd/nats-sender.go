package main

import (
	"flag"
	"fmt"
	"log"
	"main/config"
	"os"

	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig("config/config.yml")
	if err != nil {
		log.Fatal(err)
	}

	logger, err := cfg.ZapConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	jsonFilePath := flag.String("f", "", "Path to the JSON file")
	flag.Parse()

	if *jsonFilePath == "" {
		logger.Fatal("You must specify a path to the JSON file using the -f flag")
	}

	natsURL := fmt.Sprintf("nats://%s:%d", cfg.NatsStreaming.Host, cfg.NatsStreaming.Port)
	sc, err := stan.Connect(cfg.NatsStreaming.Cluster, cfg.NatsStreaming.Client, stan.NatsURL(natsURL))
	if err != nil {
		logger.Fatal("An error occured while trying to connect to NATS-streaming", zap.Error(err))
	}
	defer sc.Close()

	jsonFileContent, err := os.ReadFile(*jsonFilePath)
	if err != nil {
		logger.Fatal("An error occured while trying to read file", zap.Error(err))
	}

	err = sc.Publish(cfg.NatsStreaming.Channel, jsonFileContent)
	if err != nil {
		logger.Fatal("An error occured while trying to send content to channel", zap.Error(err))
	}

	fmt.Printf("JSON file content successfully sent to channel '%s'", cfg.NatsStreaming.Channel)
}
