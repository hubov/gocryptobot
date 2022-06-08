package strategy

import (
	"github.com/hubov/gocryptobot/internal/binance"
    "github.com/cinar/indicator"
	"fmt"
	"time"
	"log"
)

var Candles = make(map[string][]binance.Candle)
// var Candles = []binance.Candle
var client *binance.Client
var defaultTimeout time.Duration
var timeStart int64
var timeEnd int64
var IntervalsCount int64
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

func SetTimeframe(start, end int64) {
    timeStart = start
    timeEnd = end
}

func GetData() {
    defaultTimeout := time.Second * 10
    client = binance.ApiClient(defaultTimeout)
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
    err = client.GetCandles()
    if err != nil {
        log.Fatal(err)
    }

    Candles[client.Interval] = client.Candles
    IntervalsCount = client.IntervalsCount

    client1D := binance.ApiClient(defaultTimeout)
    err = client1D.GetCandlesParams(client1D.Symbol, "1d")
    if err != nil {
        log.Fatal(err)
    }
    Candles["1d"] = client1D.Candles
}

func Calculate() {
    data = GetValues(client.Interval, 30, "close")
    dataLen = len(data)
    sma = indicator.Sma(30, data)

    data = GetValues(client.Interval, 500, "close")
    dataLen = len(data)
    _, rsi = indicator.RsiPeriod(2, data)
    rsiLen = len(rsi)

    // data1D := client1D.Candles[len(client1D.Candles) - 2]
    data1D := Candles["1d"][len(Candles["1d"]) - 2]
    PivotPoint = (data1D.High + data1D.Low + data1D.Close) / 3
    S1 = 2*PivotPoint - data1D.High
    // R1 = 2*PivotPoint - data1D.Low
    // S2 = PivotPoint - (R1 - S1)
    // R2 = PivotPoint + (R1 - S1)
    // S3 = data1D.Low - 2 * (data1D.High - PivotPoint)
    // R3 = data1D.High + 2 * (PivotPoint - data1D.Low)

    fmt.Println(data[dataLen-1], S1)
}

func GetValues(interval string, period int, periodType string) (result []float64) {
    result = GetValuesParams(interval, period, periodType)

    return
}

func GetValuesParams(interval string, period int, periodType string) (result []float64) {
    var price float64
    len := len(Candles[interval])
    i := len - period
    for i < len {
        switch (periodType) {
        case "close": 
            price = Candles[interval][i].Close
        case "open":
            price = Candles[interval][i].Open
        case "low":
            price = Candles[interval][i].Low
        case "high": 
            price = Candles[interval][i].High
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