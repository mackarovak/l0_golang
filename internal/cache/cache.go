package cache

import (
	"sync"
	"l0/internal/models"
)

type Cache struct {
	mu sync.RWMutex
	store map[string]models.Order
}

func NewCache() *Cache {
	return &Cache{
		store: make(map[string]models.Order),
	}
}

func (c *Cache) Get(OrderUID string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.store[OrderUID]
	return order, ok
}

func (c *Cache) Set(order models.Order) {
	c.mu.RLock()
	defer c.mu.Unlock()
	c.store[order.OrderUID] = order
}