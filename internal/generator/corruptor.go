package generator

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"github.com/sillkiw/wb-l0/internal/domain"
)

type Corruptor struct {
	rate  float64
	kinds []string
	rnd   *rand.Rand
}

func NewCorruptor(rate float64, kindsCSV string) *Corruptor {
	kinds := parseKinds(kindsCSV)
	return &Corruptor{
		rate:  rate,
		kinds: kinds,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func parseKinds(s string) []string {
	if s == "" {
		return []string{"malformed", "unknown_field", "type_mismatch", "validation", "sums_mismatch", "future_date"}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Maybe возвращает исходные данные без изменений, либо "испорченные" и reason.
func (c *Corruptor) Maybe(order domain.Order, good []byte) ([]byte, string) {
	if c.rate <= 0 || c.rnd.Float64() >= c.rate {
		return good, ""
	}
	kind := c.kinds[c.rnd.Intn(len(c.kinds))]

	switch kind {
	case "malformed":
		return []byte(`{"order_uid":"` + order.OrderUID + `",`), "malformed_json"

	case "unknown_field":
		return []byte(`{"order_uid":"` + order.OrderUID + `","unknown":123}`), "unknown_field"

	case "type_mismatch":
		return []byte(`{"order_uid":"` + order.OrderUID + `","payment":{"payment_dt":"oops"}}`), "type_mismatch"

	case "sums_mismatch":
		o := order
		o.Payment.GoodsTotal = 42
		o.Payment.Amount = 42
		b, _ := json.Marshal(o)
		return b, "sums_mismatch"

	case "future_date":
		o := order
		o.DateCreated = time.Now().Add(24 * time.Hour)
		b, _ := json.Marshal(o)
		return b, "future_date"

	case "validation":
		o := order
		o.Items = nil
		o.Payment.Currency = "USD"
		b, _ := json.Marshal(o)
		return b, "validation_failed"
	}
	return good, ""
}
