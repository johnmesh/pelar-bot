package main

import (
	"fmt"
	"pelar-bot/bot"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	bot.Init()

	//Essayshark account information
	//acc := Account{Email: "nambengeleashap@gmail.com", Password: "Optimus#On", Bids: es.Amount, ExDisciplines: exDisciplines}
	//acc := es.Account{Email: "Jacknyangare@yahoo.com", Password: "shark attack", Bids: es.Amount, ExDisciplines: exDisciplines}
	//acc := es.Account{Email: "onderidismus85@gmail.com", Password: "my__shark", Bids: es.Amount, ExDisciplines: exDisciplines}
	//acc := es.Account{Email: "Lydiarugut@gmail.com", Password: "hustle hard", Bids: es.Amount, ExDisciplines: exDisciplines}

	fmt.Println("this is the main file:::")
}
