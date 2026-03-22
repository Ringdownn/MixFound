package main

import (
	"MixFound/core"
	"MixFound/redis"
)

func main() {
	redis.InitRedisClient()
	core.Initialize()
}
