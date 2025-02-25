package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

const url = "https://store.steampowered.com/app/2050650/Resident_Evil_4"

func main() {

	browser := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.MaxDepth(1),
		colly.UserAgent(browser.Headers.Get("user-agent")),
	)

	collector.SetClient(&http.Client{
		Transport: browser.Transport,
	})



	collector.OnHTML(".page_content_ctn", func(e *colly.HTMLElement) {
		// Extrae datos de elementos HTML
		title := e.ChildText("div#appHubAppName")
		
		price := "No disponible"  // Valor por defecto
		if prices := e.ChildTexts("div.game_purchase_price.price"); len(prices) > 0 {
    		price = prices[0]
		}

		// Limpia los datos extraídos
		title = strings.TrimSpace(title)
		price = strings.TrimSpace(price)
		// Imprime los datos extraídos
		fmt.Print("titulo: ", title, "  precio: ", price)
	})

	collector.OnRequest(func(r *colly.Request) {
		log.Println("Visitando", r.URL)
	})

	err := collector.Visit(url)
	if err != nil {
		log.Fatal(err)
	}

}
