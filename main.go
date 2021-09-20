package main

import (
	"fmt"
	"log"
	"os"
	"pelar-bot/bot"
)

func main() {
	email := os.Getenv("EMAIL")

	if email == "" {
		log.Fatal("Email required")
	}
	fmt.Println("Account::::", email)

	bot.Init(email)

}
