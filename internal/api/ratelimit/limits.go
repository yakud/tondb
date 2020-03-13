package ratelimit

// limits in requests per period
type LimitsConfig struct {
	LimitPrefix    string
	PerSecondLimit int64
	MinutelyLimit  int64
	HourlyLimit    int64
	DailyLimit     int64
	MonthlyLimit   int64
}
