package generator

import (
	"fmt"
	"time"

	"github.com/sillkiw/wb-l0/internal/domain"
)

func NewFakeOrder(i int) domain.Order {
	cities := []string{"Moscow", "Saint-Petersburg", "Kazan", "Novosibirsk", "Ekaterinburg", "Krasnodar"}
	regions := []string{"Moscow Region", "Leningrad Region", "Tatarstan", "Novosibirsk Region", "Sverdlovsk Region", "Krasnodar Krai"}
	brands := []string{"Nike", "Adidas", "Puma", "Reebok", "Asics", "NewBalance"}
	sizes := []string{"XS", "S", "M", "L", "XL", "XXL"}
	services := []string{"WB-Delivery", "CDEK", "Boxberry", "DHL", "DPD"}
	banks := []string{"Sberbank", "VTB", "Tinkoff", "Alfa-Bank"}

	// Заголовок заказа
	orderUID := newOrderID()
	// Если хочешь прежний простой формат — верни WB%08d
	orderTrack := newTrackNumber()

	// Позиции
	n := 1 + rng.Intn(3) // 1..3 items
	items := make([]domain.Item, 0, n)
	var goodsSum int
	for j := 0; j < n; j++ {
		price := 500 + rng.Intn(5000) // 500..5499
		salePct := rng.Intn(31)       // 0..30 (%)
		total := int((100 - salePct) * price / 100)

		it := domain.Item{
			ChrtID:      rng.Intn(1_000_000),
			TrackNumber: orderTrack, // согласованно с заказом
			Price:       price,
			RID:         fmt.Sprintf("RID-%s", randHex(8)),
			Name:        fmt.Sprintf("Product-%d", j+1),
			Sale:        salePct,
			Size:        pick(sizes),
			TotalPrice:  total,
			NmID:        100000 + rng.Intn(900000),
			Brand:       pick(brands),
			Status:      1,
		}
		items = append(items, it)
		goodsSum += total
	}

	// Оплата (согласованные суммы)
	deliveryCost := 200 + rng.Intn(600) // 200..799
	customFee := rng.Intn(100)          // 0..99
	amount := goodsSum + deliveryCost + customFee

	// Доставка
	dlv := domain.Delivery{
		Name:    fmt.Sprintf("User-%d", i),
		Phone:   "+7" + randDigits(10),
		Zip:     randDigits(6),
		City:    pick(cities),
		Address: fmt.Sprintf("Street %d, house %d", 1+rng.Intn(20), 1+rng.Intn(100)),
		Region:  pick(regions),
		Email:   fmt.Sprintf("user%d@example.com", i),
	}

	// Платёж (ULID-идентификаторы для устойчивости к параллелизму)
	pay := domain.Payment{
		Transaction:  newTransactionID(),
		RequestID:    newRequestID(),
		Currency:     "RUB",
		Provider:     "card",
		Amount:       amount,
		PaymentDT:    time.Now().Unix(),
		Bank:         pick(banks),
		DeliveryCost: deliveryCost,
		GoodsTotal:   goodsSum,
		CustomFee:    customFee,
	}

	// Итоговый заказ
	return domain.Order{
		OrderUID:          orderUID,
		TrackNumber:       orderTrack,
		Entry:             "WEB",
		Delivery:          dlv,
		Payment:           pay,
		Items:             items,
		Locale:            "ru",
		InternalSignature: randHex(16),
		CustomerID:        fmt.Sprintf("cust-%03d", rng.Intn(1000)),
		DeliveryService:   pick(services),
		ShardKey:          fmt.Sprintf("%d", 1+rng.Intn(12)),
		SmID:              1 + rng.Intn(100),
		DateCreated:       time.Now().Add(-time.Duration(rng.Intn(240)) * time.Minute).UTC(),
		OofShard:          fmt.Sprintf("%d", 1+rng.Intn(12)),
	}
}
