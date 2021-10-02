package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

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

var AssignedOrders = make(map[string]string)
var allDiscarded = false

//locks
var mlock = &sync.Mutex{}
var slock = &sync.Mutex{}
var dlock = &sync.Mutex{}

type Bidder struct {
	ID      int
	Port    int
	WG      *sync.WaitGroup
	Account Account
	Service *selenium.Service
	Run     bool
}

type OrderFilterDetails struct {
	MinDeadline        int64    `bson:"minDeadline"`
	MaxDeadline        int64    `bson:"maxDeadline"`
	MinPages           int      `bson:"minPages"`
	MaxPages           int      `bson:"maxPages"`
	MaxUrgencyPages    int      `bson:"maxUrgencyPages"`
	DiscardAssignments bool     `bson:"discardAssignments"`
	DiscardEditting    bool     `bson:"discardEditting"`
	ExDiscipline       []string `bson:"exDiscipline"`
}

type CustomerDetails struct {
	CompleteOrders     int     `bson:"completeOrders"`
	MinRating          float32 `bson:"minRating"`
	DiscardOfflineCust bool    `bson:"discardOfflineCust"`
	DiscardNoRatings   bool    `bson:"discardNoRatings"`
	DiscardNewCust     bool    `bson:"discardNewCust"`
}

type Bid struct {
	Rec    float32 `bson:"rec"`
	Amount float32 `bson:"amount"`
}
type BiddingPrice struct {
	Bids []Bid `bson:"bids"`
}

type Account struct {
	Address         string             `bson:"address"`
	Email           string             `bson:"email"`
	Password        string             `bson:"password"`
	OrderDetails    OrderFilterDetails `bson:"orderDetails"`
	CustomerDetails CustomerDetails    `bson:"customerDetails"`
	Bids            []Bid              `bson:"bids"`
	Status          string             `bson:"status"`
	ExDisciplines   map[string]string
	Message         string
}

type OrderDetails struct {
	ID            int               `json:"id"`
	Discipline2AR map[string]string `json:"discipline2_ar"`
}

type Order struct {
	ID             string  `json:"id"`
	OrderRead      string  `json:"order_read"`
	Pages          string  `json:"pages_qty"`
	Amount         float32 `json:"min_price_total"`
	CustomerRating string  `json:"customer_rating"`
	CustomerOrder  string  `json:"customer_orders"`
	OnlineStatus   string  `json:"online_status"`
	NewCustomer    string  `json:"customer_debut"`
	Deadline       string  `json:"deadline_dt_ts"`

	Discipline2AR struct {
		Title string `json:"title"`
	} `json:"discipline2_ar"`

	ServiceType struct {
		Slug string `json:"slug"`
	} `json:"service_type_ar"`

	OutDated string `json:"bid_outdated"`
}
type AvailableItems struct {
	Orders []Order `json:"new_items"`
}

type Ping struct {
	TimeRemain  int `json:"read_time_remain"`
	FilesRemain int `json:"files_download_remain"`
}

func getAccount(account *Account, email string) (err error) {
	err = FetchAccount(email, account)
	return
}

func Init(email string) {
	const (
		seleniumPath     = "/vendor/selenium-server-standalone-4.0.0-alpha-2.jar"
		chromeDriverPath = "/vendor/chromedriver_94linux"
	)
	selenium.SetDebug(false)

	var isBotRunnig = false
	var bidders []*Bidder

	for {
		//sync data
		var account Account
		err := getAccount(&account, email)
		if err != nil {
			fmt.Println("Error:::", err)
		}

		fmt.Println("Account:::", account.Email, account.Password, account.Status)

		if account.Status == "on" && !isBotRunnig {
			//start bidding
			exDisciplines := make(map[string]string)
			for _, v := range account.OrderDetails.ExDiscipline {
				d := formatText(v)
				exDisciplines[d] = v
			}
			account.ExDisciplines = exDisciplines

			opts := []selenium.ServiceOption{
				selenium.StartFrameBuffer(),
				selenium.ChromeDriver(chromeDriverPath),
				//selenium.Output(os.Stderr),
			}

			//launch the services
			for i := 1; i <= 3; i++ {
				p := fmt.Sprintf("801%d", i)
				port, _ := strconv.Atoi(p)
				service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)

				if err != nil {
					panic(err)
				}
				defer service.Stop()
				bidder := &Bidder{
					ID:      i,
					Port:    port,
					Account: account,
					Service: service,
					Run:     true,
				}

				bidders = append(bidders, bidder)
				//start a subroutine
				go bidder.Start()
			}
			isBotRunnig = true
		} else if account.Status == "off" && isBotRunnig {
			//stop bidding
			for i := 0; i < len(bidders); i++ {
				bidders[i].Run = false
			}
			isBotRunnig = false
		} else if account.Status == "on" && isBotRunnig {
			//sync the data
			for i := 0; i < len(bidders); i++ {
				bidders[i].Account = account
			}
		}
		//Sync every 5 seconds
		time.Sleep(5 * time.Second)

	}

}

func (b *Bidder) Start() {

	const defaultTimeOut = 10 * time.Second
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}

	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--no-sandbox",
			"--headless",
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

	var wg sync.WaitGroup

	//distribute the threads
	var noOfThreads int
	if b.ID == 1 {
		noOfThreads = 5
	} else if b.ID == 2 {
		noOfThreads = 3
	} else if b.ID == 3 {
		noOfThreads = 1
	}

	for i := 0; i < noOfThreads; i++ {
		wg.Add(1)

		//launch poller subroutines
		go func() {
			defer wg.Done()
			slock.Lock()
			wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", b.Port))

			if err != nil {
				panic(err)
			}
			slock.Unlock()

			wd.ResizeWindow("", 1400, 750)
			defer wd.Quit()

			fmt.Println("-----Driver started successfully------")

			// Navigate to the esshayshark page.
			if err := wd.Get("https://essayshark.com/"); err != nil {
				panic(err)
			}

			wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
				elem, err := wd.FindElement(selenium.ByXPATH, "/html/body/header/div/div/button[2]")
				if err = elem.Click(); err == nil {
					return true, nil
				}

				return false, nil
			}, defaultTimeOut)

			if err != nil {
				panic(err)
			}

			if err = wd.Get("https://essayshark.com/writer/orders/"); err != nil {
				panic(err)
			}

			elem, err := wd.FindElement(selenium.ByXPATH, "//input[@id='id_esauth_login_field']")
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
				return false, nil
			}, defaultTimeOut)

			wd.Get("https://essayshark.com/writer/orders/")

			cookie, _ := wd.GetCookie("a11nt3n")
			auth_token := cookie.Value

			client := &http.Client{}

			ordersURL := "https://essayshark.com/writer/orders/aj_source.html?act=load_list&nobreath=1&session_more_qty=0&session_discarded=0&_=1629218589134"
			req, _ := http.NewRequest("GET", "", bytes.NewBuffer([]byte("")))
			req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
			req.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
			req.URL, _ = url.Parse(ordersURL)
			var available AvailableItems
			//Discard all orders
			dlock.Lock()
			if !allDiscarded {
				for {

					res, err := client.Do(req)
					if err != nil {
						panic(err)
					}
					json.NewDecoder(res.Body).Decode(&available)
					if len(available.Orders) == 0 {
						wd.Refresh()
						break
					}

					var od []string
					for i := 0; i < len(available.Orders); i++ {
						od = append(od, available.Orders[i].ID)
					}
					ids := strings.Join(od, ",")

					form := url.Values{}
					form.Add("act", "discard_all")
					form.Add("nobreath", "1")
					form.Add("ids", ids)

					discardAllReq, _ := http.NewRequest("POST", "https://essayshark.com/writer/orders/aj_source.html", strings.NewReader(form.Encode()))
					discardAllReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					discardAllReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
					discardAllReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
					_, err = client.Do(discardAllReq)
					if err != nil {
						panic(err)
					}

				}
				allDiscarded = true
			}
			dlock.Unlock()
			count := 0
			fmt.Printf("[%d]:polling... \n", b.ID)

			wd.Get("https://essayshark.com/writer/orders/")

		Polling:
			for {

				res, _ := client.Do(req)
				json.NewDecoder(res.Body).Decode(&available)
				//fmt.Println("Status::::", res.Status)
				size := len(available.Orders)
				if size < b.ID {
					/* 	//stop bidding
					if !b.Run {
						b.Service.Stop()
						break
					} */
					count++
					if count > 100 {
						wd.Refresh()
						count = 0
					}

					continue Polling
				}

				req.URL, _ = url.Parse(fmt.Sprintf("https://essayshark.com/writer/orders/ping.html?order=%s", available.Orders[size-b.ID].ID))

				//ping the order 3 times
				client.Do(req)
				client.Do(req)
				//client.Do(req)

				req.URL, _ = url.Parse(ordersURL)

				order := available.Orders[size-b.ID]
				orderNo := order.ID

				mlock.Lock()
				if _, ok := AssignedOrders[orderNo]; ok {
					if order.OutDated != "Y" {
						mlock.Unlock()
						continue Polling
					}

				}

				AssignedOrders[orderNo] = orderNo
				mlock.Unlock()

				//var ping Ping
				///**
				// *  Filters section
				// */
				client := &http.Client{}
				discardReq, err := http.NewRequest("GET", fmt.Sprintf("https://essayshark.com/writer/orders/aj_source.html?act=discard&nobreath=0&id=%s", orderNo), bytes.NewBuffer([]byte("")))
				discardReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
				discardReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})

				title := order.Discipline2AR.Title
				serviceType := order.ServiceType.Slug
				noOfPages, _ := strconv.ParseInt(order.Pages, 10, 32)
				budget := order.Amount
				bidAmount := budget / float32(noOfPages)

				if _, ok := b.Account.ExDisciplines[formatText(title)]; ok {
					client.Do(discardReq)
					//remove the order
					continue Polling

				}

				if b.Account.OrderDetails.DiscardAssignments && serviceType == "assignment" {
					client.Do(discardReq)
					//remove the order
					continue Polling
				}

				if b.Account.OrderDetails.DiscardEditting && serviceType == "editing_rewriting" {
					client.Do(discardReq)
					//remove the order
					continue Polling
				}

				minPages := b.Account.OrderDetails.MinPages
				maxPages := b.Account.OrderDetails.MaxPages
				if int(noOfPages) < minPages || int(noOfPages) > maxPages && maxPages > 0 {
					//discard
					client.Do(discardReq)
					//remove the order
					continue Polling
				}

				completeOrders, _ := strconv.ParseInt(order.CustomerOrder, 10, 32)
				if int(completeOrders) < b.Account.CustomerDetails.CompleteOrders {
					//discard
					client.Do(discardReq)
					//remove the order
					continue Polling

				}

				custRating, _ := strconv.ParseFloat(order.CustomerRating, 64)
				if b.Account.CustomerDetails.DiscardNoRatings && custRating == 0 {
					//discard
					client.Do(discardReq)
					//remove the order
					continue Polling
				}

				if float32(custRating) < b.Account.CustomerDetails.MinRating {
					//discard
					client.Do(discardReq)
					//remove the order
					continue Polling
				}

				if b.Account.CustomerDetails.DiscardOfflineCust && order.OnlineStatus == "offline" {
					//discard
					client.Do(discardReq)
					//remove the order
					continue Polling
				}

				if b.Account.CustomerDetails.DiscardNewCust && order.NewCustomer == "Y" {
					//discard
					client.Do(discardReq)
					//remove the order
					continue Polling
				}
				//assumes time in seconds
				deadline, _ := strconv.ParseInt(order.Deadline, 10, 64)

				minDeadline := b.Account.OrderDetails.MinDeadline
				maxDeadline := b.Account.OrderDetails.MaxDeadline

				minTime := time.Now().Add(time.Duration(minDeadline) * time.Second)
				maxTime := time.Now().Add(time.Duration(maxDeadline) * time.Second)

				td := time.Unix(deadline, 0)
				//fmt.Println("Deadline::::", td, orderNo)
				if td.Before(minTime) || td.After(maxTime) && maxDeadline > 0 {
					//discard
					client.Do(discardReq)
					//remove the order
					continue Polling
				}

				if order.OutDated == "Y" {
					amount := fmt.Sprintf("%.2f", bidAmount)
					form := url.Values{}
					form.Add("bid_add_ua", "mmmmmm")
					form.Add("bid_add", "1")
					form.Add("bid", amount)

					orderURL := "https://essayshark.com/writer/orders/" + orderNo + ".html"
					bidReq, _ := http.NewRequest("POST", orderURL, strings.NewReader(form.Encode()))
					bidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					bidReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
					bidReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
					client.Do(bidReq)

					mlock.Lock()
					if _, ok := AssignedOrders[orderNo]; ok {
						delete(AssignedOrders, orderNo)
					}
					mlock.Unlock()
					continue Polling

				}

				orderURL := "https://essayshark.com/writer/orders/" + orderNo + ".html"
				fmt.Printf("[%d]Opening--->%s\n", b.ID, orderURL)
				wd.Get(orderURL)

				//Check for recommended bid amount
				//	var amount string
				//	wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
				//		elem, _ := wd.FindElement(selenium.ByID, "rec_bid")
				//		if elem != nil {
				//			elem, err = elem.FindElement(selenium.ByID, "rec_amount")
				//			if elem != nil {
				//				rec, _ := elem.Text()
				//				if rec != "" {
				//					amount = rec
				//					fmt.Println("Rec-amount:::", rec)
				//				}
				//			}
				//
				//		}
				//		return amount != "", nil
				//	}, 5*time.Second)
				//
				//	if amount == "" {
				//		amount = fmt.Sprintf("%.2f", bidAmount)
				//
				//	}
				amount := fmt.Sprintf("%.2f", bidAmount)

				wd.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
					elem, _ = wd.FindElement(selenium.ByCSSSelector, ".paper_instructions_view")
					if elem != nil {
						elem, err = elem.FindElement(selenium.ByXPATH, "//a[contains (@target,'download_ifm')]")
						if elem != nil {
							return true, nil
						}
					}

					return false, nil
				}, 5*time.Second, 10*time.Millisecond)

				if elem != nil {
					wd.ExecuteScript("scroll(2000, 200)", nil)
					if err = elem.Click(); err != nil {
						//unable to donwload file
					}
				}

				//var bg sync.WaitGroup
				for i := 0; i < 1; i++ {
					//launch bidding subroutines
					//bg.Add(1)
					go func(orderNo string, amount string, orderURL string) {
						//defer bg.Done()
						client = &http.Client{}
						pingReq, _ := http.NewRequest("GET", fmt.Sprintf("https://essayshark.com/writer/orders/ping.html?order=%s", orderNo), bytes.NewBuffer([]byte("")))
						pingReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
						pingReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
						var ping Ping

						res, _ := client.Do(pingReq)
						json.NewDecoder(res.Body).Decode(&ping)
						var timeRemain int
						if res.Status == "520" {
							timeRemain = 520
						} else {
							timeRemain = ping.TimeRemain
						}

						form := url.Values{}
						form.Add("bid_add_ua", "mmmmmm")
						form.Add("bid_add", "1")
						form.Add("bid", amount)

						for {

							if timeRemain < 11 {
								for i := 0; i < 60; i++ {
									go func() {
										client := &http.Client{}
										bidReq, _ := http.NewRequest("POST", orderURL, strings.NewReader(form.Encode()))
										bidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
										bidReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
										bidReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
										res, _ := client.Do(bidReq)
										fmt.Println(res.Status)
									}()
									time.Sleep(100 * time.Millisecond)
								}

								break

							} else {
								res, _ := client.Do(pingReq)
								json.NewDecoder(res.Body).Decode(&ping)
								//if res.Status != "520" {
								timeRemain = ping.TimeRemain

								if timeRemain < 11 {
									time.Sleep(8 * time.Second)
								}
								//	}
							}

						}

					}(orderNo, amount, orderURL)
				}
				//bg.Wait()

				//send a message
				//wd.Get(orderURL)
				//wd.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
				//	elem, err = driver.FindElement(selenium.ByID, "id_body")
				//	btn, _ := driver.FindElement(selenium.ByID, "id_send_message")
				//	if elem != nil && btn != nil {
				//		if b.Account.Message != "" {
				//			elem.SendKeys(b.Account.Message)
				//			btn.Click()
				//		}
				//
				//		return true, nil
				//	}
				//	return false, nil
				//}, 5*time.Second)
				continue Polling

			}
		}()

	}

	wg.Wait()
}
