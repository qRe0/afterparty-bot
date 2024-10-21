package main

import (
	"fmt"
	"log"

	configs "github.com/qRe0/afterparty-bot/config"
)

func main() {
	cfg, err := configs.LoadEnv()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Envs: %+v", cfg)
}
