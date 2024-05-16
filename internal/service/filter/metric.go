package filter

import (
	"context"
	"time"
)

func increaseCount(ctx context.Context, key string, expire time.Duration) error {
	_, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		if err == context.DeadlineExceeded {
			return nil
		}
		return err
	}
	_, err = rdb.Expire(ctx, key, expire).Result()
	if err != nil {
		return err
	}
	return nil
}
