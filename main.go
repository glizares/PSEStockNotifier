package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

type StockData struct {
	LastTradedPrice float64
	Change          string
	Value           int
	Volume          int
	High52          float64
	Open            float64
	High            float64
	Low             float64
	Average         float64
	Low52           float64
	PrevCloseDate   string
	PERatio         string
	PERatioSec      string
	BookVal         float64
	PBVRation       string
}

func main() {
	scraper := colly.NewCollector(
		colly.AllowedDomains("edge.pse.com.ph", "pse.com.ph"),
	)
	scraper.OnHTML("table", func(e *colly.HTMLElement) {
		firstColInRow := e.DOM.Find("tr").Text()
		fmt.Println(firstColInRow)
		switch firstColInRow {
		case "Last Traded Price":
			fmt.Println("INHREE")
		}

	})

	scraper.Visit("http://edge.pse.com.ph/companyPage/stockData.do?cmpy_id=222")
}
