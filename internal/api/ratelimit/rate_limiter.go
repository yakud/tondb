package ratelimit

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type RateLimiter struct {
	redis *redis.Client
}

func (r *RateLimiter) TouchAndCheckLimit(limits LimitsConfig, clientIdentifier string) (bool, error) {
	now := time.Now()
	res := RateLimitLua.EvalSha(
		r.redis,
		[]string{limits.LimitPrefix + clientIdentifier},
		limits.PerSecondLimit,
		limits.MinutelyLimit,
		limits.HourlyLimit,
		limits.DailyLimit,
		limits.MonthlyLimit,
		now.Format("2006-01-02_15-04-05"),
		now.Format("2006-01-02_15-04"),
		now.Format("2006-01-02_15"),
		now.Format("2006-01-02"),
		now.Format("2006-01"),
	)

	if res.Err() != nil {
		return false, res.Err()
	}

	val, ok := res.Val().([]interface{})
	if !ok || len(val) != 2 {
		return false, errors.New("unexpected val return count")
	}

	reason, ok := val[1].(string)
	if !ok {
		return false, errors.New("unexpected reason error")
	}

	if val[0] == nil || reason != "" {
		return true, fmt.Errorf("rate limit %q exceeded", reason)
	}

	return false, nil
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redis,
	}
}
