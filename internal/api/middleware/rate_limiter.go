package middleware

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.flora.loc/mills/tondb/internal/api/ratelimit"

	"github.com/julienschmidt/httprouter"
)

/*
2020/02/23 16:50:56 map[Accept:[application/json, text/plain, ] Accept-Encoding:[gzip] Accept-Language:[ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7] Access-Control-Allow-Origin:[*] Authorization:[Basic dG9uYXBpOlFucldXOXE0WFZ0NWZHQ2NhTkd2a05mUQ==] Cache-Control:[no-cache] Cdn-Loop:[cloudflare] Cf-Connecting-Ip:[87.248.239.215] Cf-Ipcountry:[RU] Cf-Ray:[569ab3a0fa86c3e3-LED] Cf-Visitor:[{"scheme":"https"}] Connection:[close] Origin:[http://localhost:8080] Pragma:[no-cache] Referer:[http://localhost:8080/] Sec-Fetch-Dest:[empty] Sec-Fetch-Mode:[cors] Sec-Fetch-Site:[cross-site] User-Agent:[Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.116 Safari/537.36] X-Forwarded-For:[87.248.239.215, 172.69.10.4] X-Forwarded-Proto:[https] X-Real-Ip:[172.69.10.4]]
*/

func RateLimit(rateLimiter *ratelimit.RateLimiter) func(h httprouter.Handle) httprouter.Handle {
	return func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			h(w, r, ps)
			return

			clientIp := r.Header.Get("Cf-Connecting-Ip")
			if clientIp == "" {
				clientIp = r.Header.Get("X-Real-IP")
			}

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
				"/top/whales",
				"/messages/feed":
				limits = ratelimit.LimitsConfig{
					LimitPrefix:    "main:",
					PerSecondLimit: 15,
				}

			default:
				limits = ratelimit.LimitsConfig{
					LimitPrefix:    "core:",
					PerSecondLimit: 5,
				}
			}

			if limitExceeded, err := rateLimiter.TouchAndCheckLimit(limits, clientIp); limitExceeded {
				if err != nil {
					// TODO: fallback to inmemory counter
					log.Println(fmt.Errorf("rateLimiter error: %w", err))
				}
				log.Printf("limitExceeded: %s", clientIp)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusTooManyRequests)
				return
			}
			h(w, r, ps)
		}
	}
}
