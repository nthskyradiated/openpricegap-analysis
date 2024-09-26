package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nthskyradiated/openpricegap-analysis/cmd"
	"github.com/nthskyradiated/openpricegap-analysis/internal/news"
	"github.com/nthskyradiated/openpricegap-analysis/internal/pos"
	"github.com/nthskyradiated/openpricegap-analysis/internal/raw"
	"github.com/nthskyradiated/openpricegap-analysis/internal/trade"
	"github.com/nthskyradiated/openpricegap-analysis/pkg/csv"
	"github.com/nthskyradiated/openpricegap-analysis/pkg/json"
	"github.com/nthskyradiated/openpricegap-analysis/pkg/process"
	"github.com/nthskyradiated/openpricegap-analysis/pkg/seekingalpha"
)


func main() {
	err := godotenv.Load() // 👈 load .env file
	if err != nil {
		log.Fatal(err)
	}
	
	var seekingAlphaURL = os.Getenv("URL")
	var apiKey = os.Getenv("APIKEY") 

		// Validate environment variables
		if seekingAlphaURL == "" {
			fmt.Println("Missing variable: URL")
			os.Exit(1)
		}

		if apiKey == "" {
			fmt.Println("Missing variable APIKEY")
			os.Exit(1)
		}
		
	inputPath := flag.String("i", "", "Path to input file (required)")
	accountBalance := flag.Float64("b", 0.0, "Account balance (required)")
	outPath := flag.String("o","./opg.json", "Path to outputfile")
	lossTolerance := flag.Float64("l", 0.02, "Loss tolerance percentage")
	profitPercent := flag.Float64("p", 0.8, "Percentage of the gap to take as profit")
	minGap := flag.Float64("m", 0.1, "Minimum gap value to consider")

	flag.Parse()

	if *inputPath =="" || *accountBalance == 0.0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var loader raw.Loader = csv.NewLoader(*inputPath)
	var f raw.Filterer = process.NewFilterer(*minGap)
	var c pos.Calculator = process.NewCalculator(*accountBalance, *lossTolerance, *profitPercent)
	var fet news.Fetcher = seekingalpha.NewClient(seekingAlphaURL, apiKey)
	var del trade.Deliverer = json.NewDeliverer(*outPath)

	err = cmd.Run(loader, f, c, fet, del)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}