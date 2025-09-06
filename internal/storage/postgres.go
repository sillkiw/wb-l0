package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sillkiw/wb-l0/internal/dbutils"
	"github.com/sillkiw/wb-l0/internal/domain"
)

type Storage struct {
	db *sql.DB
}

// NewStorage инициализирует подключение к PostgreSQL
func NewStorage(connStr string) (*Storage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveOrder(ctx context.Context, order domain.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tx begin failed: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// --- orders (UPSERT 1:1) ---
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders(
			order_uid, track_number, entry, locale, internal_signature, customer_id,
			delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number       = EXCLUDED.track_number,
			entry              = EXCLUDED.entry,
			locale             = EXCLUDED.locale,
			internal_signature = EXCLUDED.internal_signature,
			customer_id        = EXCLUDED.customer_id,
			delivery_service   = EXCLUDED.delivery_service,
			shardkey           = EXCLUDED.shardkey,
			sm_id              = EXCLUDED.sm_id,
			date_created       = EXCLUDED.date_created,
			oof_shard          = EXCLUDED.oof_shard
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated.UTC(), order.OofShard)
	if err != nil {
		return fmt.Errorf("upsert order failed: %w", err)
	}

	// --- deliveries (UPSERT 1:1) ---
	_, err = tx.ExecContext(ctx, `
		INSERT INTO deliveries(order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO UPDATE SET
			name   = EXCLUDED.name,
			phone  = EXCLUDED.phone,
			zip    = EXCLUDED.zip,
			city   = EXCLUDED.city,
			address= EXCLUDED.address,
			region = EXCLUDED.region,
			email  = EXCLUDED.email
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("upsert delivery failed: %w", err)
	}

	// --- payments (UPSERT по PK transaction) ---
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments(
			transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank,
			delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10, $11)
		ON CONFLICT (transaction) DO UPDATE SET
			order_uid     = EXCLUDED.order_uid,
			request_id 	  = EXCLUDED.request_id,
			currency      = EXCLUDED.currency,
			provider      = EXCLUDED.provider,
			amount        = EXCLUDED.amount,
			payment_dt    = EXCLUDED.payment_dt,
			bank          = EXCLUDED.bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total   = EXCLUDED.goods_total,
			custom_fee    = EXCLUDED.custom_fee
	`, order.Payment.Transaction, order.OrderUID, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT,
		order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("upsert payment failed: %w", err)
	}

	// --- items ---
	// Удаление старых заказов
	if _, err := tx.ExecContext(ctx, `DELETE FROM items WHERE order_uid = $1`, order.OrderUID); err != nil {
		return fmt.Errorf("delete items failed: %w", err)
	}

	// батч-вставка новых (если они есть)
	if len(order.Items) > 0 {
		cols := []string{
			"order_uid", "chrt_id", "track_number", "price", "rid", "name",
			"sale", "size", "total_price", "nm_id", "brand", "status",
		}
		rows := make([][]interface{}, 0, len(order.Items))
		for _, it := range order.Items {
			rows = append(rows, []interface{}{
				order.OrderUID, it.ChrtID, it.TrackNumber, it.Price,
				it.RID, it.Name, it.Sale, it.Size,
				it.TotalPrice, it.NmID, it.Brand, it.Status,
			})
		}

		query, args := dbutils.BuildBatchInsert("items", cols, rows)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("insert items failed: %w", err)
		}
	}

	// --- commit ---
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx commit failed: %w", err)
	}
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
