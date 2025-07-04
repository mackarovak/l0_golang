package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"l0/internal/models"
)

var DB *pgxpool.Pool

func InitDB(ctx context.Context, dsn string) error {
	var err error
	DB, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	log.Println("Connected to Postgres")
	return nil
}

func LoadOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := DB.Query(ctx, "SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(
			&o.OrderUID,
			&o.TrackNumber,
			&o.Entry,
			&o.Locale,
			&o.InternalSignature,
			&o.CustomerID,
			&o.DeliveryService,
			&o.Shardkey,
			&o.SmID,
			&o.DateCreated,
			&o.OofShard,
		)
		if err != nil {
			continue
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func SaveOrder(ctx context.Context, order models.Order) error {
	// внутри транзакции, чтобы данные были согласованными
	tx, err := DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (order_uid) DO NOTHING
	`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return err
	}

	// доставка
	_, err = tx.Exec(ctx, `
		INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO NOTHING
	`,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	// платеж
	_, err = tx.Exec(ctx, `
		INSERT INTO payment (
			transaction, order_uid, request_id, currency, provider, amount, 
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (transaction) DO NOTHING
	`,
		order.Payment.Transaction,
		order.OrderUID,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	// товары
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name, sale, 
				size, total_price, nm_id, brand, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}