package validation

import (
	"strings"
	"time"

	"github.com/sillkiw/wb-l0/internal/domain"
)

func ValidateOrder(o domain.Order) *MultiError {
	me := &MultiError{Retriable: false}

	// Required
	req := func(v string, path string) {
		if !nonEmpty(v) {
			me.Add(path, CodeRequired, "must not be empty")
		}
	}

	req(o.OrderUID, "order_uid")
	req(o.TrackNumber, "track_number")
	req(o.Entry, "entry")
	req(o.CustomerID, "customer_id")
	req(o.DeliveryService, "delivery_service")
	req(o.Locale, "locale")
	req(o.ShardKey, "shardkey")
	req(o.OofShard, "oof_shard")

	// DateCreated
	if o.DateCreated.IsZero() {
		me.Add("date_created", CodeRequired, "missing timestamp")
	} else {
		if !tsNotFuture(o.DateCreated) {
			me.Add("date_created", CodeOutOfRange, "timestamp in the future")
		}
		// (опционально) слишком старая
		_ = time.Now()
	}

	// Delivery
	req(o.Delivery.Name, "delivery.name")
	req(o.Delivery.Phone, "delivery.phone")
	req(o.Delivery.City, "delivery.city")
	req(o.Delivery.Address, "delivery.address")
	req(o.Delivery.Region, "delivery.region")
	if o.Delivery.Email != "" && !parseableEmail(o.Delivery.Email) {
		me.Add("delivery.email", CodeFormat, "invalid email")
	}
	if o.Delivery.Phone != "" && !phoneLike(o.Delivery.Phone) {
		me.Add("delivery.phone", CodeFormat, "invalid phone")
	}

	// Items
	if len(o.Items) == 0 {
		me.Add("items", CodeRequired, "at least one item")
	} else {
		total := 0
		seenRID := map[string]struct{}{}
		for i, it := range o.Items {
			p := func(n string) string { return pathItem(i, n) }
			if it.ChrtID <= 0 {
				me.Add(p("chrt_id"), CodeOutOfRange, "must be > 0")
			}
			req(it.Name, p("name"))
			if it.Price < 0 {
				me.Add(p("price"), CodeOutOfRange, "negative")
			}
			if it.TotalPrice < 0 {
				me.Add(p("total_price"), CodeOutOfRange, "negative")
			}
			if it.Sale < 0 {
				me.Add(p("sale"), CodeOutOfRange, "negative")
			}
			if it.Status < 0 {
				me.Add(p("status"), CodeOutOfRange, "negative")
			}
			if it.RID != "" {
				if _, ok := seenRID[it.RID]; ok {
					me.Add(p("rid"), CodeInconsistent, "duplicate rid")
				}
				seenRID[it.RID] = struct{}{}
			}
			total += it.TotalPrice
		}
		// сверка сумм
		if o.Payment.GoodsTotal != total {
			me.Add("payment.goods_total", CodeInconsistent, "!= sum(items.total_price)")
		}
	}

	// Payment
	req(o.Payment.Transaction, "payment.transaction")
	req(o.Payment.Provider, "payment.provider")
	req(o.Payment.Currency, "payment.currency")
	if o.Payment.Amount < 0 {
		me.Add("payment.amount", CodeOutOfRange, "negative")
	}
	if o.Payment.DeliveryCost < 0 {
		me.Add("payment.delivery_cost", CodeOutOfRange, "negative")
	}
	if o.Payment.GoodsTotal < 0 {
		me.Add("payment.goods_total", CodeOutOfRange, "negative")
	}
	if o.Payment.CustomFee < 0 {
		me.Add("payment.custom_fee", CodeOutOfRange, "negative")
	}
	cur := strings.ToUpper(strings.TrimSpace(o.Payment.Currency))
	if cur == "" {
		me.Add("payment.currency", CodeRequired, "must not be empty")
	} else if len(cur) != 3 {
		me.Add("payment.currency", CodeFormat, "expect 3-letter ISO code")
	} else if !isAllowedCurrency(cur) {
		me.Add("payment.currency", CodeFormat, "unsupported currency; allowed: "+allowedCurrenciesList())
	}
	if o.Payment.PaymentDT <= 0 || o.Payment.PaymentDT > 4102444800 { // ~2100-01-01
		me.Add("payment.payment_dt", CodeOutOfRange, "invalid unix seconds")
	}

	want := o.Payment.GoodsTotal + o.Payment.DeliveryCost + o.Payment.CustomFee
	if o.Payment.Amount != want {
		me.Add("payment.amount", CodeInconsistent, "amount != goods_total + delivery_cost + custom_fee")
	}

	if me.HasErrors() {
		return me
	}
	return nil
}
