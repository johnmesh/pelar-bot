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

	//err := getAccount(&account, email)
	//if err != nil {
	//	panic(err)
	//}
	//account 1
	//account.Email = "Lydiarugut@gmail.com"
	//account.Password = "hustle hard"
	//account.Message = "Hi, I deliver high-quality and plagiarism free work.Expect great communication and strict compliance with instructions and deadlines"

	//account 2
	//account.Email = "Jacknyangare@yahoo.com"
	//account.Password = "shark attack"
	//account.Message = "Hi, I am a versatile professional research and academic writer, specializing in research papers, essays, term papers, theses, and dissertations. NO PLAGIARISM..."

	//account 3
	//account.Email = "nambengeleashap@gmail.com"
	//	account.Password = "Optimus#On"
	//account 4
	//account.Email = "onderidismus85@gmail.com"
	//account.Password = "my__shark"

}
