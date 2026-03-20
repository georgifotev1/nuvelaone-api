package cache

import (
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Users *UserStore
}

func New(rdb *redis.Client) *Cache {
	return &Cache{
		Users: NewUserStore(rdb),
	}
}
