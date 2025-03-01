package main

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

var (
	url        = "https://store.steampowered.com/app/2050650/Resident_Evil_4/"
	sleepTime  = 5 * time.Minute
	maxRetries = 3
	browser    = req.DefaultClient().ImpersonateChrome()
)

type Data struct {
	Title    string
	Price    string
	Discount string
	Error    error
}

func main() {

	count := 0
	for {
		inf := DataCollector(browser)
		log.Println(inf.Title, " ", inf.Price)
		if inf.Error != nil {
			SendAlert(inf.Error.Error(), inf.Error.Error())
			count += 1
			if count == 8 {
				break
			}
		}
		if inf.Discount != "0%" {
			SendAlert(inf.Price, inf.Title)
			count += 1
			if count == 8 {
				break
			}
		}
		time.Sleep(sleepTime)
	}
}

func SendAlert(price string, title string) error {
	from := os.Getenv("EMAIL")
	password := os.Getenv("PASS") // Usa contraseñas de aplicación para Gmail
	to := []string{"govannytgoz@gmail.com"}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	status := DataCollector(browser)
	var message []byte
	if status.Error != nil {
		message = []byte("From: " + from + "\r\n" +
			"To: " + to[0] + "\r\n" +
			"Subject: Errorr\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" + // Añadimos soporte para HTML
			"\r\n" +
			"<html><body>" +
			"<h2 style='font-size: 20px;'>" + title + "</h2>" + // Aquí se agrega el título
			"<p>Ocurrio un error: " + price + "</p>" +
			"</body></html>")
	} else {
		message = []byte("From: " + from + "\r\n" +
			"To: " + to[0] + "\r\n" +
			"Subject: ¡Nuevo precio detectado!\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" + // Añadimos soporte para HTML
			"\r\n" +
			"<html><body>" +
			"<h2 style='font-size: 20px;'>" + title + "</h2>" + // Aquí se agrega el título
			"<p>Descuento de: " + price + "</p>" +
			"</body></html>")
	}
	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		from,
		to,
		message,
	)
}
func DataCollector(browser *req.Client) Data {
	data := Data{}
	collector := colly.NewCollector(
		colly.MaxDepth(maxRetries),
		colly.UserAgent(browser.Headers.Get("user-agent")),
	)
	collector.SetClient(&http.Client{
		Transport: browser.Transport,
	})
	collector.OnHTML(".page_content_ctn", func(e *colly.HTMLElement) {
		title := e.ChildText("div#appHubAppName")
		price := e.DOM.Find(".game_purchase_price.price").First()
		princeDiscount := e.DOM.Find(".discount_final_price").First()
		if discount := e.ChildText(".discount_pct"); discount != "" {
			data.Discount = strings.TrimSpace(discount)
			data.Price = strings.TrimSpace(princeDiscount.Text())
		} else {
			data.Discount = "0%"
			data.Price = strings.TrimSpace(price.Text())
		}
		data.Title = strings.TrimSpace(title)
		if data.Title == "" {
			data.Error = fmt.Errorf("no se encontró el título")
		}
		if data.Price == "" {
			data.Error = fmt.Errorf("no se encontró el precio")
		}
		if data.Discount == "" {
			data.Error = fmt.Errorf("no se encontro el descuento")
		}
	})
	collector.OnRequest(func(r *colly.Request) {
		log.Println("Visitando", r.URL)
	})
	err := collector.Visit(url)
	if err != nil {
		data.Error = err
	}
	return data

}
