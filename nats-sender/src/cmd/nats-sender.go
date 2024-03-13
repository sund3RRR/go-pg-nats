package main

import (
	"flag"
	"fmt"
	"log"
	"main/config"
	"os"
	"path/filepath"

	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

func displayHelp() {
	fmt.Println("Usage: utility_name -c <config_file_path> -f <data_file_path>")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func parseCmd() (*string, *string) {
	configFlag := flag.String("c", "", "Path to the config file (yaml format)")
	fileFlag := flag.String("f", "", "Path to the data file (json format)")

	flag.Parse()

	helpFlag := flag.Bool("help", false, "Display help message")
	hFlag := flag.Bool("h", false, "Display help message")
	flag.Parse()

	if *helpFlag || *hFlag {
		displayHelp()
		os.Exit(0)
	}

	if *fileFlag == "" {
		fmt.Println("You must specify a path to the JSON file using the -f flag")
		displayHelp()
		os.Exit(1)
	}

	return configFlag, fileFlag
}
func main() {
	configFlag, fileFlag := parseCmd()

	executable, err := os.Executable()
	if err != nil {
		log.Fatal("Error getting executable path:", err)
	}

	dir := filepath.Dir(executable)

	configPath := filepath.Join(dir, "../config/config.yml")
	if *configFlag != "" {
		configPath = *configFlag
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := cfg.ZapConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	natsURL := fmt.Sprintf("nats://%s:%d", cfg.NatsStreaming.Host, cfg.NatsStreaming.Port)
	sc, err := stan.Connect(cfg.NatsStreaming.Cluster, cfg.NatsStreaming.Client, stan.NatsURL(natsURL))
	if err != nil {
		logger.Fatal("An error occured while trying to connect to NATS-streaming", zap.Error(err))
	}
	defer sc.Close()

	jsonFileContent, err := os.ReadFile(*fileFlag)
	if err != nil {
		logger.Fatal("An error occured while trying to read file", zap.Error(err))
	}

	err = sc.Publish(cfg.NatsStreaming.Channel, jsonFileContent)
	if err != nil {
		logger.Fatal("An error occured while trying to send content to channel", zap.Error(err))
	}

	fmt.Printf("JSON file content successfully sent to channel '%s'", cfg.NatsStreaming.Channel)
}
