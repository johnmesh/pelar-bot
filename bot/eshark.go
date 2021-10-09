package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
	"github.com/tebeka/selenium"
)

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func FormatText(s string) string {
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
	Token   string
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
	Token           string `bson:"token"`
}

type Message struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Time string `json:"time"`
}

type Text struct {
	Event       string `json:"event"`
	Order       string `json:"order"`
	Status_From int    `json:"status_from"`
	Status_To   int    `json:"status_to"`
	Writer      string `json:"writer"`
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

type OrderInfo struct {
	Amount string
}
type Ping struct {
	TimeRemain  int `json:"read_time_remain"`
	FilesRemain int `json:"files_download_remain"`
}

func getAccount(account *Account, email string) (err error) {
	err = FetchAccount(email, account)
	return
}

func Init(email string, token string) {

	selenium.SetDebug(false)

	var isBotRunnig = false
	var bidders []*Bidder

	for {
		//sync data
		var account Account
		//err := getAccount(&account, email)
		//if err != nil {
		//	fmt.Println("Error:::", err)
		//}
		//token := account.Token
		account.Status = "on"
		account.OrderDetails.DiscardAssignments = true
		account.OrderDetails.DiscardEditting = true
		account.OrderDetails.MinDeadline = 21600

		fmt.Println("Account:::", account.Email, account.Password, account.Status)

		if account.Status == "on" && !isBotRunnig {
			//start bidding
			exDisciplines := make(map[string]string)
			for _, v := range account.OrderDetails.ExDiscipline {
				d := FormatText(v)
				exDisciplines[d] = v
			}
			account.ExDisciplines = exDisciplines

			bidder := &Bidder{
				ID:      1,
				Account: account,
				Run:     true,
				Token:   token,
			}

			bidders = append(bidders, bidder)
			go bidder.Start()

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

	client := &http.Client{}
	//ordersURL := "https://essayshark.com/writer/orders/aj_source.html?act=load_list&nobreath=1&session_more_qty=0&session_discarded=0&_=1629218589134"
	//req, _ := http.NewRequest("GET", "", bytes.NewBuffer([]byte("")))
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
	//req.AddCookie(&http.Cookie{Name: "a11nt3n", Value: b.Token})
	//req.URL, _ = url.Parse(ordersURL)
	//var available AvailableItems
	//
	//for {
	//
	//	res, err := client.Do(req)
	//	if err != nil {
	//		panic(err)
	//	}
	//	json.NewDecoder(res.Body).Decode(&available)
	//	if len(available.Orders) == 0 {
	//		break
	//	}
	//
	//	var od []string
	//	for i := 0; i < len(available.Orders); i++ {
	//		od = append(od, available.Orders[i].ID)
	//	}
	//	ids := strings.Join(od, ",")
	//
	//	form := url.Values{}
	//	form.Add("act", "discard_all")
	//	form.Add("nobreath", "1")
	//	form.Add("ids", ids)
	//
	//	discardAllReq, _ := http.NewRequest("POST", "https://essayshark.com/writer/orders/aj_source.html", strings.NewReader(form.Encode()))
	//	discardAllReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//	discardAllReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
	//	discardAllReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: b.Token})
	//	_, err = client.Do(discardAllReq)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//}
	fmt.Printf("[%d]:Listening... \n", b.ID)

	pingReq, _ := http.NewRequest("GET", "", bytes.NewBuffer([]byte("")))
	pingReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
	pingReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: b.Token})

	for {

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		u := url.URL{Scheme: "ws", Host: "box1.essayshark.com", Path: "/live/ws/to_writers"}
		log.Printf("connecting to %s", u.String())

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			defer c.Close()
			for {
				_, data, err := c.ReadMessage()

				if err != nil {
					log.Println("read:", err)
					if err.Error() == "websocket: close 1006 (abnormal closure): unexpected EOF" {
						panic(err)
					}
					return
				}

				go func(data []byte, pReq *http.Request) {
					fmt.Println("NEW-EVENT:::", string(data))

					var message Message
					var text Text

					json.Unmarshal(data, &message)
					json.Unmarshal([]byte(message.Text), &text)

					orderNo := text.Order

					if text.Event != "orders_change_status" || orderNo == "" {
						return

					}

					if text.Status_From != 0 || text.Status_To < 20 {
						return
					}

					var pingReq = *pReq
					pingReq.URL, _ = url.Parse(fmt.Sprintf("https://essayshark.com/writer/orders/ping.html?order=%s", orderNo))

					//ping the order 2 times
					var ping Ping
					client.Do(&pingReq)
					res, _ := client.Do(&pingReq)

					err := json.NewDecoder(res.Body).Decode(&ping)
					fmt.Println("PING-STATUS:::", res.Status, orderNo, ping.TimeRemain)

					if err != nil {
						//panic(err)
						fmt.Println("error::", err, orderNo)
						return
					}

					mlock.Lock()
					if _, ok := AssignedOrders[orderNo]; ok {
						mlock.Unlock()
						return
					}

					AssignedOrders[orderNo] = orderNo
					mlock.Unlock()

					ordersURL := "https://essayshark.com/writer/orders/aj_source.html?act=load_list&nobreath=1&session_more_qty=0&session_discarded=0&_=1629218589134"
					req, _ := http.NewRequest("GET", "", bytes.NewBuffer([]byte("")))
					req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
					req.AddCookie(&http.Cookie{Name: "a11nt3n", Value: b.Token})
					req.URL, _ = url.Parse(ordersURL)

					var available AvailableItems
					var order Order

					res, err = client.Do(req)
					json.NewDecoder(res.Body).Decode(&available)

					for i := 0; i < len(available.Orders); i++ {
						if available.Orders[i].ID == orderNo {
							order = available.Orders[i]
							break
						}
					}

					if order.ID == "" {
						return
					}

					//var ping Ping
					///**
					// *  Filters section
					// */
					client := &http.Client{}
					discardReq, _ := http.NewRequest("GET", fmt.Sprintf("https://essayshark.com/writer/orders/aj_source.html?act=discard&nobreath=0&id=%s", orderNo), bytes.NewBuffer([]byte("")))
					discardReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
					discardReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: b.Token})

					//title := order.Discipline2AR.Title
					serviceType := order.ServiceType.Slug
					noOfPages, _ := strconv.ParseInt(order.Pages, 10, 32)
					budget := order.Amount
					bidAmount := budget / float32(noOfPages)

					//if _, ok := b.Account.ExDisciplines[formatText(title)]; ok {
					//	client.Do(discardReq)
					//	//remove the order
					//	continue Polling
					//
					//}

					if b.Account.OrderDetails.DiscardAssignments && serviceType == "assignment" {
						client.Do(discardReq)
						//remove the order
						return
					}

					if b.Account.OrderDetails.DiscardEditting && serviceType == "editing_rewriting" {
						client.Do(discardReq)
						//remove the order
						return
					}

					//minPages := b.Account.OrderDetails.MinPages
					//maxPages := b.Account.OrderDetails.MaxPages
					//if int(noOfPages) < minPages || int(noOfPages) > maxPages && maxPages > 0 {
					//	//discard
					//	client.Do(discardReq)
					//	//remove the order
					//	continue Polling
					//}
					//
					//completeOrders, _ := strconv.ParseInt(order.CustomerOrder, 10, 32)
					//if int(completeOrders) < b.Account.CustomerDetails.CompleteOrders {
					//	//discard
					//	client.Do(discardReq)
					//	//remove the order
					//	continue Polling
					//
					//}
					//
					//custRating, _ := strconv.ParseFloat(order.CustomerRating, 64)
					//if b.Account.CustomerDetails.DiscardNoRatings && custRating == 0 {
					//	//discard
					//	client.Do(discardReq)
					//	//remove the order
					//	continue Polling
					//}
					//
					//if float32(custRating) < b.Account.CustomerDetails.MinRating {
					//	//discard
					//	client.Do(discardReq)
					//	//remove the order
					//	continue Polling
					//}
					//
					//if b.Account.CustomerDetails.DiscardOfflineCust && order.OnlineStatus == "offline" {
					//	//discard
					//	client.Do(discardReq)
					//	//remove the order
					//	continue Polling
					//}
					//
					//if b.Account.CustomerDetails.DiscardNewCust && order.NewCustomer == "Y" {
					//	//discard
					//	client.Do(discardReq)
					//	//remove the order
					//	continue Polling
					//}
					//assumes time in seconds
					deadline, _ := strconv.ParseInt(order.Deadline, 10, 64)

					minDeadline := b.Account.OrderDetails.MinDeadline
					maxDeadline := b.Account.OrderDetails.MaxDeadline

					minTime := time.Now().Add(time.Duration(minDeadline) * time.Second)
					maxTime := time.Now().Add(time.Duration(maxDeadline) * time.Second)

					td := time.Unix(deadline, 0)
					////fmt.Println("Deadline::::", td, orderNo)
					if td.Before(minTime) || td.After(maxTime) && maxDeadline > 0 {
						//discard
						client.Do(discardReq)

						return
					}

					if order.OutDated == "Y" {
						fmt.Println("Out-dated:::", orderNo)
						amount := fmt.Sprintf("%.2f", bidAmount)
						form := url.Values{}
						form.Add("bid_add_ua", "mmmmmm")
						form.Add("bid_add", "1")
						form.Add("bid", amount)

						orderURL := "https://essayshark.com/writer/orders/" + orderNo + ".html"
						bidReq, _ := http.NewRequest("POST", orderURL, strings.NewReader(form.Encode()))
						bidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						bidReq.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
						bidReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: b.Token})
						client.Do(bidReq)
						return

					}

					if ping.FilesRemain != 0 {
						var orderInfo OrderInfo
						err = DownloadFile(orderNo, b.Token, &orderInfo, &ping)
					}

					orderURL := "https://essayshark.com/writer/orders/" + orderNo + ".html"
					fmt.Printf("[%d]Bidding--->%s\n", b.ID, orderURL)

					amount := fmt.Sprintf("%.2f", bidAmount)
					fmt.Println("Order-amount:::", amount)

					//Prepare to bid
					form := url.Values{}
					form.Add("bid_add_ua", "mmmmmm")
					form.Add("bid_add", "1")
					form.Add("bid", amount)

					var requests []*http.Request
					for i := 0; i < 5; i++ {
						bidReq, _ := http.NewRequest("POST", orderURL, strings.NewReader(form.Encode()))
						bidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						bidReq.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
						bidReq.AddCookie(&http.Cookie{Name: "a11nt3n", Value: b.Token})
						requests = append(requests, bidReq)
					}

					if ping.TimeRemain == 0 {
						res, _ := client.Do(requests[0])
						fmt.Println(res.Status, orderNo)
						return
					}

					//Wait for the countdown
					for {
						res, _ := client.Do(&pingReq)
						json.NewDecoder(res.Body).Decode(&ping)

						if ping.TimeRemain > 20 {
							time.Sleep(5 * time.Second)
						}

						if ping.TimeRemain < 11 {
							time.Sleep(8500 * time.Millisecond)
							break
						}
					}

					//Start bidding
					for i := 0; i < len(requests); i++ {
						go func() {
							client.Do(&pingReq)
						}()

						go func(i int) {
							res, _ := client.Do(requests[i])
							fmt.Println(res.Status, orderNo)

							mlock.Lock()
							if _, ok := AssignedOrders[orderNo]; ok {
								delete(AssignedOrders, orderNo)
							}
							mlock.Unlock()
						}(i)
						time.Sleep(100 * time.Millisecond)

					}

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
				}(data, pingReq)

			}

		}()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
				if err != nil {
					log.Println("write:", err)
					return
				}
			case <-interrupt:
				log.Println("interrupt")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}

	}

}

func DownloadFile(orderNo string, token string, orderInfo *OrderInfo, ping *Ping) error {

	client := &http.Client{}
	orderURL := fmt.Sprintf("https://essayshark.com/writer/orders/%s.html", orderNo)
	req, _ := http.NewRequest("GET", orderURL, bytes.NewBuffer([]byte("")))
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
	//req.Header.Add("Content-Type", "text/html")
	req.AddCookie(&http.Cookie{Name: "a11nt3n", Value: token})
	res, _ := client.Do(req)

	doc, _ := goquery.NewDocumentFromResponse(res)

	var fileURL string

	doc.Find(".paper_instructions_view").Each(func(i int, s *goquery.Selection) {
		var att string
		s.Children().Each(func(i int, s *goquery.Selection) {
			att, ok := s.Attr("href")
			if ok && att != "" {
				fileURL = att
				return
			}
		})

		if att != "" {
			return
		}

	})

	downloadURL := fmt.Sprintf("https://essayshark.com%s", fileURL)
	req.URL, _ = url.Parse(downloadURL)
	res, _ = client.Do(req)

	fmt.Println("DownloadFile:::", fileURL)

	return nil

}
