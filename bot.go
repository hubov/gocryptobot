package main

import (
    "github.com/hubov/gocryptobot/internal/binance"
    "github.com/hubov/gocryptobot/internal"
    "time"
    "log"
    "fmt"
)

type (
    Candle interface {
        set()
    }
)

func main() {
    defaultTimeout := time.Second * 10
    client := binance.ApiClient(defaultTimeout)
    wallet, err := client.SpotBalance()
    if err != nil {
        log.Fatal(err)
    }
    for _, coin := range wallet {
        fmt.Println(coin)
    }
    candles, err := client.GetCandles()
    if err != nil {
        log.Fatal(err)
    }
    for i, candle := range candles {
        if (i >= 0) {
            fmt.Println(candle)
            // Indicators.Candle.Set()
        }
    }
    inds := indicators.SetCandles(candles)
    fmt.Println(inds.EMA(12, "close"))
}