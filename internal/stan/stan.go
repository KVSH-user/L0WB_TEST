package stan

import (
	"L0WB/internal/cache"
	"L0WB/internal/entity"
	"encoding/json"
	"fmt"
	"github.com/nats-io/stan.go"
	"log/slog"
)

type OrderToDB interface {
	AddToDB(order entity.Order) error
}

type Client struct {
	sc stan.Conn
}

func NewClient(clusterID, clientID, natsURL string) (*Client, error) {
	const op = "handlers.stan.NewClient"

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &Client{sc: sc}, nil
}

func (c *Client) Subscribe(subject string, cb stan.MsgHandler) (stan.Subscription, error) {
	return c.sc.Subscribe(subject, cb, stan.DurableName("my-durable"))
}

func OrderMessage(log *slog.Logger, orderToDB OrderToDB, m *stan.Msg) error {
	const op = "handlers.stan.OrderMessage"

	var order entity.Order
	if err := json.Unmarshal(m.Data, &order); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	// Добавляем заказ в базу данных
	if err := orderToDB.AddToDB(order); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	// Добавляем заказ в кеш
	if err := cache.AddToCache(order); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	log.Info("order processed successfully")

	return nil
}

func (c *Client) Close() {
	if c.sc != nil {
		c.sc.Close()
	}
}
