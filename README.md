A set of functions that scrapes the PSE Edge Website (http://edge.pse.com.ph) for stock lists and data which can be used to notify a user via email.

Current State:

Functions needed to retrieve necessary stock data are running. These are: 

- findStockInfo: Generate a list of stocks and their information that partially match a query symbol
- getStockData: Scrape the stock table from the stock data page of a given stock symbol
- parseStockData: Convert a scraped stock table into a golang struct
- notifyStockDataWatcher: Send an email notifying a user with the stocks data
  
A sample using the STI stock symbol of how the functions are to be used to create a stock price notification can be seen in main().

To Install:

1. Install colly and goquery
   
    `go get -u github.com/gocolly/colly/...`\
    `go get github.com/PuerkitoBio/goquery`  
2. Replace the email accounts with your own, generate an app token if needed
3. build and run main.go
