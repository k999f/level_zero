package cache

import (
	"database/sql"
	"level_zero/internal/db"
	"level_zero/internal/models"
)

type Cache struct {
	Data map[string]models.Order
}

func CreateCache() *Cache {
	return &Cache {
		Data: make(map[string]models.Order),
	}
}

func (c* Cache) InitialzeCache(conn *sql.DB) error {
	orders, err := db.GetAllOrders(conn)
	if err != nil {
		return err
	}
	for _, o := range orders {
		c.Data[o.OrderUid] = o
	}
	return nil
}

func (c *Cache) AddCache (order models.Order) {
	c.Data[order.OrderUid] = order
}

func (c *Cache) GetCache (uid string) (models.Order, bool) {
	data, ok := c.Data[uid]
	return data, ok
}