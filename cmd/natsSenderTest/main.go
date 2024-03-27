package main

import (
	"L0WB/internal/entity"
	"L0WB/internal/natspublisher"
	"github.com/nats-io/stan.go"
	"log"
	"time"
)

func main() {
	sc, err := stan.Connect("test-cluster", "publisher-client", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal("cant connect to NATS: ", err)
	}
	defer sc.Close()

	// тестовые данные
	order := entity.Order{
		OrderUID:    "wborder",
		TrackNumber: "TN123456",
		Entry:       "entry1",
		Delivery: entity.Delivery{
			Name:    "John Wb",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "City Name",
			Address: "Address Line 1",
			Region:  "Region Name",
			Email:   "john.doe@example.com",
		},
		Payment: entity.Payment{
			Transaction:  "txn_123456",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "provider_name",
			Amount:       100,
			PaymentDt:    time.Now(),
			Bank:         "Bank Name",
			DeliveryCost: 5,
			GoodsTotal:   95,
			CustomFee:    0,
		},
		Items: []entity.Item{
			{
				ChrtID:      1,
				TrackNumber: "TN123456",
				Price:       95,
				Rid:         "rid123456",
				Name:        "Item Name",
				Sale:        10,
				Size:        "M",
				TotalPrice:  85,
				NmID:        123456,
				Brand:       "Brand Name",
				Status:      1,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "cust123456",
		DeliveryService:   "delivery_service_name",
		Shardkey:          "shardkey123",
		SmID:              1,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	// публикация
	if err := natspublisher.PublishOrder(sc, order); err != nil {
		log.Fatal("error to send message: ", err)
	}
}
