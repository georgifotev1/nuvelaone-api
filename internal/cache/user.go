package cache

import (
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type UserStore struct {
	*BaseStore[domain.User]
}

func NewUserStore(rdb *redis.Client) *UserStore {
	return &UserStore{
		BaseStore: NewBaseStore[domain.User](rdb, "user", 5*time.Minute),
	}
}
