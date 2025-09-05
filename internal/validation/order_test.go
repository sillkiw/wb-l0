package validation

import (
	"testing"
	"time"

	"github.com/sillkiw/wb-l0/internal/domain"
)

// полностью валидный заказв
func mkValidOrder() domain.Order {
	now := time.Now().Add(-1 * time.Minute) // не в будущем
	item := domain.Item{
		ChrtID:      1,
		TrackNumber: "TN1",
		Price:       1000,
		Sale:        100,
		TotalPrice:  900,
		RID:         "rid-1",
		Name:        "Item name",
		Size:        "M",
		NmID:        100,
		Brand:       "Brand",
		Status:      1,
	}
	return domain.Order{
		OrderUID:          "ord-1",
		TrackNumber:       "TN1",
		Entry:             "WB",
		Delivery:          domain.Delivery{Name: "Ivan", Phone: "+79957105599", City: "Moscow", Address: "Tverskaya 1", Region: "RU", Email: "test@example.com"},
		Payment:           domain.Payment{Transaction: "tx-1", RequestID: "", Currency: "RUB", Provider: "visa", Amount: 1000, PaymentDT: time.Now().Unix(), Bank: "Bank", DeliveryCost: 100, GoodsTotal: 900, CustomFee: 0},
		Items:             []domain.Item{item},
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        "cust-1",
		DeliveryService:   "cdek",
		ShardKey:          "9",
		SmID:              1,
		DateCreated:       now,
		OofShard:          "1",
	}
}

// поиск ошибки по path и коду
func hasErr(me *MultiError, path string, code Code) bool {
	if me == nil {
		return false
	}
	for _, fe := range me.Fields {
		if fe.Path == path && fe.Code == code {
			return true
		}
	}
	return false
}

// --- тесты ---

func TestValidateOrder_OK(t *testing.T) {
	o := mkValidOrder()

	if me := ValidateOrder(o); me != nil {
		t.Fatalf("expected no errors, got: %v", me)
	}
}

func TestValidateOrder_RequiredFields(t *testing.T) {
	o := mkValidOrder()
	o.OrderUID = ""            // required
	o.Items = nil              // required
	o.Delivery.Phone = ""      // required
	o.Payment.Transaction = "" // required

	me := ValidateOrder(o)
	if me == nil {
		t.Fatalf("expected validation errors")
	}
	if !hasErr(me, "order_uid", CodeRequired) {
		t.Errorf("want error on order_uid: required")
	}
	if !hasErr(me, "items", CodeRequired) {
		t.Errorf("want error on items: required")
	}
	if !hasErr(me, "delivery.phone", CodeRequired) {
		t.Errorf("want error on delivery.phone: required")
	}
	if !hasErr(me, "payment.transaction", CodeRequired) {
		t.Errorf("want error on payment.transaction: required")
	}
}

func TestValidateOrder_FutureDate(t *testing.T) {
	o := mkValidOrder()
	o.DateCreated = time.Now().Add(10 * time.Minute)

	me := ValidateOrder(o)
	if me == nil || !hasErr(me, "date_created", CodeOutOfRange) {
		t.Fatalf("expected out_of_range on date_created, got: %v", me)
	}
}

func TestValidateOrder_SumsMismatch(t *testing.T) {
	o := mkValidOrder()
	// ломаем согласованность: goods_total не равен сумме items
	o.Payment.GoodsTotal = 888
	// и итоговая сумма станет неверной
	o.Payment.Amount = o.Payment.GoodsTotal + o.Payment.DeliveryCost + o.Payment.CustomFee

	me := ValidateOrder(o)
	if me == nil {
		t.Fatalf("expected validation errors")
	}
	if !hasErr(me, "payment.goods_total", CodeInconsistent) {
		t.Errorf("want inconsistent payment.goods_total")
	}
	// теперь ломаем amount:
	o = mkValidOrder()
	o.Payment.Amount = 999 // не равно goods_total + delivery_cost + custom_fee
	me = ValidateOrder(o)
	if me == nil || !hasErr(me, "payment.amount", CodeInconsistent) {
		t.Errorf("want inconsistent payment.amount, got: %v", me)
	}
}

func TestValidateOrder_ItemConstraints(t *testing.T) {
	o := mkValidOrder()
	o.Items[0].TotalPrice = -1

	me := ValidateOrder(o)
	if me == nil || !hasErr(me, "items[0].total_price", CodeOutOfRange) {
		t.Fatalf("expected out_of_range on items[0].total_price, got: %v", me)
	}
}

func TestValidateOrder_EmailPhoneFormat(t *testing.T) {
	o := mkValidOrder()
	o.Delivery.Email = "not-an-email"
	o.Delivery.Phone = "bad"

	me := ValidateOrder(o)
	if me == nil {
		t.Fatalf("expected validation errors")
	}
	if !hasErr(me, "delivery.email", CodeFormat) {
		t.Errorf("want format error on delivery.email")
	}
	if !hasErr(me, "delivery.phone", CodeFormat) {
		t.Errorf("want format error on delivery.phone")
	}
}

func TestValidateOrder_CurrencyFormat(t *testing.T) {
	o := mkValidOrder()
	o.Payment.Currency = "USD"

	me := ValidateOrder(o)
	if me == nil || !hasErr(me, "payment.currency", CodeFormat) {
		t.Fatalf("expected format error on payment.currency, got: %v", me)
	}
}
