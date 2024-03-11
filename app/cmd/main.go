package main

import (
	"app/config"
	"app/db"
	"log"
	"sync"

	_ "github.com/lib/pq"

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
