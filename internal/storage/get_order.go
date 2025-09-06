package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sillkiw/wb-l0/internal/domain"
)

var ErrNotFound = sql.ErrNoRows

// GetOrder возвращает заказ целиком по order_uid.
func (s *Storage) GetOrder(ctx context.Context, orderUID string) (domain.Order, error) {
	var o domain.Order

	// orders
	err := s.db.QueryRowContext(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`, orderUID).Scan(
		&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID,
		&o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Order{}, ErrNotFound
		}
		return domain.Order{}, fmt.Errorf("select orders: %w", err)
	}

	// delivery (1:1)
	err = s.db.QueryRowContext(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM deliveries WHERE order_uid = $1
	`, orderUID).Scan(
		&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
		&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
	)
	if err != nil && err != sql.ErrNoRows {
		return domain.Order{}, fmt.Errorf("select deliveries: %w", err)
	}

	// payment (1:1 по order_uid)
	err = s.db.QueryRowContext(ctx, `
		SELECT p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank,
		       p.delivery_cost, p.goods_total, p.custom_fee
		FROM payments p WHERE p.order_uid = $1
	`, orderUID).Scan(
		&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider,
		&o.Payment.Amount, &o.Payment.PaymentDT, &o.Payment.Bank, &o.Payment.DeliveryCost,
		&o.Payment.GoodsTotal, &o.Payment.CustomFee,
	)
	if err != nil && err != sql.ErrNoRows {
		return domain.Order{}, fmt.Errorf("select payments: %w", err)
	}

	// items (1:N)
	rows, err := s.db.QueryContext(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1 ORDER BY chrt_id
	`, orderUID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("select items: %w", err)
	}
	defer rows.Close()

	o.Items = o.Items[:0]
	for rows.Next() {
		var it domain.Item
		if err := rows.Scan(
			&it.ChrtID, &it.TrackNumber, &it.Price, &it.RID, &it.Name, &it.Sale, &it.Size,
			&it.TotalPrice, &it.NmID, &it.Brand, &it.Status,
		); err != nil {
			return domain.Order{}, fmt.Errorf("scan item: %w", err)
		}
		o.Items = append(o.Items, it)
	}
	if err := rows.Err(); err != nil {
		return domain.Order{}, fmt.Errorf("rows items: %w", err)
	}

	o.DateCreated = o.DateCreated.UTC()

	return o, nil
}

// Вспомогательный timeout-обёртка
func (s *Storage) GetOrderWithTimeout(parent context.Context, orderUID string, d time.Duration) (domain.Order, error) {
	ctx, cancel := context.WithTimeout(parent, d)
	defer cancel()
	return s.GetOrder(ctx, orderUID)
}
