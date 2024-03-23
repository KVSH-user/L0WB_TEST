package natspublisher

import (
	"L0WB/internal/entity"
	"encoding/json"
	"github.com/nats-io/stan.go"
	"log"
)

func PublishOrder(sc stan.Conn, order entity.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		log.Printf("Marshaling error: %v", err)
		return err
	}

	if err := sc.Publish("orders", data); err != nil {
		log.Print(err)
		return err
	}

	log.Println("Message published!")
	return nil
}
