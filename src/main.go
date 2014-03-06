package main

import (
	"skimmer"
)

func main() {
	api := skimmer.GetApi(&skimmer.Config{
		SessionSecret: "secret123",
	})
	api.Run()
}
