package db

import (
	"app/config"
	"app/order"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type DatabaseService struct {
	Logger *zap.Logger
	DB     *sqlx.DB
	Cache  *cache.Cache
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

// The function create all required tables if not exists
func (dbService *DatabaseService) PrepareDb() error {
	_, err := dbService.DB.Exec(
		`
		CREATE TABLE IF NOT EXISTS orders (
			order_uid VARCHAR(50) PRIMARY KEY,
			track_number VARCHAR(50),
			entry VARCHAR(50),
			locale VARCHAR(10),
			internal_signature VARCHAR(255),
			customer_id VARCHAR(50),
			delivery_service VARCHAR(50),
			shardkey VARCHAR(10),
			sm_id INT,
			date_created TIMESTAMP WITH TIME ZONE,
			oof_shard VARCHAR(10)
		);
		CREATE TABLE IF NOT EXISTS delivery (
			id SERIAL PRIMARY KEY,
			order_uid VARCHAR(50) REFERENCES orders(order_uid),
			name VARCHAR(255),
			phone VARCHAR(50),
			zip VARCHAR(50),
			city VARCHAR(100),
			address VARCHAR(255),
			region VARCHAR(100),
			email VARCHAR(100)
		);
		CREATE TABLE IF NOT EXISTS payment (
			id SERIAL PRIMARY KEY,
			order_uid VARCHAR(50) REFERENCES orders(order_uid),
			transaction VARCHAR(50),
			request_id VARCHAR(50),
			currency VARCHAR(10),
			provider VARCHAR(50),
			amount DECIMAL,
			payment_dt BIGINT,
			bank VARCHAR(50),
			delivery_cost DECIMAL,
			goods_total DECIMAL,
			custom_fee DECIMAL
		);
		CREATE TABLE IF NOT EXISTS items (
			id SERIAL PRIMARY KEY,
			order_uid VARCHAR(50) REFERENCES orders(order_uid),
			chrt_id INT,
			track_number VARCHAR(50),
			price DECIMAL,
			rid VARCHAR(50),
			name VARCHAR(255),
			sale INT,
			size VARCHAR(50),
			total_price DECIMAL,
			nm_id INT,
			brand VARCHAR(100),
			status INT
		);
		`)
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
func (dbService *DatabaseService) insertItems(tx *sqlx.Tx, items *[]order.Item) error {
	item_query := `INSERT INTO items (
		order_uid,
		chrt_id,
		track_number,
		price,
		rid,
		name,
		sale,
		size,
		total_price,
		nm_id,
		brand,
		status
	)
	VALUES (
		:order_uid,
		:chrt_id,
		:track_number,
		:price,
		:rid,
		:name,
		:sale,
		:size,
		:total_price,
		:nm_id,
		:brand,
		:status
	);`

	for _, item := range *items {
		dbService.Logger.Info(fmt.Sprintf("Item order_uid: %s", item.OrderUID))
		_, err := tx.NamedExec(item_query, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dbService *DatabaseService) insertPayment(tx *sqlx.Tx, payment *order.Payment) error {
	payment_query := `INSERT INTO payment (
		order_uid,
		transaction,
		request_id,
		currency,
		provider,
		amount,
		payment_dt,
		bank,
		delivery_cost,
		goods_total,
		custom_fee
	)
	VALUES (
		:order_uid,
		:transaction,
		:request_id,
		:currency,
		:provider,
		:amount,
		:payment_dt,
		:bank,
		:delivery_cost,
		:goods_total,
		:custom_fee
	);`

	dbService.Logger.Info(fmt.Sprintf("Payment order_uid: %s", payment.OrderUID))
	_, err := tx.NamedExec(payment_query, payment)

	return err
}
func (dbService *DatabaseService) insertDelivery(tx *sqlx.Tx, delivery *order.Delivery) error {
	delivery_query := `INSERT INTO delivery (
		order_uid,
		name,
		phone,
		zip,
		city,
		address,
		region,
		email
	)
	VALUES (
		:order_uid,
		:name,
		:phone,
		:zip,
		:city,
		:address,
		:region,
		:email
	);`

	dbService.Logger.Info(fmt.Sprintf("Delivery order_uid: %s", delivery.OrderUID))
	_, err := tx.NamedExec(delivery_query, delivery)

	return err
}

func (dbService *DatabaseService) insertOrder(tx *sqlx.Tx, order *order.Order) error {
	orders_query := `INSERT INTO orders (
		order_uid,
		track_number,
		entry,
		locale,
		internal_signature,
		customer_id,
		delivery_service,
		shardkey,
		sm_id,
		date_created,
		oof_shard
	)
	VALUES (
		:order_uid,
		:track_number,
		:entry,
		:locale,
		:internal_signature,
		:customer_id,
		:delivery_service,
		:shardkey,
		:sm_id,
		:date_created,
		:oof_shard
	);`

	_, err := tx.NamedExec(orders_query, order)
	return err
}

func (dbService *DatabaseService) AddOrder(tx *sqlx.Tx, order *order.Order) error {
	err := dbService.insertOrder(tx, order)
	if err != nil {
		dbService.Logger.Error("An error occured while trying to insert order")
		return err
	}
	dbService.Logger.Info(fmt.Sprintf("Successfully insert order order_uid:%s", order.OrderUID))

	err = dbService.insertItems(tx, &order.Items)
	if err != nil {
		dbService.Logger.Error("An error occured while trying to insert items")
		return err
	}
	dbService.Logger.Info(fmt.Sprintf("Successfully insert items order_uid:%s", order.OrderUID))

	err = dbService.insertDelivery(tx, &order.Delivery)
	if err != nil {
		dbService.Logger.Error("An error occured while trying to insert delivery")
		return err
	}
	dbService.Logger.Info(fmt.Sprintf("Successfully insert delivery order_uid:%s", order.Delivery.OrderUID))

	err = dbService.insertPayment(tx, &order.Payment)
	if err != nil {
		dbService.Logger.Error("An error occured while trying to insert payment")
		return err
	}
	dbService.Logger.Info(fmt.Sprintf("Successfully insert payment order_uid:%s", order.OrderUID))

	return nil
}
