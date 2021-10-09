package main

import (
	"fmt"
	"os"
	"pelar-bot/bot"
)

func main() {

	email := os.Getenv("EMAIL")
	token := os.Getenv("TOKEN")
	//
	//if email == "" {
	//	log.Fatal("Email or token required")
	//}
	fmt.Println("Account::::", email, token)

	//v1token = "c1v05t3888ea7ef019baf357b95ddd0c"
	bot.Init(email, token)

	//err := getAccount(&account, email)
	//if err != nil {
	//	panic(err)
	//}
	//account 1
	//account.Email = "Lydiarugut@gmail.com"
	//account.Password = "hustle hard"
	//account.Message = "Hi, I deliver high-quality and plagiarism free work.Expect great communication and strict compliance with instructions and deadlines"

	//account 2
	//account.Email = "Jacknyangare@yahoo.com"
	//account.Password = "shark attack"
	//account.Message = "Hi, I am a versatile professional research and academic writer, specializing in research papers, essays, term papers, theses, and dissertations. NO PLAGIARISM..."

	//account 3
	//account.Email = "nambengeleashap@gmail.com"
	//	account.Password = "Optimus#On"
	//account 4
	//account.Email = "onderidismus85@gmail.com"
	//account.Password = "my__shark
	//token := "29v05tae722887f674f1ab9973793964"
	//orderNo := "209892163"
	//
	//client := &http.Client{}
	//orderURL := fmt.Sprintf("https://essayshark.com/writer/orders/%s.html", orderNo)
	//req, _ := http.NewRequest("GET", orderURL, bytes.NewBuffer([]byte("")))
	//req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Mobile Safari/537.36")
	//req.Header.Add("Content-Type", "text/html")
	//req.AddCookie(&http.Cookie{Name: "a11nt3n", Value: token})
	//res, _ := client.Do(req)
	//fmt.Println("DownloadFile:::", res.Status)
	//
	//doc, _ := goquery.NewDocumentFromResponse(res)
	//serviceType, _ := doc.Find(".order-id").Attr("data-title")
	//
	//fmt.Println("serviceType:::", bot.FormatText(serviceType))

}
