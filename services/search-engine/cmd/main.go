package main

import (
	"MixFound/services/search-engine/internal/core"
	"MixFound/services/search-engine/internal/redis"
)

func main() {
	redis.InitRedisClient()
	core.Initialize()
}
