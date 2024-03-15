package db

import (
	"app/order"
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

func (dbService *DatabaseService) LoadCache() {
	var orders []order.Order

	err := dbService.DB.Select(&orders, "SELECT * FROM orders")
	if err != nil {
		dbService.Logger.Fatal("An error occured while trying to select orders", zap.Error(err))
	}

	for i := 0; i < len(orders); i++ {
		order_uid := orders[i].OrderUID

		tx, err := dbService.DB.BeginTxx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			dbService.Logger.Error("Error trying to begin transaction", zap.Error(err))
			return
		}

		payment, err := dbService.selectPayment(tx, order_uid)
		if err != nil {
			if err == ErrIndexOutOfRange {
				dbService.Logger.Error(
					"No such payment with selected order_uid",
					zap.String("order_uid", order_uid),
				)
			} else {
				dbService.Logger.Error("An error occured trying to select payment", zap.Error(err))
			}
			_ = tx.Rollback()
			return
		}

		delivery, err := dbService.selectDelivery(tx, order_uid)
		if err != nil {
			if err == ErrIndexOutOfRange {
				dbService.Logger.Error(
					"No such delivery with selected order_uid",
					zap.String("order_uid", order_uid),
				)
			} else {
				dbService.Logger.Error("An error occured trying to select delivery", zap.Error(err))
			}
			_ = tx.Rollback()
			return
		}

		items, err := dbService.selectItems(tx, order_uid)
		if err != nil {
			if err == ErrIndexOutOfRange {
				dbService.Logger.Error(
					"No such items with selected order_uid",
					zap.String("order_uid", order_uid),
				)
			} else {
				dbService.Logger.Error("An error occured trying to select items", zap.Error(err))
			}
			_ = tx.Rollback()
			return
		}

		err = tx.Commit()
		if err != nil {
			dbService.Logger.Error("An error occured while trying to commit transaction", zap.Error(err))
			return
		}

		orders[i].Payment = *payment
		orders[i].Delivery = *delivery
		orders[i].Items = *items

		dbService.Logger.Info("Load order to cache", zap.String("order_uid", order_uid))

		dbService.Cache.Set(order_uid, orders[i], cache.NoExpiration)
	}
}

func (dbService *DatabaseService) DumpCachedOrder(order_uid string) {
	for {
		tx, err := dbService.DB.BeginTxx(context.Background(), &sql.TxOptions{Isolation: sql.LevelDefault})
		if err != nil {
			dbService.Logger.Error("Error trying to begin transaction", zap.Error(err))
			time.Sleep(time.Second)
			dbService.Logger.Info("Retrying to dump cache")
			continue
		}

		order_i, found := dbService.Cache.Get(order_uid)
		if !found {
			dbService.Logger.Error("Order not found in cache! Aborting cache dumping")
			tx.Rollback()
			return
		}
		order_obj := order_i.(order.Order)
		err = dbService.AddOrder(tx, &order_obj)
		if err != nil {
			dbService.Logger.Error("Failed to dump cached order to database", zap.Error(err))
			tx.Rollback()
			time.Sleep(time.Second)
			dbService.Logger.Info("Retrying to dump cache")
			continue
		}

		err = tx.Commit()
		if err != nil {
			dbService.Logger.Error("Failed to commit transaction", zap.Error(err))
			tx.Rollback()
			time.Sleep(time.Second)
			dbService.Logger.Info("Retrying to dump cache")
			continue
		}

		dbService.Logger.Info(
			"Successfully dumped cached order to database",
			zap.String("order_uid", order_uid),
		)

		return
	}
}

func (dbService *DatabaseService) selectPayment(tx *sqlx.Tx, order_uid string) (*order.Payment, error) {
	payment := []order.Payment{}
	err := tx.Select(&payment,
		`SELECT
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
		FROM payment WHERE order_uid = $1`,
		order_uid,
	)

	if len(payment) < 1 {
		return nil, ErrIndexOutOfRange
	}

	return &payment[0], err
}

func (dbService *DatabaseService) selectDelivery(tx *sqlx.Tx, order_uid string) (*order.Delivery, error) {
	delivery := []order.Delivery{}
	err := tx.Select(&delivery,
		`SELECT
			order_uid,
			name,
			phone,
			zip,
			city,
			address,
			region,
			email
		FROM delivery WHERE order_uid = $1`,
		order_uid,
	)

	if len(delivery) < 1 {
		return nil, ErrIndexOutOfRange
	}

	return &delivery[0], err
}

func (dbService *DatabaseService) selectItems(tx *sqlx.Tx, order_uid string) (*[]order.Item, error) {
	items := []order.Item{}
	err := tx.Select(&items,
		`SELECT
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
		FROM items WHERE order_uid = $1`,
		order_uid,
	)

	if len(items) < 1 {
		return nil, ErrIndexOutOfRange
	}

	return &items, err
}
