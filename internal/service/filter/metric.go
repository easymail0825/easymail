package filter

import (
	"context"
	"time"
)

// truncateTimeToTenMinutes truncates the given time to the nearest 10 minutes boundary
func truncateTimeToTenMinutes(t time.Time) time.Time {
	// Get the integer part of the minutes
	minutes := t.Minute()
	// Calculate the remaining minutes to determine if a 10 minute increment is needed to reach the next 10 minutes boundary
	remainder := minutes % 10
	// If the remaining minutes are not 0, subtract the remainder and add 10 minutes
	if remainder != 0 {
		t = t.Add(-time.Duration(remainder) * time.Minute).Add(10 * time.Minute)
	}
	// Truncate the seconds and nanoseconds to ensure only the date and time hours and minutes are included
	return t.Truncate(10 * time.Minute)
}

// formatTimeForTenMinutes formats the truncated time as a string in the specified format
func formatTimeForTenMinutes(t time.Time) string {
	truncated := truncateTimeToTenMinutes(t)
	// Format the time string
	return truncated.Format("200601021504")
}

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
