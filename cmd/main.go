package main

import (
	"log"

	_ "github.com/lib/pq"
	"github.com/qRe0/afterparty-bot/internal/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
