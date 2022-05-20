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
    fmt.Println(data)
    sma := indicator.Sma(30, data)
    fmt.Println(sma)

    if (sma[len(sma)-1] < data[dataLen-1]) {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    _, rsi := indicator.Rsi([]float64{data[dataLen-3], data[dataLen-2], data[dataLen-1]})
    fmt.Println(rsi)

    // if (rsi[len(sma)-1] < data[dataLen-1]) {
    //     tests = append(tests, true)
    // } else {
    //     tests = append(tests, false)
    // }

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