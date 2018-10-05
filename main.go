package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

//StockData contains stock data from the stock table
type StockData struct {
	LastTradedPrice float64
	Change          string
	ChangeUp        bool
	ChangeVal       float64
	ChangePercent   float64
	Value           string
	Volume          string
	High52          float64
	Open            float64
	High            float64
	Low             float64
	Average         float64
	Low52           float64
	PrevCloseDate   string
}

func httpGetRequest(requestURL string) (body string, err error) {
	response, err := http.Get(requestURL)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	body = string(contents)
	return body, err
}
func trimDuplicateSpaces(str string) (trimmedString string) {
	extra := true
	for _, c := range str {
		if unicode.IsSpace(c) {
			if !extra {
				trimmedString += " "
			}
			extra = true
		} else {
			trimmedString += string(c)
			extra = false
		}
	}
	return trimmedString
}
func parseStockData(table *colly.HTMLElement) (stock StockData, err error) {
	table.DOM.Find("th").Each(func(index int, item *goquery.Selection) {
		itemHead := string(trimDuplicateSpaces(item.Text()))
		itemVal := string(trimDuplicateSpaces(item.Next().Text()))
		switch itemHead {
		case "Last Traded Price":
			stock.LastTradedPrice, err = strconv.ParseFloat(itemVal, 64)
		case "Open":
			stock.Open, err = strconv.ParseFloat(itemVal, 64)
		case "Previous Close and Date":
			stock.PrevCloseDate = itemVal
		case "Change(% Change)":
			stock.Change = itemVal
			changeSeparated := strings.Split(itemVal, " ")
			if strings.Contains(strings.ToLower(changeSeparated[0]), "up") {
				stock.ChangeUp = true
			} else {
				stock.ChangeUp = false
			}
			stock.ChangeVal, err = strconv.ParseFloat(string(changeSeparated[1]), 64)
			changeSeparated[2] = strings.Replace(changeSeparated[2], "(", "", -1)
			changeSeparated[2] = strings.Replace(changeSeparated[2], ")", "", -1)
			changeSeparated[2] = strings.Replace(changeSeparated[2], "%", "", -1)
			stock.ChangePercent, err = strconv.ParseFloat(string(changeSeparated[2]), 64)
		case "High":
			stock.High, err = strconv.ParseFloat(itemVal, 64)
		case "Value":
			stock.Value = itemVal
		case "Low":
			stock.Low, err = strconv.ParseFloat(itemVal, 64)
		case "Volume":
			stock.Volume = itemVal
		case "Average Price":
			stock.Average, err = strconv.ParseFloat(itemVal, 64)
		case "52-Week High":
			stock.High52, err = strconv.ParseFloat(itemVal, 64)
		case "52-Week Low":
			stock.Low52, err = strconv.ParseFloat(itemVal, 64)
		}
		if err != nil {
			return
		}
	})
	return stock, err
}
func getStockData(companyID int) (stockData StockData, err error) {
	pseEdgeURL := "http://edge.pse.com.ph"
	companyURL := fmt.Sprintf("%s/companyPage/stockData.do?cmpy_id=%d", pseEdgeURL, companyID)

	scraper := colly.NewCollector(
		colly.AllowedDomains("edge.pse.com.ph", "pse.com.ph"),
	)
	scraper.OnHTML("table", func(tableElement *colly.HTMLElement) {
		firstRowAndCol := tableElement.DOM.Find("tr").Children().First().Text()
		if strings.Contains(firstRowAndCol, "Last Traded Price") {
			stockData, err = parseStockData(tableElement)
		}
	})
	scraper.Visit(companyURL)
	return stockData, err
}
func findStockInfo(symbol string) (companyID int, err error) {
	symbol = strings.ToUpper(symbol)
	requestURL := fmt.Sprintf("http://edge.pse.com.ph/autoComplete/searchCompanyNameSymbol.ax?term=%s", symbol)
	data, err := httpGetRequest(requestURL) 
	if err != nil {
		return -1, nil
	}
	
	fmt.Println(data)
	return -1,nil
}
func main() {
	findStockInfo("st")
	stock, err := getStockData(222)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", stock)
}
