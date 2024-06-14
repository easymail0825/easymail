package filter

import (
	"context"
	"time"
)

func increaseCount(ctx context.Context, key string, expire time.Duration) (int64, error) {
	n, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		if err == context.DeadlineExceeded {
			return 0, err
		}
		return 0, err
	}
	_, err = rdb.Expire(ctx, key, expire).Result()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func addSet(ctx context.Context, key string, value string, expire time.Duration) (n int64, err error) {
	_, err = rdb.SAdd(ctx, key, value).Result()
	if err != nil {
		if err == context.DeadlineExceeded {
			return 0, err
		}
		return 0, err
	}
	n, err = rdb.SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	_, err = rdb.Expire(ctx, key, expire).Result()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func queryCacheInString(ctx context.Context, key string) (string, error) {
	value, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == context.DeadlineExceeded {
			return "", err
		}
		return "", err
	}
	return value, nil
}

func setCacheInString(ctx context.Context, key string, value string, expire time.Duration) error {
	err := rdb.Set(ctx, key, value, expire).Err()
	if err != nil {
		return err
	}
	return nil
}
