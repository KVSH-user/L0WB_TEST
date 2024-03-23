package cache

import (
	"L0WB/internal/entity"
	"errors"
	"log/slog"
)

var cache = map[string]entity.Order{}

type GetterCache interface {
	GetCache() ([]entity.Order, error)
}

func Init(log *slog.Logger, getterCache GetterCache) error {
	const op = "cache.Init"

	log = log.With(slog.String("op", op))

	// выгрузка данных из БД в слайс структур
	orders, err := getterCache.GetCache()
	if err != nil {
		log.Error("failed to get order list: ", err)
		return err
	}

	// добавление в кеш(мапу)
	for _, order := range orders {
		cache[order.OrderUID] = order
	}

	log.Info("cache initialized")

	return nil
}

func AddToCache(getedOrder entity.Order) error {
	cache[getedOrder.OrderUID] = getedOrder
	return nil
}

func GetOrder(id string) (entity.Order, error) {
	cacheData, ok := cache[id]
	if !ok {
		return entity.Order{}, errors.New("bad order id")
	}
	return cacheData, nil
}
