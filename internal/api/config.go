package api

type Config struct {
	Addr                 string `envconfig:"ADDR" default:"0.0.0.0:8512"`
	ChAddr               string `envconfig:"CH_ADDR" default:"http://0.0.0.0:8123/default?compress=false&debug=false"`
	TlbBlocksFetcherAddr string `envconfig:"TLB_BLOCKS_FETCHER_ADDR" default:"127.0.0.1:13699"`

	RedisAddr     string `envconfig:"REDIS_ADDR" default:"127.0.0.1:6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`
}
