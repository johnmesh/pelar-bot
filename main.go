package main

import (
	"fmt"
	"os"
	es "pelar-bot/selenium"
	"sync"

	"github.com/tebeka/selenium"
)

func main() {
	const (
		// These paths will be different on your system.
		seleniumPath     = "/Users/mesh/vendor/selenium-server-standalone-3.141.0.jar"
		chromeDriverPath = "/Users/mesh/vendor/chromedriver_mac64"
		geckoDriverPath  = "/Users/mesh/vendor/geckodriver"
		port             = 4015
	)

	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),             // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),              // Output debug information to STDERR.
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
	acc := es.Account{Email: "nambengeleashap@gmail.com", Password: "Optimus#On"}

	for i := 1; i <= 4; i++ {
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
	wg.Wait()

	fmt.Println("this is the main file")
}
