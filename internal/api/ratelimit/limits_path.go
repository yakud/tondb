package ratelimit

type LimitsPath struct {
}

func (l *LimitsPath) GetLimitsForPath(path string) *LimitsConfig {
	switch path {

	}
}

func NewLimitsPath() *LimitsPath {
	return &LimitsPath{}
}
