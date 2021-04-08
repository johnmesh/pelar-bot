package selenium

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func Start(ctx *Context, port int) {

	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	const PATH = "/Users/mesh"
	const (
		// These paths will be different on your system.
		seleniumPath     = PATH + "/vendor/selenium-server-standalone-3.141.0.jar"
		chromeDriverPath = PATH + "/vendor/chromedriver_mac64"
		geckoDriverPath  = PATH + "/vendor/geckodriver"
	)

	const defaultTimeOut = 20 * time.Second

	selenium.SetDebug(false)

	_, err := selenium.NewChromeDriverService(chromeDriverPath, port)
	//defer service.Stop()
	if err != nil {
		//panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}

	chromeCaps := chrome.Capabilities{
		Args: []string{"--auto-open-devtools-for-tabs"},
	}

	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))

	if err != nil {
		panic(err)
	}
	//wd.ResizeWindow("", 600, 750)
	wd.MaximizeWindow("")
	//defer wd.Quit()

	// Navigate to the esshayshark page.
	if err := wd.Get("http://localhost:8080/login"); err != nil {
		panic(err)
	}

	elem, err := wd.FindElement(selenium.ByID, "loginEmail")
	if err != nil {
		panic(err)
	}
	elem.SendKeys("john.mengere@reliantid.com")

	elem, err = wd.FindElement(selenium.ByID, "loginPassword")
	if err != nil {
		panic(err)
	}
	elem.SendKeys("Daemon_5595")

	elem, err = wd.FindElement(selenium.ByID, "submitBtn")
	if err != nil {
		panic(err)
	}
	elem.Click()
	time.Sleep(2 * time.Second)
	wd.Get("http://localhost:8080/patients/user-438825")
	count := 0
	for {
		wd.Refresh()
		count++
		var elems []selenium.WebElement
		wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
			elem, _ = wd.FindElement(selenium.ByID, "chartEntry")
			if elem != nil {
				return true, nil
			}
			return false, nil
		}, defaultTimeOut)

		elems, _ = wd.FindElements(selenium.ByID, "chartEntry")
		if len(elems) < 2 {
			fmt.Println("The PCC note is missing")
			panic(err)
		}

		fmt.Println("Elements length:", len(elems), count)

		//time.Sleep(1 * time.Second)
	}

}
