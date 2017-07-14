package core

import "github.com/bysir-zl/bygo/cache"

var Redis *cache.BRedis

func init() {
	Redis = cache.NewRedis("127.0.0.1:6639")
	Redis.SetPrefix("chess")
}
