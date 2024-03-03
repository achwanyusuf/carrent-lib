package redis

import (
	"context"

	goredislib "github.com/redis/go-redis/v9"
)

type Redis struct {
	Url        string `mapstructure:"url"`
	Password   string `mapstructure:"password"`
	Username   string `mapstructure:"user_name"`
	PoolSize   int    `mapstructure:"pool_size"`
	MaxRetries int    `mapstructure:"max_retries"`
}

func RedisConnect(conf Redis) *goredislib.Client {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:       conf.Url,
		MaxRetries: conf.MaxRetries,
		PoolSize:   conf.PoolSize,
		Username:   conf.Username,
		Password:   conf.Password,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	return client
}
