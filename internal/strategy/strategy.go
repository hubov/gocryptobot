package strategy

import (
	"github.com/hubov/gocryptobot/internal/binance"
    "github.com/cinar/indicator"
	"fmt"
	"time"
	"log"
)

var Candles = make(map[string][]binance.Candle)
var Update = make(map[string][]binance.Candle)
var Client *binance.Client
var defaultTimeout time.Duration
var timeStart int64
var timeEnd int64
var IntervalsCount int64
var Data []float64
var DataLen int
var Sma []float64
var Rsi []float64
var RsiLen int
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
    Client = binance.ApiClient(defaultTimeout)
    if timeStart != 0 {
        Client.SetTimeframe(timeStart, timeEnd)
    }
    wallet, err := Client.SpotBalance()
    if err != nil {
        log.Fatal(err)
    }
    for _, coin := range wallet {
        if coin.Free > 0 || coin.Locked > 0 {
            fmt.Println(coin)
        }
    }
    err = Client.GetCandles()
    if err != nil {
        log.Fatal(err)
    }

    Candles[Client.Interval] = Client.Candles
    IntervalsCount = Client.IntervalsCount

    client1D := binance.ApiClient(defaultTimeout)
    err = client1D.GetCandlesParams(client1D.Symbol, "1d")
    if err != nil {
        log.Fatal(err)
    }
    Candles["1d"] = client1D.Candles
}

func SetData(candles map[string][]binance.Candle) {
    Candles = candles
}

func Calculate() {
    Data = GetValues(Client.Interval, 30, "close")
    DataLen = len(Data)
    Sma = indicator.Sma(30, Data)

    Data = GetValues(Client.Interval, 500, "close")
    DataLen = len(Data)
    _, Rsi = indicator.RsiPeriod(2, Data)
    RsiLen = len(Rsi)

    data1D := Candles["1d"][len(Candles["1d"]) - 1]
    PivotPoint = (data1D.High + data1D.Low + data1D.Close) / 3
    S1 = 2*PivotPoint - data1D.High
    R1 = 2*PivotPoint - data1D.Low
    // S2 = PivotPoint - (R1 - S1)
    // R2 = PivotPoint + (R1 - S1)
    // S3 = data1D.Low - 2 * (data1D.High - PivotPoint)
    // R3 = data1D.High + 2 * (PivotPoint - data1D.Low)

    // fmt.Println(data[dataLen-1], S1)
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


// ######################################
// ######################################
// ######################################

// pos := if close > R1 {
//     1 KUP
// } else {
//     if close < S1 {
//         -1 SPRZEDAJ
//     }
// }






// S = S1
// B = R1
// pos := if (close > R1)
//         1 (buy)
//         else
//        if (close < S1)
//          -1 (sell)
//         else
//             0 




// ######################################
// ######################################
// ######################################



func SignalOrderLong() (result bool) {
    var tests []bool
    result = true

    if Sma[len(Sma)-1] < Data[DataLen-1] {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    // if Rsi[RsiLen-2] < 5 && Rsi[RsiLen-1] >= 5 {
    if Rsi[RsiLen-1] < 5 && Rsi[RsiLen-2] >= 5 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    // if (Data[DataLen-1] > S1) {
    if (Data[DataLen-1] > R1) {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    // fmt.Println(tests)

    for _, test := range tests {
        if test == false {
            result = false
        }
    }

    return
}

func SingalCloseLong() (result bool) {
    var tests []bool
    if Rsi[RsiLen-2] > 95 && Rsi[RsiLen-1] <= 95 {
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

    if Data[DataLen-1] < Sma[len(Sma)-1] {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    // if Rsi[RsiLen-2] >= 95 && Rsi[RsiLen-1] < 95 {
    if Rsi[RsiLen-1] >= 95 && Rsi[RsiLen-2] < 95 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    // if (Data[DataLen-1] < R1) {
    if (Data[DataLen-1] < S1) {
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
    if Rsi[RsiLen-2] < 5 && Rsi[RsiLen-1] >= 5 {
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
        return "WAIT"/* + fmt.Sprintf("%f", data[dataLen-1])*/
    }

}