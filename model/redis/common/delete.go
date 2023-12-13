package common

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

func DeleteAll(ctx context.Context, c *redis.Client, prefix string) (int64, error) {
	var cursor uint64
	var err error
	var counter int64
	for {
		var keys []string
		keys, cursor, err = c.Scan(ctx, cursor, prefix, 1000).Result()
		if err != nil {
			break
		}
		i, e := c.Del(ctx, keys...).Result()
		counter += i
		if e != nil {
			err = errors.Join(err, e)
		}
		if cursor == 0 {
			break
		}
	}
	return counter, err
}
