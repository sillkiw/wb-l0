package generator

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sillkiw/wb-l0/internal/domain"
)

func NewFakeOrder(i int) domain.Order {
	cities := []string{"Moscow", "SPB", "Kazan", "Novosibirsk", "Ekaterinburg"}
	brands := []string{"Nike", "Adidas", "Puma", "Reebok"}

	n := 1 + rand.Intn(3)
	items := make([]domain.Item, 0, n)
	for j := 0; j < n; j++ {
		items = append(items, domain.Item{
			ChrtID:      rand.Intn(1000000),
			TrackNumber: fmt.Sprintf("track-%04d", rand.Intn(10000)),
			Price:       500 + rand.Intn(2000),
			Name:        fmt.Sprintf("Product-%d", j),
			Brand:       brands[rand.Intn(len(brands))],
			Status:      1,
		})
	}

	return domain.Order{
		OrderUID:    fmt.Sprintf("order-%d", i),
		TrackNumber: fmt.Sprintf("track-%04d", rand.Intn(10000)),
		Entry:       "WEB",
		Delivery: domain.Delivery{
			Name:    fmt.Sprintf("User-%d", i),
			Phone:   "+79998887766",
			City:    cities[rand.Intn(len(cities))],
			Address: "Some street 1",
			Region:  "Region",
			Email:   fmt.Sprintf("user%d@example.com", i),
		},
		Payment: domain.Payment{
			Transaction:  fmt.Sprintf("tx-%d", rand.Intn(100000)),
			Currency:     "RUB",
			Provider:     "card",
			Amount:       1000 + rand.Intn(5000),
			PaymentDT:    time.Now().Unix(),
			Bank:         "Sberbank",
			DeliveryCost: 200 + rand.Intn(500),
			GoodsTotal:   n,
		},
		Items:       items,
		Locale:      "ru",
		CustomerID:  fmt.Sprintf("cust-%d", rand.Intn(100)),
		DateCreated: time.Now().Add(-time.Duration(rand.Intn(100)) * time.Hour),
	}
}
