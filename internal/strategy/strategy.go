package strategy

import (
	"github.com/hubov/gocryptobot/internal/binance"
    "github.com/cinar/indicator"
	"fmt"
	"time"
	"log"
)

var candles []binance.Candle
var defaultTimeout time.Duration
// var client *binance.Client
var timeStart int64
var timeEnd int64
var data []float64
var dataLen int
var sma []float64
var rsi []float64
var rsiLen int
var PivotPoint float64
var S1 float64
var R1 float64
var S2 float64
var R2 float64
var S3 float64
var R3 float64

// func SetCandleStart (start int64) {
//     client.SetStartTime(start)

//     return
// }

// func SetCandleEnd (end int64) {
//     client.SetEndTime(end)

//     return
// }

func SetTimeframe(start, end int64) {
    timeStart = start
    timeEnd = end
}

func Calculate() {
    defaultTimeout := time.Second * 10
    client := binance.ApiClient(defaultTimeout)
    if timeStart != 0 {
        client.SetTimeframe(timeStart, timeEnd)
    }
    wallet, err := client.SpotBalance()
    if err != nil {
        log.Fatal(err)
    }
    for _, coin := range wallet {
        if coin.Free > 0 || coin.Locked > 0 {
            fmt.Println(coin)
        }
    }
    candles, err = client.GetCandles()
    if err != nil {
        log.Fatal(err)
    }

//  ###################################
// USTALIĆ MINIMALNY PERIOD DLA STRATEGII !!!!!!!
//  ###################################
    data = GetValues(30, "close")
    dataLen = len(data)
    sma = indicator.Sma(30, data)

    data = GetValues(500, "close")
    dataLen = len(data)
    _, rsi = indicator.RsiPeriod(2, data)
    rsiLen = len(rsi)

    client1D := binance.ApiClient(defaultTimeout)
    candles1D, err := client1D.GetCandlesParams(client1D.Symbol, "1d")
    if err != nil {
        log.Fatal(err)
    }
    data1D := candles1D[len(candles1D) - 2]
    PivotPoint = (data1D.High + data1D.Low + data1D.Close) / 3
    S1 = 2*PivotPoint - data1D.High
    // R1 = 2*PivotPoint - data1D.Low
    // S2 = PivotPoint - (R1 - S1)
    // R2 = PivotPoint + (R1 - S1)
    // S3 = data1D.Low - 2 * (data1D.High - PivotPoint)
    // R3 = data1D.High + 2 * (PivotPoint - data1D.Low)

    fmt.Println(data[dataLen-1], S1)
}

func GetValues(period int, periodType string) (result []float64) {
    result = GetValuesParams(period, periodType, candles)

    return
}

func GetValuesParams(period int, periodType string, paramCandles []binance.Candle) (result []float64) {
    var price float64
    funcCandles := paramCandles
    // fmt.Println(funcCandles)
    // fmt.Println(period)
    len := len(funcCandles)
    // fmt.Println(len)
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

func SignalOrderLong() (result bool) {
    var tests []bool
    result = true

    if sma[len(sma)-1] < data[dataLen-1] {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if rsi[rsiLen-2] < 5 && rsi[rsiLen-1] >= 5 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if (data[dataLen-1] < S1) {
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

func SingalCloseLong() (result bool) {
    var tests []bool
    if rsi[rsiLen-2] > 95 && rsi[rsiLen-1] <= 95 {
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

func SingalExitLong() bool {
    // based on purchase price
    return false
}

func SignalOrderShort() (result bool) {
    var tests []bool
    result = true

    if sma[len(sma)-1] > data[dataLen-1] {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if rsi[rsiLen-2] > 95 && rsi[rsiLen-1] <= 95 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if (data[dataLen-1] > R1) {
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

func SingalCloseShort() (result bool) {
    var tests []bool
    if rsi[rsiLen-2] < 5 && rsi[rsiLen-1] >= 5 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    return
}

func SingalExitShort() bool {
    // based on purchase price
    return false
}

func GetSignal() string {
    Calculate()

    if SignalOrderLong() {
        return "Order LONG"
    } else if SingalCloseLong() {
        return "Close LONG"
    } else if SingalExitLong() {
        return "Exit LONG"
    } else if SignalOrderShort() {
        return "Order SHORT"
    } else if SingalCloseShort() {
        return "Close SHORT"
    } else if SingalExitLong() {
        return "Exit SHORT"
    } else {
        return "WAIT" + fmt.Sprintf("%f", data[dataLen-1])
    }

}