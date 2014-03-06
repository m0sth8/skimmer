package main

import (
	"skimmer"
	"flag"
)

var (
	config = skimmer.Config{
		SessionSecret: "secret123",
		RedisConfig: skimmer.RedisConfig{
			RedisAddr: "127.0.0.1:6379",
			RedisPassword: "",
			RedisPrefix: "skimmer",
		},
	}
)

func init() {
	flag.StringVar(&config.Storage, "storage", "memory", "available storages: redis, memory")
	flag.StringVar(&config.SessionSecret, "sessionSecret", config.SessionSecret, "")
	flag.StringVar(&config.RedisAddr, "redisAddr", config.RedisAddr, "redis storage only")
	flag.StringVar(&config.RedisPassword, "redisPassword", config.RedisPassword, "redis storage only")
	flag.StringVar(&config.RedisPrefix, "redisPrefix", config.RedisPrefix, "redis storage only")
}

func main() {
	flag.Parse()
	api := skimmer.GetApi(&config)
	api.Run()
}
