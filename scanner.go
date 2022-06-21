package main

import (
	"github.com/hubov/gocryptobot/internal/binance"
	"time"
	"flag"
	"fmt"
)

func main() {
	base := flag.String("base", "", "filter by base currency")
	quote := flag.String("quote", "", "filter by quote currency")
	interval = "15m"
	intervalMilli = 900000
	

	flag.Parse()

	baseFilter := *base
	quoteFilter := *quote

	timeout := time.Second * 10
	client := binance.ApiClient(timeout)
	availablePairs := client.GetAllMarginPairs()

	i := 0
    for _, pair := range availablePairs {
    	if pair.Base == baseFilter || pair.Quote == quoteFilter {
        	fmt.Println(pair)
        	i++
    	}
    }
    fmt.Println("Count:", i)


}