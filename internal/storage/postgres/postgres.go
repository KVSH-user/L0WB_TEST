package postgres

import (
	"L0WB/internal/entity"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

type Storage struct {
	db *sql.DB
}

var ErrNotFound = errors.New("record not found")

// New инициализация БД
func New(host, port, user, password, dbName string) (*Storage, error) {
	const op = "storage.postgres.New"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := &Storage{db: db}

	// запуск миграций
	err = goose.Up(storage.db, "db/migrations")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return storage, nil
}

func (s *Storage) GetCache() ([]entity.Order, error) {
	const op = "storage.postgres.GetCache"

	// запрос с сортировкой
	query := `SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
  				o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
  				d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
  				p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
  				p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
  				i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale,
  				i.size, i.total_price, i.nm_id, i.brand, i.status
				FROM orders o
				LEFT JOIN delivery d ON o.order_uid = d.order_uid
				LEFT JOIN payment p ON o.order_uid = p.order_uid
				LEFT JOIN items i ON o.order_uid = i.order_uid
				ORDER BY o.order_uid;
				`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var orders []entity.Order
	ordersMap := make(map[string]*entity.Order)

	for rows.Next() {
		var (
			o           entity.Order
			d           entity.Delivery
			p           entity.Payment
			i           entity.Item
			paymentDt   sql.NullTime
			dateCreated sql.NullTime
		)

		if err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID,
			&o.DeliveryService, &o.Shardkey, &o.SmID, &dateCreated, &o.OofShard,
			&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &paymentDt,
			&p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
			&i.ChrtID, &i.TrackNumber, &i.Price, &i.Rid, &i.Name, &i.Sale,
			&i.Size, &i.TotalPrice, &i.NmID, &i.Brand, &i.Status,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if paymentDt.Valid {
			p.PaymentDt = paymentDt.Time
		}
		if dateCreated.Valid {
			o.DateCreated = dateCreated.Time
		}

		o.Delivery = d
		o.Payment = p

		// проверяем, есть ли уже заказ в мапе
		if existingOrder, exists := ordersMap[o.OrderUID]; exists {
			existingOrder.Items = append(existingOrder.Items, i)
		} else {
			o.Items = append(o.Items, i)
			ordersMap[o.OrderUID] = &o
		}
	}

	// переносим данные из мапы в слайс
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// возвращаем слайс
	return orders, nil
}

func (s *Storage) AddToDB(order entity.Order) error {
	const op = "storage.postgres.AddToDB"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// запрос для order
	orderQuery := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id,
						delivery_service, shardkey, sm_id, date_created, oof_shard)
				   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err := tx.Exec(orderQuery, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID,
		order.DateCreated, order.OofShard); err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	// запрос для delivery
	deliveryQuery := `INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
					  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.Exec(deliveryQuery, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email); err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	// запрос для payment
	paymentQuery := `INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt,
						 bank, delivery_cost, goods_total, custom_fee)
					  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err := tx.Exec(paymentQuery, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt,
		order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee); err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	// запрос для items
	itemQuery := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
				   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, item := range order.Items {
		if _, err := tx.Exec(itemQuery, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status); err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Close закрытие подключения к бд
func (s *Storage) Close() error {
	return s.db.Close()
}
