package db

import (
	"app/config"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DatabaseService struct {
	DB *sqlx.DB
}

type Repo struct {
	Id        int    `db:"id"`
	ChatID    int64  `db:"chat_id"`
	Host      string `db:"host"`
	Owner     string `db:"owner"`
	Repo      string `db:"repo"`
	LastTag   string `db:"last_tag"`
	IsRelease bool   `db:"is_release"`
}

var DBInstance *sqlx.DB

func (dbService *DatabaseService) PrepareDb() error {
	_, err := dbService.DB.Exec("")
	return err
}

func (dbService *DatabaseService) GetDatabaseUrl(cfg *config.AppConfig) string {
	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
	)

	return databaseUrl
}

func (dbService *DatabaseService) Connect(cfg *config.AppConfig) error {
	db, err := sqlx.Connect("postgres", dbService.GetDatabaseUrl(cfg))
	if err != nil {
		return err
	}

	dbService.DB = db

	return nil
}
