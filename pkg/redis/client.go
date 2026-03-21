package redis

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(opt asynq.RedisClientOpt) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     opt.Addr,
		Password: opt.Password,
		DB:       opt.DB,
	})
}

func NewRedisOpts(addr, password string, db int) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
}
