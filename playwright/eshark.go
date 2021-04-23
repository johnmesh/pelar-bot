package playwright

import (
	"fmt"
	"log"
	es "pelar-bot/selenium"
	"strconv"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

type Bidder struct {
	ID int
}

func (b *Bidder) Start(ctx *es.Context) {
	const defaultTimeOut float64 = 5

	fmt.Println("Starting....")
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	//no := false
	browser, err := pw.Chromium.Launch( /* playwright.BrowserTypeLaunchOptions{Headless: &no} */ )

	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage()
	page.SetViewportSize(1000, 1080)
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	fmt.Printf("-----[%d]Driver started successfully------", b.ID)
	if _, err = page.Goto("https://essayshark.com/"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	elem, err := page.QuerySelector("#es-cookie-button-submit")
	err = elem.Click()
	if err != nil {
		panic(err)
	}

	elem, err = page.QuerySelector("#id_esauth_myaccount_login_link")
	if err != nil {
		panic(err)
	}

	if err := elem.Click(); err != nil {
		panic(err)
	}

	elem, err = page.QuerySelector("#id_esauth_login_field")
	if err != nil {
		panic(err)
	}

	elem.Fill("nambengeleashap@gmail.com")

	elem, err = page.QuerySelector("#id_esauth_pwd_field")
	if err != nil {
		panic(err)
	}

	elem.Fill("Optimus#On")
	page.Keyboard().Down("Enter")
	page.WaitForLoadState("domcontentloaded")

	page.Goto("https://essayshark.com/writer/orders")
	/* if _, err = page.Goto("https://essayshark.com/writer/orders"); err != nil {
		log.Fatalf("could not goto: %v", err)
	} */
	for {
		var orders []playwright.ElementHandle
		start := time.Now()
	Polling:
		for {

			diff := time.Now().Sub(start).Seconds()
			if diff > 30 {
				page.Reload()
				start = time.Now()
			}
			fmt.Printf("[%d]:polling... \n", b.ID)
			orders, _ = page.QuerySelectorAll("//*[contains(@id,'id_order_container')]")
			if len(orders) > 0 {
				break Polling
			}
		}
		var orderNo string

	FindOrders:
		for i := range orders {
			//order = nil
			dataID, err := orders[i].GetAttribute("data-id")
			//fmt.Println("orderNo", i)

			if err != nil {
				continue FindOrders
			}

			if _, ok := ctx.Assigned[dataID]; ok {
				//The order is already taken
				//fmt.Println("The order is already taken", dataID)
				continue FindOrders
			}
			//Add the order to the list
			ctx.Assigned[dataID] = "processing"
			orderNo = dataID
			break FindOrders

		}

		if orderNo == "" {
			continue
		}

		//dataID, err := elems[0].GetAttribute("data-id")
		fmt.Println("dataID:", orderNo)
		orderURL := "https://essayshark.com/writer/orders/" + orderNo + ".html"

		page.Goto(orderURL)

		//download atleast one file
		filepath :=
			"//div[@class='paper_instructions_view']/a[contains (@data-url-raw,'/writer/get_additional_material.html')]"
		elem, err = page.QuerySelector(filepath)
		if elem == nil {
			fmt.Println("No files to donwnload", err)
		} else {
			if err = elem.Click(); err != nil {
				panic(err)
			}
		}

		start = time.Now()

		/*
		 * This section checks the remommended bidding amount
		 */
		var amount float64
		elem, err = page.QuerySelector("#id_order_bidding_form")
		var amt, r string
		if elem != nil {
			elem, err = page.QuerySelector("#rec_amount")

			if elem != nil {
				rec, _ := elem.TextContent()
				if rec != "" {
					r = rec
					fmt.Println("REC--->", r)
					if v, ok := es.Amount[r]; ok {
						amt = v
					}
				}

			}

			if amt != "" {
				amount, _ = strconv.ParseFloat(amt, 64)
			} else if amount == 0 {
				r = strings.TrimSuffix(r, "\n")
				fmt.Print(r)
				amount, err = strconv.ParseFloat(r, 64)
				if err != nil {
					panic(err)
				}
			}

		} else {
			page.Reload()
		}

		fmt.Println("Amount-->", amount, "AMT:", amt, "Rec:", r)
		page.Reload()

	TimerChecker:
		for {
			elem, err = page.QuerySelector("//span[@id='id_read_timeout_sec']")
			if elem != nil {
				if timer, _ := elem.TextContent(); timer != "" {
					break TimerChecker
				}

			}
			if err := makeBid(amount, amt, page, orderNo, 0, 0); err != nil {
				elem = nil
				break TimerChecker
			}
			//break after the timeout
			if time.Now().Sub(start).Seconds() >= defaultTimeOut {
				break TimerChecker
			}
		}

		if elem == nil {
			fmt.Println("element not found:----->", err)
			//Remove the order from the list
			//delete(ctx.Assigned, orderNo)
			//ctx.Assigned[orderNo] = "done"
			page.Goto("https://essayshark.com/writer/orders")
			continue

			//try bidding here

		}

		timer, _ := elem.TextContent()
		fmt.Println(timer)

		countDown, _ := strconv.ParseInt(timer, 10, 64)
		start = time.Now()

	Bidding:
		for {
			page.Goto(orderURL)

			//break after the timeout

			d := time.Now().Sub(start).Seconds()
			duration := int(d)
			diff := int(countDown) - duration

			if err := makeBid(amount, amt, page, orderNo, 0, diff); err != nil {
				break Bidding
			}

			fmt.Println("countdown")
			if diff < 0 {
				break Bidding
			}
		}

		fmt.Println("Done:", orderNo)
		page.Goto("https://essayshark.com/writer/orders")

	}

}

func makeBid(amount float64, amt string, page playwright.Page, orderNo string, id int, countDown int) error {
	fmt.Println("make bid---->", amount, "amt:", amt, "#Order:", orderNo, "ID:", id, "Count Down:", countDown)
	elem, err := page.QuerySelector("#id_bid")
	if err != nil {
		return err
	}

	fmt.Println("bidding....0")
	bidAmount := fmt.Sprintf("%.2f", amount)
	if elem != nil {
		var timeout float64 = 1000
		err = elem.Fill(bidAmount, playwright.ElementHandleFillOptions{Timeout: &timeout})
		page.Keyboard().Down("Enter")

	}

	return nil
}
