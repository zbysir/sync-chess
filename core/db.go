package core

import "github.com/bysir-zl/bygo/cache"

var Redis *cache.BRedis

func init() {
	Redis = cache.NewRedis("127.0.0.1:6379")
	Redis.SetPrefix("chess")
}
