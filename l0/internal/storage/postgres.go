package storage

import (
	"context"
	"fmt"
	"time"
	"log"
	"l0/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB(ctx context.Context, dsn string) error {
	var err error
	DB, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Connected to PostgreSQL")
	return nil
}

func LoadOrders(ctx context.Context) ([]models.Order, error) {
	tx, err := DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, "SELECT * FROM orders")
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
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
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Load delivery
		err = tx.QueryRow(ctx, 
			"SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1", 
			o.OrderUID,
		).Scan(
			&o.Delivery.Name,
			&o.Delivery.Phone,
			&o.Delivery.Zip,
			&o.Delivery.City,
			&o.Delivery.Address,
			&o.Delivery.Region,
			&o.Delivery.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load delivery: %w", err)
		}

		// Load payment
		err = tx.QueryRow(ctx,
			"SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid = $1",
			o.OrderUID,
		).Scan(
			&o.Payment.Transaction,
			&o.Payment.RequestID,
			&o.Payment.Currency,
			&o.Payment.Provider,
			&o.Payment.Amount,
			&o.Payment.PaymentDt,
			&o.Payment.Bank,
			&o.Payment.DeliveryCost,
			&o.Payment.GoodsTotal,
			&o.Payment.CustomFee,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load payment: %w", err)
		}

		// Load items
		itemRows, err := tx.Query(ctx,
			"SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1",
			o.OrderUID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query items: %w", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item models.Item
			err = itemRows.Scan(
				&item.ChrtID,
				&item.TrackNumber,
				&item.Price,
				&item.Rid,
				&item.Name,
				&item.Sale,
				&item.Size,
				&item.TotalPrice,
				&item.NmID,
				&item.Brand,
				&item.Status,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan item: %w", err)
			}
			o.Items = append(o.Items, item)
		}

		orders = append(orders, o)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orders, nil
}

func SaveOrder(ctx context.Context, order models.Order) error {
    tx, err := DB.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    // Если дата пустая, используем текущее время
    if order.DateCreated == "" {
        order.DateCreated = time.Now().Format(time.RFC3339)
    }

    _, err = tx.Exec(ctx, `
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature, 
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        ON CONFLICT (order_uid) DO NOTHING`,
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
		return fmt.Errorf("failed to insert order: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO NOTHING`,
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
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payment (
			transaction, order_uid, request_id, currency, provider, amount, 
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (transaction) DO NOTHING`,
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
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name, sale, 
				size, total_price, nm_id, brand, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
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
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	return tx.Commit(ctx)
}