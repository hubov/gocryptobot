package main

import (
    "github.com/hubov/gocryptobot/internal/binance"
    "time"
    "log"
    "fmt"
)

func main() {
    defaultTimeout := time.Second * 10
    client := binance.ApiClient(defaultTimeout)
    wallet, err := client.GetWallet()
    if err != nil {
        log.Fatal(err)
    }
    for i, coin := range wallet {
        if (i >= 0) {
            fmt.Println(coin)
        }
    }
}