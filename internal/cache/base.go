package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type BaseStore[T any] struct {
	rdb    *redis.Client
	prefix string
	ttl    time.Duration
}

func NewBaseStore[T any](rdb *redis.Client, prefix string, ttl time.Duration) *BaseStore[T] {
	return &BaseStore[T]{
		rdb:    rdb,
		prefix: prefix,
		ttl:    ttl,
	}
}

func (s *BaseStore[T]) key(id string) string {
	return fmt.Sprintf("%s:%s", s.prefix, id)
}

func (s *BaseStore[T]) Get(ctx context.Context, id string) (*T, error) {
	data, err := s.rdb.Get(ctx, s.key(id)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var val T
	if data != "" {
		if err := json.Unmarshal([]byte(data), &val); err != nil {
			return nil, err
		}
	}

	return &val, nil
}

func (s *BaseStore[T]) Set(ctx context.Context, id string, val *T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return s.rdb.SetEx(ctx, s.key(id), data, s.ttl).Err()
}

func (s *BaseStore[T]) Delete(ctx context.Context, id string) error {
	return s.rdb.Del(ctx, s.key(id)).Err()
}
