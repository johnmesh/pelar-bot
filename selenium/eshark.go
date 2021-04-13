package selenium

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type Context struct {
	Assigned map[string]string
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func convertTimeToMills(date string) int64 {
	var duration = strings.Split(date, " ")
	var t int64

	if len(duration) > 1 {
		if strings.Contains(duration[0], "h") {
			d, _ := time.ParseDuration(duration[0])
			t += d.Milliseconds()
		}
		if strings.Contains(duration[0], "d") {
			hours, _ := strconv.Atoi(strings.Split(duration[0], "d")[0])
			h := fmt.Sprintf("%dh", hours*24)
			d, _ := time.ParseDuration(h)
			t += d.Milliseconds()
		}
		if strings.Contains(duration[1], "h") {
			d, _ := time.ParseDuration(duration[1])
			t += d.Milliseconds()
		}
		if strings.Contains(duration[1], "m") {
			d, _ := time.ParseDuration(duration[1])
			t += d.Milliseconds()
		}
	} else {
		if strings.Contains(duration[0], "h") {
			d, _ := time.ParseDuration(duration[0])
			t += d.Milliseconds()
		}
		if strings.Contains(duration[0], "d") {
			hours, _ := strconv.Atoi(strings.Split(duration[0], "d")[0])
			h := fmt.Sprintf("%dh", hours*24)
			d, _ := time.ParseDuration(h)
			t += d.Milliseconds()
		}
		if strings.Contains(duration[0], "m") {
			d, _ := time.ParseDuration(duration[0])
			t += d.Milliseconds()
		}
	}

	return t

}

type Bidder struct {
	ID      int
	Port    int
	WG      *sync.WaitGroup
	Account Account
}

type Account struct {
	Email    string
	Password string
	Bids     map[string]string
}

func (b *Bidder) Start(ctx *Context) {
	defer b.WG.Done()
	defer func() {
		fmt.Println("The functions has ended")
	}()

	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	const PATH = "/Users/mesh"
	const (
		// These paths will be different on your system.
		seleniumPath     = "./vendor/selenium-server-standalone-3.141.0.jar"
		chromeDriverPath = "./vendor/chromedriver89_linux"
	)
	selenium.SetDebug(false)
	const defaultTimeOut = 20 * time.Second

	service, err := selenium.NewChromeDriverService(chromeDriverPath, b.Port)
	defer service.Stop()
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome", "pageLoadStrategy": "eager"}

	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless",
			"--no-sandbox",
			"--window-size=600,750",
			"--disable-dev-shm-usage",
			"--disable-gpu",
		},
		Path: "/usr/bin/google-chrome",
	}

	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", b.Port))

	if err != nil {
		panic(err)
	}
	wd.ResizeWindow("", 600, 750)
	defer wd.Quit()

	fmt.Println("-----Driver started successfully------")

	// Navigate to the esshayshark page.
	if err := wd.Get("https://essayshark.com/"); err != nil {
		panic(err)
	}
	elem, err := wd.FindElement(selenium.ByID, "es-cookie-button-submit")
	if err != nil {
		//panic(err)
	}
	elem.Click()

	wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		elem, _ := driver.FindElement(selenium.ByID, LOGIN_ACCOUNT_LINK_ID)
		if err := elem.Click(); err == nil {
			return true, nil
		}
		return false, nil
	}, defaultTimeOut)

	elem, err = wd.FindElement(selenium.ByXPATH, "//input[@id='id_esauth_login_field']")
	wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		err = elem.SendKeys(b.Account.Email)
		if err == nil {
			return true, nil
		}
		return false, nil
	}, defaultTimeOut)

	elem, err = wd.FindElement(selenium.ByXPATH, "//input[@id='id_esauth_pwd_field']")
	if err != nil {
		panic(err)
	}

	elem.SendKeys(b.Account.Password)
	wd.KeyDown(selenium.EnterKey)

	wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		elem, _ := driver.FindElement(selenium.ByID, ORDER_LOADING)

		if elem != nil {
			text, _ := elem.Text()
			if strings.ToLower(text) != "loading..." {
				return true, nil
			}
		}

		return false, nil
	}, defaultTimeOut)

	var count int
	fmt.Println("polling....")
	for {

		fmt.Printf("[%d]:polling... \n", b.ID)
		//Refresh the page to prevent the site from loggin out.
		if count > 1000 {
			wd.Get("https://essayshark.com/writer/orders/")
			count = 0
		}
		count++

		var orders []selenium.WebElement
		orders, err = wd.FindElements(selenium.ByXPATH, "//*[contains(@id,'id_order_container')]")
		if err != nil {
			continue

		}
		//sfmt.Println("ORDERS--->", len(orders))

		var order selenium.WebElement
		var orderNo string
		for i := range orders {
			order = nil
			dataID, err := orders[i].GetAttribute("data-id")
			//fmt.Println("orderNo", i)

			if err != nil {
				continue
			}

			if _, ok := ctx.Assigned[dataID]; ok {
				//The order is already taken
				//fmt.Println("The order is already taken", dataID)
				continue
			}
			//Add the order to the list
			ctx.Assigned[dataID] = "processing"
			orderNo = dataID
			break

		}

		if orderNo != "" {
			order = nil
			order, _ = wd.FindElement(selenium.ByXPATH, "//tr[@data-id ='"+orderNo+"']")
			od, _ := order.Text()
			fmt.Println("-----------------------------------------------", orderNo)
			fmt.Println(od)
			fmt.Println("-----------------------------------------------")

			order, _ = order.FindElement(selenium.ByXPATH, "//tr[@data-id ='"+orderNo+"']")
			orderType, err := order.FindElement(selenium.ByCSSSelector, ".service_type")
			if err != nil {
				panic(err)
			}

			if err := orderType.Click(); err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}

			text, err := orderType.Text()
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}
			fmt.Println("Service Name ------->", text)

			elem, err := order.FindElement(selenium.ByCSSSelector, ".pagesamount")
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}

			pages, err := elem.Text()
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}
			noOfPages := strings.Split(pages, "\n")[0]

			fmt.Println("No of Pages--->", noOfPages)

			elem, err = order.FindElement(selenium.ByCSSSelector, ".td_deadline")
			elem, err = elem.FindElement(selenium.ByCSSSelector, ".d-left")
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}

			deadline, err := elem.Text()
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}
			d := convertTimeToMills(deadline)
			fmt.Println("Deadline--->", d)

			elem, err = order.FindElement(selenium.ByCSSSelector, ".order_number")
			elem, err = elem.FindElement(selenium.ByCSSSelector, ".customer-rating")

			var customerRating string
			if elem != nil {
				customerRating, err = elem.Text()
				if err != nil {
					//no customer rating
					//continue

				}
			}
			fmt.Println("Customer Rating--->", customerRating)

			elem, err = order.FindElement(selenium.ByCSSSelector, ".order_number")
			elem, err = elem.FindElement(selenium.ByTagName, "a")
			custStatus, err := elem.GetAttribute("title")

			fmt.Println("Customer Status--->", custStatus)

			elem, err = order.FindElement(selenium.ByCSSSelector, ".budget")
			elem, err = elem.FindElement(selenium.ByCSSSelector, ".amount")
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}
			budget, err := elem.Text()

			bg, _ := strconv.ParseFloat(budget, 10)
			p, _ := strconv.Atoi(noOfPages)

			minBid := bg / float64(p)
			minBid = toFixed(minBid, 2)

			fmt.Println("MinBid--->", minBid, budget, noOfPages)

			elem, err = order.FindElement(selenium.ByCSSSelector, ".order_number")
			elem, err = elem.FindElement(selenium.ByCSSSelector, ".new-customer")
			var newCustomer string
			if elem != nil {
				newCustomer, err = elem.Text()
				if err != nil {
					//not a new customer rating
					//continue

				}
			}
			fmt.Println("New Customer--->", newCustomer)

			wd.Get("https://essayshark.com/writer/orders/" + orderNo + ".html")
			wd.Refresh()

			//download atleast one file
			filepath :=
				"//div[@class='paper_instructions_view']/a[contains (@data-url-raw,'/writer/get_additional_material.html')]"
			elem, err = wd.FindElement(selenium.ByXPATH, filepath)
			if elem == nil {
				fmt.Println("No files to donwnload", err)
			} else {
				wd.ExecuteScript("scroll(2000, 200)", nil)
				if err = elem.Click(); err != nil {
					//unable to donwload file
				}
			}
			amount := fmt.Sprintf("%.2f", minBid)

			var timer string
			wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
				elem, err = driver.FindElement(selenium.ByXPATH, "//span[@id='id_read_timeout_sec']")
				if elem != nil {
					timer, err = elem.Text()
					return timer != "", nil
				}
				//try bidding here
				if err := makeBid(amount, wd); err != nil {
					elem = nil
					return true, nil
				}

				return false, nil
			}, defaultTimeOut)

			if elem == nil {
				fmt.Println("element not found:----->", err)
				//Remove the order from the list
				//delete(ctx.Assigned, orderNo)
				ctx.Assigned[orderNo] = "done"
				wd.Get("https://essayshark.com/writer/orders/")
				continue
				//try bidding here
			}
			//Get the recommended bidding amount for the order
			elem, err = wd.FindElement(selenium.ByID, "id_order_bidding_form")
			elem, err = elem.FindElement(selenium.ByID, "rec_bid")

			rec, err := elem.Text()
			var amt string
			if rec != "" {
				r := strings.Split(rec, "$")[1]
				if v, ok := b.Account.Bids[r]; ok {
					amt = v
				}
			}

			fmt.Println("Rec-amount", amt, rec)
			if amt != "" {
				amount = amt
			}

			countDown, _ := strconv.Atoi(timer)
			start := time.Now()

			for countDown > 0 {
				//watch the timer

				elem, err = wd.FindElement(selenium.ByXPATH, "//span[@id='id_read_timeout_sec']")
				var tleft int
				if elem == nil {
					//no count down
					//try bidding
					wd.Refresh()
					fmt.Println("no count down--->", err)
				} else {
					timer, err = elem.Text()
					tleft, err = strconv.Atoi(timer)
				}

				d := time.Now().Sub(start).Seconds()
				duration := int(d)
				diff := countDown - duration

				if tleft < 30 || diff < 30 {
					fmt.Println("countdown", diff)
					//bid here
					wd.Refresh()
					if err := makeBid(amount, wd); err != nil {
						//Remove the order from the list
						//delete(ctx.Assigned, orderNo)
						ctx.Assigned[orderNo] = "done"
						wd.Get("https://essayshark.com/writer/orders/")
						break
					}

					//fmt.Println("bidding here")
				}

				if duration >= countDown {
					//The bidding has ended.This prevents infinite loops
					fmt.Println("The countdown has ended")
					//Remove the order from the list
					//delete(ctx.Assigned, orderNo)
					ctx.Assigned[orderNo] = "done"
					wd.Get("https://essayshark.com/writer/orders/")
					break
				}

			}

			/* time.Sleep(2 * time.Second)

			wd.ExecuteScript("scroll(2000, 200)", nil)
			elem, err = wd.FindElement(selenium.ByXPATH, "//input[contains(@class,'discard')]")
			elem.Click()

			wd.Get("https://essayshark.com/writer/orders/") */

			//Remove the order from the list
			//delete(ctx.Assigned, orderNo)

		}
		//time.Sleep(2 * time.Second)
	}

	//time.Sleep(20 * time.Second)

}

func makeBid(amount string, wd selenium.WebDriver) error {
	fmt.Println("make bid---->", amount)
	elem, err := wd.FindElement(selenium.ByID, "id_bid")
	if err != nil {
		return err
	}
	elem.Clear()
	elem.SendKeys(amount)
	wd.KeyDown(selenium.EnterKey)

	elem, err = wd.FindElement(selenium.ByID, "apply_order")
	if err != nil {

		return err
	}
	return nil
}

//CleanOrders delete the complete orders every 1min
func (b *Bidder) CleanOrders(ctx *Context) {
	start := time.Now()

	for {
		d := time.Now().Sub(start).Seconds()
		duration := int(d)

		if duration >= 60 {
			fmt.Println("cleaning---->", len(ctx.Assigned))
			for k, v := range ctx.Assigned {
				if v == "done" {
					delete(ctx.Assigned, k)
				}
			}

			start = time.Now()
			fmt.Println("cleaning complete---->", len(ctx.Assigned))
		}
	}

}
