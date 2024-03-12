package main

import (
	"app/config"
	"app/db"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
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

	dbService := db.DatabaseService{}
	err = dbService.Connect(cfg)
	if err != nil {
		logger.Fatal(
			"An error occured while trying to connect postgreSQL",
			zap.Error(err),
			zap.String("Host", cfg.Postgres.Host),
			zap.Int("Port", cfg.Postgres.Port),
			zap.String("User", cfg.Postgres.User),
			zap.String("Database", cfg.Postgres.Database),
		)
	}

	defer dbService.DB.Close()
	logger.Info("Successfully connected to PostgeSQL")

	err = dbService.PrepareDb()
	if err != nil {
		logger.Fatal("An error occured while trying to prepare DB", zap.Error(err))
	}

	natsURL := fmt.Sprintf("nats://%s:%d", cfg.NatsStreaming.Host, cfg.NatsStreaming.Port)
	sc, err := stan.Connect(cfg.NatsStreaming.Cluster, cfg.NatsStreaming.Client, stan.NatsURL(natsURL))
	if err != nil {
		logger.Fatal("An error occured while trying to connect to NATS-streaming", zap.Error(err))
	}
	defer sc.Close()

	logger.Info("Successfully connected to NATS-streaming")

	sub, err := sc.Subscribe(cfg.NatsStreaming.Channel, func(msg *stan.Msg) {
		log.Printf("Received a message: %s\n", string(msg.Data))
	})
	if err != nil {
		logger.Fatal("An error occured while trying to subscribe NATS channel", zap.Error(err))
	}
	defer sub.Unsubscribe()

	logger.Info(fmt.Sprintf("Successfully subscribed to NATS-streaming channel %s", cfg.NatsStreaming.Channel))

	logger.Info("Starting server...")

	var wg sync.WaitGroup
	wg.Add(3)

	// botService := bot.BotService{
	// 	DatabaseService: &dbService,
	// 	Logger:          logger,
	// }

	// go func() {
	// 	defer wg.Done()
	// 	botService.StartBot()
	// }()

	// go func() {
	// 	defer wg.Done()
	// 	botService.StartRepoSender()
	// }()

	wg.Wait()
}
