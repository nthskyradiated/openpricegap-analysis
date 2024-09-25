package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"
)

type Stock struct {
	Ticker string
	Gap float64
	OpeningPrice float64
}

type Position struct {
	EntryPrice float64
	Shares int
	TakeProfitPrice float64
	StopLossPrice float64
	Profit float64
}

type Selection struct {
	Ticker string
	Position
	Articles []Article
}

type attributes struct {
	PublishOn time.Time `json:"publishOn"`
	Title string `json:"title"`
}

type seekingAlphaNews struct {
	Attributes attributes `json:"attributes"`
}

type SeekingAlphaResponse struct {
	Data []seekingAlphaNews `json:"data"`
}

type Article struct {
	PublishOn time.Time
	Headline string
}

var accountBalance = 10000.00
var lossTolerance = 0.02
var maxLossPerTrade = accountBalance * lossTolerance
var profitPercent = .8


var url = os.Getenv("URL")
var apiKeyHeader = os.Getenv("APIKEYHEADER")
var apiKey = os.Getenv("APIKEY")




func main() {
	stocks, err := Load("./opg.csv")
	if err != nil {
		log.Fatal(err)
		return
	}

	stocks = slices.DeleteFunc(stocks, func(s Stock) bool {
		return math.Abs(s.Gap) < .1 })

	var selections []Selection

	selectionsChan := make(chan Selection, len(stocks))

	for _, stock := range stocks {

		go func(s Stock, selected chan <- Selection) {

			position := Calculate(stock.Gap, stock.OpeningPrice)
	
			articles, err := FetchNews(stock.Ticker)
			if err != nil {
				log.Printf("error loading news about %s, %v", stock.Ticker, err)
				selected <- Selection{}
			} else {
				log.Printf("Found %d articles about %s", len(articles), stock.Ticker)
			}
			sel := Selection {
				Ticker: stock.Ticker,
				Position: position,
				Articles: articles,
			}

			selected <- sel
			
			}(stock, selectionsChan)

		}
		for sel := range selectionsChan {
			selections = append(selections, sel)
			if len(selections) == len(stocks) {
				close(selectionsChan)
			}

		}

	outputPath := "./opg.json"
	err = Deliver(outputPath, selections)
	if err != nil {
		log.Printf("Error writing output, %v", err)
		return
	}
	log.Printf("Finished writing output to %s\n", outputPath)
}

func Load(path string) ([]Stock, error){
	opg, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer opg.Close()

	r := csv.NewReader(opg)
	rows, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
		return nil, err 
	}

	rows = slices.Delete(rows, 0, 1)

	var stocks []Stock

	for _, row := range rows {
		ticker := row[0]

		gap, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}

		openingPrice, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			continue
		}


		stocks = append(stocks, Stock{
			Ticker: ticker,
			Gap: gap,
			OpeningPrice: openingPrice,
		})
	}
	return stocks, err
}

func Calculate(gapPercent, openingPrice float64) Position {

	closingPrice := openingPrice / (1 + gapPercent)
	gapValue := closingPrice - openingPrice
	profitFromGap := profitPercent * gapValue

	stopLoss := openingPrice - profitFromGap
	takeProfit := openingPrice + profitFromGap

	shares := int(maxLossPerTrade / math.Abs(stopLoss - openingPrice))

	profit := math.Abs(openingPrice - takeProfit) * float64(shares)
	profit = math.Round(profit*100) / 100

	return Position{
		EntryPrice: math.Round(openingPrice*100) / 100,
		Shares: shares,
		TakeProfitPrice: math.Round(takeProfit*100) / 100,
		StopLossPrice: math.Round(stopLoss*100) / 100,
		Profit: math.Round(profit*100) / 100,
	}
}

func FetchNews(ticker string) ([]Article, error) {
	req, err := http.NewRequest(http.MethodGet, url+ticker, nil )
	if err != nil {
		return nil, err
	}
	req.Header.Add(apiKeyHeader, apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("error Code Received %d", res.StatusCode)
	}

	resp := &SeekingAlphaResponse{}
	json.NewDecoder(res.Body).Decode(resp)

	var articles []Article

	for _, item := range resp.Data {
		art := Article {
			PublishOn: item.Attributes.PublishOn,
			Headline: item.Attributes.Title,
		}
		articles = append(articles, art)
	}
	return articles, nil
}

func Deliver(filePath string, selections []Selection) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(selections)
	if err != nil {
		return fmt.Errorf("error encoding selections: %w", err)
	}
	return nil
}