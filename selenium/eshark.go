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

func formatText(s string) string {
	d := strings.Join(strings.Fields(s), "")
	d = strings.ToLower(d)
	return d
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
	Email         string
	Password      string
	Bids          map[string]string
	ExDisciplines map[string]string
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
			"--dns-prefetch-disable",
			"--window-size=1920,1080",
			"enable-automation",
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
		panic(err)
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

	wd.Get("https://essayshark.com/writer/orders/")

	//Discard all orders
	for {
		var orders []selenium.WebElement
		wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
			orders, err = wd.FindElements(selenium.ByXPATH, "//*[contains(@id,'id_order_container')]")
			if len(orders) > 0 {
				return true, nil
			}

			return false, nil
		}, 5*time.Second)

		if len(orders) < 1 {
			wd.Refresh()
			break
		}
		//Find the discard button
		elem, err = wd.FindElement(selenium.ByID, "discard_all_visible")
		wd.ExecuteScript("scroll(2000, 10)", nil)
		if elem != nil {
			err = elem.Click()
		}

		//Click the modal popup
		elem, err = wd.FindElement(selenium.ByCSSSelector, ".ZebraDialog_Buttons")
		if elem != nil {
			elem, err = elem.FindElement(selenium.ByCSSSelector, ".ZebraDialog_Button_1")
			if elem != nil {
				err = elem.Click()
			}
		}

	}

	//start looking for work
	//var count int
	for {
		fmt.Printf("[%d]:polling... \n", b.ID)
		//wd.Refresh()
		//Refresh the page to prevent the site from loggin out.
		/* 	if count > 1000 {
			wd.Get("https://essayshark.com/writer/orders/")
			count = 0
		}
		count++
		*/
		var orders []selenium.WebElement

		wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
			wd.Refresh()
			orders, err = wd.FindElements(selenium.ByXPATH, "//*[contains(@id,'id_order_container')]")
			if len(orders) > 0 {
				return true, nil
			}

			return false, nil
		}, 60*time.Second)

		if len(orders) < 1 {
			continue
		}

		//fmt.Println("orders--->", len(orders))

		//var order selenium.WebElement
		var orderNo string
		for i := range orders {
			//order = nil
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
			//order = nil
			//order, _ = wd.FindElement(selenium.ByXPATH, "//tr[@data-id ='"+orderNo+"']")
			/* od, _ := order.Text()
			fmt.Println("-----------------------------------------------", orderNo)
			fmt.Println(od)
			fmt.Println("-----------------------------------------------")

			order, _ = order.FindElement(selenium.ByXPATH, "//tr[@data-id ='"+orderNo+"']")
			elem, err := order.FindElement(selenium.ByCSSSelector, ".service_type")
			if err != nil {
				panic(err)
			}

			if err := elem.Click(); err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}

			orderType, err := elem.Text()
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			}
			fmt.Println("Service Name ------->", orderType)

			elem, err = order.FindElement(selenium.ByCSSSelector, ".pagesamount")
			if err != nil {
				//delete(ctx.Assigned, orderNo)
				//continue
			} */
			/* 	var noOfPages string
			if elem != nil {
				pages, err := elem.Text()
				if err != nil {
					//delete(ctx.Assigned, orderNo)
					//continue
				}
				noOfPages := strings.Split(pages, "\n")[0]

				fmt.Println("No of Pages--->", noOfPages)
			} else {
				wd.Refresh()
			} */
			/*
			 * This section gets the order deadline
			 */
			/* elem, err = order.FindElement(selenium.ByCSSSelector, ".td_deadline")
			if elem != nil {
				elem, err = elem.FindElement(selenium.ByCSSSelector, ".d-left")
				if err != nil {
					//delete(ctx.Assigned, orderNo)
					//continue
				}
			} else {
				wd.Refresh()
			}

			if elem != nil {
				deadline, err := elem.Text()
				if err != nil {
					//delete(ctx.Assigned, orderNo)
					//continue
				}
				d := convertTimeToMills(deadline)
				fmt.Println("Deadline--->", d)
			} else {
				wd.Refresh()
			} */

			/*
			 * This section checks the customer ratings
			 */
			/* elem, err = order.FindElement(selenium.ByCSSSelector, ".order_number")
			if elem != nil {
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
			} else {
				wd.Refresh()
			} */

			/*
			 * This section checks the customer status
			 */
			/* 	elem, err = order.FindElement(selenium.ByCSSSelector, ".order_number")
			if elem != nil {
				elem, err = elem.FindElement(selenium.ByTagName, "a")
				if elem != nil {
					custStatus, _ := elem.GetAttribute("title")
					fmt.Println("Customer Status--->", custStatus)
				}
			} else {
				wd.Refresh()
			} */

			/*
			 * This section checks the budget amount
			 */
			/* 	var minBid float64
			elem, err = order.FindElement(selenium.ByCSSSelector, ".budget")
			if elem != nil {
				elem, err = elem.FindElement(selenium.ByCSSSelector, ".amount")
				if err != nil {
					//delete(ctx.Assigned, orderNo)
					//continue
				}

				if elem != nil {
					budget, _ := elem.Text()
					bg, _ := strconv.ParseFloat(budget, 10)
					p, _ := strconv.Atoi(noOfPages)

					minBid = bg / float64(p)
					minBid = toFixed(minBid, 2)

					fmt.Println("MinBid--->", minBid, budget, noOfPages)
				}

			} else {
				wd.Refresh()
			} */

			/*
			 * This section checks if its a new custmomer
			 */
			/* 	elem, err = order.FindElement(selenium.ByCSSSelector, ".order_number")
			if elem != nil {
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
			} else {
				wd.Refresh()
			} */

			wd.Get("https://essayshark.com/writer/orders/" + orderNo + ".html")
			wd.Refresh()

			/*
			 * This section checks the order discipline
			 */
			elem, err = wd.FindElement(selenium.ByCSSSelector, ".fast_order_details")
			if elem != nil {
				elem, err = elem.FindElement(selenium.ByCSSSelector, ".d50")
			}
			if elem != nil {
				elems, _ := elem.FindElements(selenium.ByCSSSelector, "dl")
				elem, err = elems[3].FindElement(selenium.ByCSSSelector, "dd")
				if elem != nil {
					discipline, _ := elem.Text()
					fmt.Println("Order-Discipline----->", formatText(discipline))
				}
			}

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
			//amount := fmt.Sprintf("%.2f", minBid)
			var amount string
			/*
			 * This section checks the remommended bidding amount
			 */
			elem, err = wd.FindElement(selenium.ByID, "id_order_bidding_form")
			var amt, r string
			if elem != nil {
				elem, err = elem.FindElement(selenium.ByID, "rec_bid")
				if elem != nil {
					rec, _ := elem.Text()
					if rec != "" {
						r = strings.Split(rec, "$")[1]
						if v, ok := b.Account.Bids[r]; ok {
							amt = v
						}
					}
					fmt.Println("Rec-amount", amt, rec)
				}

				if amt != "" {
					amount = amt
				} else if amount == "" {
					amount = r
				}

			} else {
				wd.Refresh()
			}

			var timer string
			wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
				elem, err = driver.FindElement(selenium.ByXPATH, "//span[@id='id_read_timeout_sec']")
				if elem != nil {
					timer, err = elem.Text()
					return timer != "", nil
				}
				//try bidding here
				if err := makeBid(amount, wd, amt, 0); err != nil {
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
			//input, err := wd.FindElement(selenium.ByID, "id_bid")
			countDown, _ := strconv.ParseInt(timer, 10, 64)

			start := time.Now()
			timeout := time.Duration(countDown)*time.Second + 1

			wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
				/* elem, err = driver.FindElement(selenium.ByXPATH, "//span[@id='id_read_timeout_sec']")
				if elem != nil {
					timer, err = elem.Text()
					return timer != "", nil
				} */
				//try bidding here
				if err = makeBid(amount, wd, amt, 0); err != nil {
					return true, nil
				}
				wd.Refresh()

				return false, nil
			}, timeout)

			fmt.Println("Done:", time.Now().Sub(start).Seconds(), timeout, err)

			ctx.Assigned[orderNo] = "done"
			wd.Get("https://essayshark.com/writer/orders/")

			// This is where the migic happens
			/* Loop:
			for countDown > 0 {
				//watch the timer
				elem, err = wd.FindElement(selenium.ByXPATH, "//span[@id='id_read_timeout_sec']")
				var tleft int
				if elem == nil {
					//no count down
					//try bidding
					//wd.Refresh()
					fmt.Println("no count down--->", err)
				} else {
					timer, err = elem.Text()
					tleft, err = strconv.Atoi(timer)
				}

				d := time.Now().Sub(start).Seconds()
				duration := int(d)
				diff := countDown - duration */

			/* if tleft < 30 || diff < 30 {
				fmt.Println("countdown", diff)
				//bid here

				if err := makeBid(amount, wd, amt, orderType, countDown); err != nil {
					//Remove the order from the list
					//delete(ctx.Assigned, orderNo)

				}
				//wd.Refresh()
				//fmt.Println("bidding here")
			} */

			/* 	if tleft < 30 || diff < 30 {
				fmt.Println("duration:", duration, "diff:", diff, "tleft:", tleft)
				wd.Refresh()
			Nested:
				for {
					fmt.Println("OrderNo:", orderNo, "Amount:", amount)
					input, _ := wd.FindElement(selenium.ByID, "id_bid")
					fmt.Println("Input:", input)
					if input == nil {
						ctx.Assigned[orderNo] = "done"
						wd.Get("https://essayshark.com/writer/orders/")
						break Nested

					}
					input.Clear()
					input.SendKeys(amount)
					wd.KeyDown(selenium.EnterKey)

					//wd.Refresh()
				}
				break Loop */

			//The bidding has ended.This prevents infinite loops
			//fmt.Println("The countdown has ended")
			//Remove the order from the list
			//delete(ctx.Assigned, orderNo)
			//ctx.Assigned[orderNo] = "done"
			//wd.Get("https://essayshark.com/writer/orders/")

		}

	}

	/*
	 * time.Sleep(2 * time.Second)
	 * wd.ExecuteScript("scroll(2000, 200)", nil)
	 * elem, err = wd.FindElement(selenium.ByXPATH, "//input[contains(@class,'discard')]")
	 * elem.Click()
	 */

	//}

	//}

}

func makeBid(amount string, wd selenium.WebDriver, amt string, countDown int) error {
	fmt.Println("make bid---->", amount, "amt:", amt, "Count down:", countDown)
	elem, err := wd.FindElement(selenium.ByID, "id_bid")
	if err != nil {
		return err
	}
	elem.Clear()
	elem.SendKeys(amount)
	wd.KeyDown(selenium.EnterKey)

	/* elem, err = wd.FindElement(selenium.ByID, "apply_order")
	if elem == nil {
		//return errors.New("elem not found")
	} else {
		//elem.Click()
	}
	/* if err = elem.Click(); err != nil {
		return err
	} */
	/* if err != nil {
		return err
	} */
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
