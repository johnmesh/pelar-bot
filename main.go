package main

import (
	"fmt"
	es "pelar-bot/selenium"
	"strconv"
	"strings"
	"sync"

	"github.com/tebeka/selenium"
)

/* func main() {

	assigned := make(map[string]string)
	ctx := es.Context{Assigned: assigned}
	var wg sync.WaitGroup

	for i := 0; i < 1; i++ {
		bid := &pr.Bidder{
			ID: i,
		}
		wg.Add(1)
		//Start a subroutine
		go bid.Start(&ctx)
	}

	wg.Wait()

} */
func main() {
	const (
		// These paths will be different on your system.
		seleniumPath     = "/vendor/selenium-server-standalone-4.0.0-alpha-1.jar"
		chromeDriverPath = "/vendor/chromedriver_92linux"
	/* 	port             = 4015 */
	)

	//format disciplines
	exDisciplines := make(map[string]string)
	for _, v := range es.ExDiscipines {
		d := strings.Join(strings.Fields(v), "")
		d = strings.ToLower(d)
		exDisciplines[d] = v
	}

	//fmt.Println("Fomarted Disclipines", exDisciplines)

	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),             // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		//selenium.Output(os.Stderr),              // Output debug information to STDERR.
	}

	selenium.SetDebug(false)
	assigned := make(map[string]string)
	ctx := es.Context{Assigned: assigned}
	var wg sync.WaitGroup

	//Essayshark account information
	acc := es.Account{Email: "nambengeleashap@gmail.com", Password: "Optimus#On", Bids: es.Amount, ExDisciplines: exDisciplines}
	//acc := es.Account{Email: "Jacknyangare@yahoo.com", Password: "shark attack", Bids: es.Amount, ExDisciplines: exDisciplines}

	for i := 1; i <= 3; i++ {
		p := fmt.Sprintf("403%d", i)
		port, _ := strconv.Atoi(p)
		service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
		defer service.Stop()
		if err != nil {
			panic(err) // panic is used only as an example and is not otherwise recommended.
		}

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
