package strategy

import (
	"github.com/hubov/gocryptobot/internal/binance"
    "github.com/cinar/indicator"
	"fmt"
	"time"
	"log"
)

var candles []binance.Candle

func init() {
	defaultTimeout := time.Second * 10
    client := binance.ApiClient(defaultTimeout)
    wallet, err := client.SpotBalance()
    if err != nil {
        log.Fatal(err)
    }
    for _, coin := range wallet {
        fmt.Println(coin)
    }
    candles, err = client.GetCandles()
    if err != nil {
        log.Fatal(err)
    }
    // for _, candle := range candles {
    //     fmt.Println(candle)
    // }
}

func GetValues(period int, periodType string) (result []float64) {
    var price float64
    len := len(candles)
    i := len - period
    for i < len {
        switch (periodType) {
        case "close": 
            price = candles[i].Close
        case "open":
            price = candles[i].Open
        case "low":
            price = candles[i].Low
        case "high": 
            price = candles[i].High
        }

        result = append(result, price)
        i++
    }

    return
}

func SignalBuy() (result bool) {
    var tests []bool
    result = true

    data := GetValues(30, "close")
    dataLen := len(data)
    sma := indicator.Sma(30, data)

    if sma[len(sma)-1] < data[dataLen-1] {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    data = GetValues(500, "close")
    dataLen = len(data)
    _, rsi := indicator.Rsi(2, data)
    fmt.Println(rsi)
    rsiLen := len(rsi)

    fmt.Println(rsi[rsiLen-2], rsi[rsiLen-1])
    if rsi[rsiLen-2] < 5 && rsi[rsiLen-1] >= 5 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    for _, test := range tests {
        if test == false {
            result = false
        }
    }
    return
}

func SingalSell() bool {
    return false
}

func GetSignal() string {
    if SignalBuy() {
        return "buy"
    } else if SingalSell() {
        return "sell"
    } else {
        return "wait"
    }

}