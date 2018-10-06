package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

//StockData contains stock data from the stock table
type StockData struct {
	Symbol          string
	LastTradedPrice float64
	Change          string
	ChangeUp        bool
	ChangeVal       float64
	ChangePercent   float64
	Value           string
	Volume          string
	High52          float64
	Open            float64
	PrevClosePrice  float64
	High            float64
	Low             float64
	Average         float64
	Low52           float64
	PrevCloseDate   string
}

//StockInfo contains data retrieved from search stock symbol call
type StockInfo struct {
	CmpyID int    `json:"cmpyId,string"`
	CmpyNm string `json:"cmpyNm"`
	Symbol string `json:"symbol"`
	EtfYn  int    `json:"etfYn,string"`
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
func parseStockData(symbol string, table *colly.HTMLElement) (stock StockData, err error) {
	stock.Symbol = symbol
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
			stock.PrevClosePrice, err = strconv.ParseFloat(string(strings.Split(itemVal, " ")[0]), 64)
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
func getStockData(symbol string) (stockData StockData, err error) {

	stockInfo, err := findStockInfo(symbol)
	if err != nil {
		return stockData, err
	}
	if len(stockInfo) != 1 {
		return stockData, errors.New("Symbol invalid")
	}
	companyID := stockInfo[0].CmpyID
	pseEdgeURL := "http://edge.pse.com.ph"
	companyURL := fmt.Sprintf("%s/companyPage/stockData.do?cmpy_id=%d", pseEdgeURL, companyID)

	scraper := colly.NewCollector(
		colly.AllowedDomains("edge.pse.com.ph", "pse.com.ph"),
	)
	scraper.OnHTML("table", func(tableElement *colly.HTMLElement) {
		firstRowAndCol := tableElement.DOM.Find("tr").Children().First().Text()
		if strings.Contains(firstRowAndCol, "Last Traded Price") {
			stockData, err = parseStockData(symbol, tableElement)
		}
	})
	scraper.Visit(companyURL)
	return stockData, err
}
func findStockInfo(symbol string) (stockInfo []StockInfo, err error) {
	symbol = strings.ToUpper(symbol)
	requestURL := fmt.Sprintf("http://edge.pse.com.ph/autoComplete/searchCompanyNameSymbol.ax?term=%s", symbol)
	data, err := httpGetRequest(requestURL)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(data), &stockInfo)
	if err != nil {
		return nil, err
	}
	return stockInfo, nil
}
func notifyStockDataWatcher(watcher string, stockData StockData) (err error) {
	subject := fmt.Sprintf("PSE Notifier: %s is at target price %f", stockData.Symbol, stockData.LastTradedPrice)

	body := fmt.Sprintf("Symbol: %s\n", stockData.Symbol)
	body += fmt.Sprintf("Last Traded Price: %f\n", stockData.LastTradedPrice)
	body += fmt.Sprintf("Change: %s\n", stockData.Change)
	body += fmt.Sprintf("Open: %f\n", stockData.Open)
	body += fmt.Sprintf("Close: %f\n", stockData.PrevClosePrice)
	body += fmt.Sprintf("High: %f\n", stockData.High)
	body += fmt.Sprintf("Low: %f\n", stockData.Low)
	body += fmt.Sprintf("Average: %f\n", stockData.Average)
	body += fmt.Sprintf("Value: %s\n", stockData.Value)
	body += fmt.Sprintf("Volume: %s\n", stockData.Volume)
	body += fmt.Sprintf("52-Week High: %f\n", stockData.High52)
	body += fmt.Sprintf("52-Week Low: %f\n", stockData.Low52)
	_, err = sendEmail(watcher, subject, body)
	return err

}
func sendEmail(recipient string, subject string, body string) (string, error) {
	from := "replace_me@gmail.com"
	pass := "replace_me_apptoken"
	to := recipient

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return "error", err
	}
	return "Sent Email", nil

}
func main() {
	stockInfo, err := findStockInfo("sti")
	if err != nil {
		fmt.Println(err)
	}
	if len(stockInfo) == 1 {
		stockData, err := getStockData(stockInfo[0].Symbol)
		if err != nil {
			fmt.Println(err)
		}
		err = notifyStockDataWatcher("replace_me@gmail.com", stockData)
		if err != nil {
			fmt.Println(err)
		}
	}
}
