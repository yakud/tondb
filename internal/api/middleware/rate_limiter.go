package middleware

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/api/ratelimit"

	"github.com/julienschmidt/httprouter"
)

//const _user, _pass = "tonapi", "QnrWW9q4XVt5fGCcaNGvkNfQ"

func RateLimit(rateLimiter *ratelimit.RateLimiter) func(h httprouter.Handle) httprouter.Handle {
	return func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			clientIp := r.Header.Get("X-Real-IP")

			var limits ratelimit.LimitsConfig
			switch r.URL.Path {
			case
				"/timeseries/blocks-by-workchain",
				"/timeseries/messages-by-type",
				"/timeseries/volume-by-grams",
				"/timeseries/messages-ord-count",
				"/messages/latest",
				"/b/feed",
				"/block/feed",
				"/blocks/feed",
				"/addr/top-by-message-count",
				"/top/whales":
				limits = ratelimit.LimitsConfig{
					LimitPrefix:    "fast",
					PerSecondLimit: 15,
				}

			default:
				limits = ratelimit.LimitsConfig{
					LimitPrefix:    "core",
					PerSecondLimit: 5,
				}
			}

			if limitExceeded, err := rateLimiter.TouchAndCheckLimit(limits, clientIp); limitExceeded {
				if err != nil {
					// TODO: fallback to inmemory counter
					log.Println(fmt.Errorf("rateLimiter error: %w", err))
				}
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusTooManyRequests)
				return
			}
			h.Handler(w, r, ps)
		}
	}
}
