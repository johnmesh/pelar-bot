package bot

import (
	"bytes"
	"context"
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
var mlock = &sync.Mutex{}

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

type Worker struct {
	wd      *selenium.WebDriver
	ID      int
	Token   string
	Account Account
}

type Filters struct {
	OrderDetails    OrderDetails    `bson:"orderDetails"`
	CustomerDetails CustomerDetails `bson:"customerDetails"`
	BiddingPrice
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
}
type AvailableItems struct {
	Orders []Order `json:"new_items"`
}

type Ping struct {
	TimeRemain  int `json:"read_time_remain"`
	FilesRemain int `json:"files_download_remain"`
}

func getAccount(account *Account, email string) (err error) {
	client, _ := Connect()
	defer client.Disconnect(context.Background())

	err = GetAccount(client, email, account)
	return
}

func Init() {
	const (
		seleniumPath     = "/vendor/selenium-server-standalone-4.0.0-alpha-1.jar"
		chromeDriverPath = "/vendor/chromedriver_92linux"
	)
	selenium.SetDebug(false)

	var account Account
	var isBotRunnig = false
	var bidders []*Bidder

	err := getAccount(&account, "johnmesh4@gmail.com")
	if err != nil {
		panic(err)
	}
	//account 1
	//account.Email = "Lydiarugut@gmail.com"
	//account.Password = "hustle hard"
	//account.Message = "Hi, I deliver high-quality and plagiarism free work.Expect great communication and strict compliance with instructions and deadlines"

	//account 2
	//account.Email = "Jacknyangare@yahoo.com"
	//account.Password = "shark attack"
	//account.Message = "Hi, I am a versatile professional research and academic writer, specializing in research papers, essays, term papers, theses, and dissertations. NO PLAGIARISM..."

	//account 3
	account.Email = "nambengeleashap@gmail.com"
	account.Password = "Optimus#On"

	//account 4
	//account.Email = "onderidismus85@gmail.com"
	//account.Password = "my__shark"

	for {
		//fetch db info

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
				p := fmt.Sprintf("401%d", i)
				port, _ := strconv.Atoi(p)
				service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
				if err != nil {
					panic(err) // panic is used only as an example and is not otherwise recommended.
				}
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
				bidders[i].Service.Stop()
				bidders[i].Run = false
			}
			isBotRunnig = false
		} else if account.Status == "on" && isBotRunnig {
			//sync the data
			for i := 0; i < len(bidders); i++ {
				bidders[i].Account = account
			}
		}
		//Repeat after every 5 seconds
		time.Sleep(5 * time.Second)
	}

}

var webDrivers []selenium.WebDriver

func (b *Bidder) Start() {

	const defaultTimeOut = 10 * time.Second
	//var lock = &sync.Mutex{}

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
	for i := 0; i < 5; i++ {
		wg.Add(1)

		//launch pollers subroutines
		go func() {
			defer wg.Done()
			wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", b.Port))

			if err != nil {
				panic(err)
			}
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

			fmt.Printf("[%d]:polling... \n", b.ID)

			client := &http.Client{}
			var available AvailableItems
			ordersURL := "https://essayshark.com/writer/orders/aj_source.html?act=load_list&nobreath=1&session_more_qty=0&session_discarded=0&_=1629218589134"
			req, _ := http.NewRequest("GET", "", bytes.NewBuffer([]byte("")))
			req.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})

			req.URL, _ = url.Parse(ordersURL)
			count := 0
		Polling:
			for {
				res, _ := client.Do(req)
				json.NewDecoder(res.Body).Decode(&available)

				if len(available.Orders) < b.ID {
					count++
					if count > 2000 {
						wd.Refresh()
						count = 0
					}
					continue Polling
				}

				//fmt.Println("Available orders:::", len(available.Orders))

				req.URL, _ = url.Parse(fmt.Sprintf("https://essayshark.com/writer/orders/ping.html?order=%s", available.Orders[b.ID-1].ID))

				//ping the order 3 times
				client.Do(req)
				client.Do(req)
				client.Do(req)

				req.URL, _ = url.Parse(ordersURL)

				orders := available.Orders
				order := orders[b.ID-1]
				orderNo := order.ID

				mlock.Lock()
				if _, ok := AssignedOrders[orderNo]; ok {
					mlock.Unlock()
					continue Polling
				}

				AssignedOrders[orderNo] = orderNo
				mlock.Unlock()

				var ping Ping
				client := &http.Client{}
				discardReq, err := http.NewRequest("GET", fmt.Sprintf("https://essayshark.com/writer/orders/aj_source.html?act=discard&nobreath=0&id=%s", orderNo), bytes.NewBuffer([]byte("")))
				discardReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})

				/**
				 *  Filters section
				 */
				title := order.Discipline2AR.Title
				serviceType := order.ServiceType.Slug
				noOfPages, _ := strconv.ParseInt(order.Pages, 10, 32)
				budget := order.Amount
				bidAmount := budget / float32(noOfPages)

				if _, ok := b.Account.ExDisciplines[formatText(title)]; ok {
					client.Do(discardReq)
					//remove the order
					mlock.Lock()
					if _, ok := AssignedOrders[orderNo]; ok {
						delete(AssignedOrders, orderNo)
					}
					mlock.Unlock()
					continue Polling

				}

				if serviceType == "assignment" || serviceType == "editing_rewriting" {
					client.Do(discardReq)
					//remove the order
					mlock.Lock()
					if _, ok := AssignedOrders[orderNo]; ok {
						delete(AssignedOrders, orderNo)
					}
					mlock.Unlock()
					continue Polling
				}

				minPages := b.Account.OrderDetails.MinPages
				maxPages := b.Account.OrderDetails.MaxPages
				if int(noOfPages) < minPages || int(noOfPages) > maxPages && maxPages > 0 {
					//discard
				}

				completeOrders, _ := strconv.ParseInt(order.CustomerOrder, 10, 32)
				if int(completeOrders) < b.Account.CustomerDetails.CompleteOrders {
					//discard

				}

				custRating, _ := strconv.ParseFloat(order.CustomerRating, 64)
				if b.Account.CustomerDetails.DiscardNoRatings && custRating == 0 {
					//discard
				}

				if float32(custRating) < b.Account.CustomerDetails.MinRating {
					//discard
				}

				if b.Account.CustomerDetails.DiscardOfflineCust && order.OnlineStatus == "offline" {
					//discard
				}

				if b.Account.CustomerDetails.DiscardNewCust && order.NewCustomer == "Y" {
					//discard
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
				}

				orderURL := "https://essayshark.com/writer/orders/" + orderNo + ".html"
				fmt.Printf("[%d]Opening--->%s\n", b.ID, orderURL)

				wd.Get(orderURL)

				/**
				 * Bidding Section
				 */

				var amount string

				//elem, err := wd.FindElement(selenium.ByID, "rec_bid")
				//if elem != nil {
				//	var amt, r string
				//	elem, err = elem.FindElement(selenium.ByID, "rec_amount")
				//	if elem != nil {
				//		rec, _ := elem.Text()
				//		if rec != "" {
				//			r, _ := strconv.ParseFloat(rec, 32)
				//			for i := 0; i < len(b.Account.Bids); i++ {
				//				if b.Account.Bids[i].Rec == float32(r) {
				//					//amt = fmt.Sprintf("%.2f", b.Account.Bids[i].Amount)
				//					break
				//				}
				//			}
				//			fmt.Println("Rec-amount", amt, rec)
				//		}
				//	}
				//	if amt != "" {
				//		amount = amt
				//	} else if amount == "" {
				//		amount = r
				//	}
				//
				//} else {
				//	fmt.Println("error:::no  amount found", orderNo)
				//
				//}

				amount = fmt.Sprintf("%.2f", bidAmount)

				client = &http.Client{}
				pingReq, _ := http.NewRequest("GET", fmt.Sprintf("https://essayshark.com/writer/orders/ping.html?order=%s", orderNo), bytes.NewBuffer([]byte("")))
				pingReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
				res, _ = client.Do(pingReq)
				json.NewDecoder(res.Body).Decode(&ping)

				if ping.FilesRemain != 0 {
					//download atleast one file
					/* filepath :=
					"//div[@class='paper_instructions_view']/a[contains (@data-url-raw,'/writer/get_additional_material.html')]" */
					wd.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
						elem, err = driver.FindElement(selenium.ByXPATH, "//a[contains (@target,'download_ifm')]")
						if elem != nil {
							return true, nil
						}

						return false, nil
					}, 5*time.Second, 1*time.Millisecond)

					if elem == nil {
						fmt.Printf("[%d]No files to donwnload\n", b.ID)
					} else {
						wd.ExecuteScript("scroll(2000, 200)", nil)
						if err = elem.Click(); err != nil {
							//unable to donwload file
						}
					}
				}

				amount = fmt.Sprintf("%.2f", bidAmount)
				fmt.Println("Amount", amount)

				var bg sync.WaitGroup
				for i := 0; i < 10; i++ {
					//launch bidding subroutines
					bg.Add(1)
					go func() {
						defer bg.Done()
						client = &http.Client{}
						pingReq, _ := http.NewRequest("GET", fmt.Sprintf("https://essayshark.com/writer/orders/ping.html?order=%s", orderNo), bytes.NewBuffer([]byte("")))
						pingReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
						var ping Ping

						form := url.Values{}
						form.Add("bid_add_ua", "m")
						form.Add("bid_add", "1")
						form.Add("bid", amount)

						count := 0

						res, _ := client.Do(pingReq)
						json.NewDecoder(res.Body).Decode(&ping)
						timeRemain := ping.TimeRemain

						for {

							if timeRemain < 11 {
								//fmt.Println(amount, orderNo, orderURL)
								bidReq, _ := http.NewRequest("POST", orderURL, strings.NewReader(form.Encode()))
								bidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
								bidReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: auth_token})
								client.Do(bidReq)
								count++
								if count > 30 {
									//remove the order
									mlock.Lock()
									if _, ok := AssignedOrders[orderNo]; ok {
										delete(AssignedOrders, orderNo)
									}
									mlock.Unlock()
									break
								}

							} else {
								res, _ := client.Do(pingReq)
								json.NewDecoder(res.Body).Decode(&ping)
								timeRemain = ping.TimeRemain
							}

						}

					}()
				}
				bg.Wait()

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
