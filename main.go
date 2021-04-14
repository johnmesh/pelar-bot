package main

import (
	"fmt"
	es "pelar-bot/selenium"
	"strings"
	"sync"

	"github.com/tebeka/selenium"
)

func main() {
	const (
		// These paths will be different on your system.
		seleniumPath     = "./vendor/selenium-server-standalone-3.141.0.jar"
		chromeDriverPath = "./vendor/chromedriver89_linux"
		port             = 4015
	)

	//format disciplines
	exDisciplines := make(map[string]string)
	for _, v := range es.ExDiscipines {
		d := strings.Join(strings.Fields(v), "")
		d = strings.ToLower(d)
		exDisciplines[d] = d
	}

	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),             // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		/* 	selenium.Output(os.Stderr),   */ // Output debug information to STDERR.
	}

	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)

	defer service.Stop()
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	selenium.SetDebug(false)
	assigned := make(map[string]string)
	ctx := es.Context{Assigned: assigned}
	var wg sync.WaitGroup

	//Essayshark account information
	acc := es.Account{Email: "nambengeleashap@gmail.com", Password: "Optimus#On", Bids: es.Amount, ExDisciplines: exDisciplines}

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		bid := &es.Bidder{
			ID:      i,
			Port:    port,
			WG:      &wg,
			Account: acc,
		}
		//Start a subroutine
		go bid.Start(&ctx)
	}

	bid := &es.Bidder{}
	go bid.CleanOrders(&ctx)
	wg.Wait()

	fmt.Println("this is the main file")
}
