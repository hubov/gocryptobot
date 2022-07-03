package strategy

import (
	"github.com/hubov/gocryptobot/internal/binance"
    "github.com/cinar/indicator"
	"time"
	"log"
    "math"
    "fmt"
    "strconv"
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
var LastBuyPrice float64
var SymbolWorth float64
var PivotSignal []int64
var DataExitLow []float64
var DataExitLowLen int
var DataExitHigh []float64
var DataExitHighLen int
var ExitPrice float64
var Response []string

func GetBaseQuantity() float64 {
    return binance.GetBaseQuantity()
}

func GetQuoteQuantity() float64 {
    return binance.GetQuoteQuantity()
}

func SetTimeframe(start, end int64) {
    timeStart = start
    timeEnd = end
}

func GetData(isLive bool) {
    defaultTimeout := time.Second * 10
    Client = binance.ApiClient(defaultTimeout)
    if timeStart != 0 {
        Client.SetTimeframe(timeStart, timeEnd)
    }
    Client.GetWallet(isLive)
    LastBuyPrice = binance.LastBuyPrice
    SymbolWorth = binance.SymbolWorth
    err := Client.GetCandles()
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
    Response = nil
}

func Calculate() {
    Data = GetValues(Client.Interval, 30, "close")
    DataLen = len(Data)
    Sma = indicator.Sma(30, Data)

    Data = GetValues(Client.Interval, 500, "close")
    DataLen = len(Data)
    _, Rsi = indicator.RsiPeriod(2, Data)
    RsiLen = len(Rsi)

    DataExitLow = GetValues(Client.Interval, 1, "low")
    DataExitLowLen = len(DataExitLow)
    DataExitHigh = GetValues(Client.Interval, 1, "high")
    DataExitHighLen = len(DataExitHigh)

    data1DLen := len(Candles["1d"])
    dataLen := len(Candles[Client.Interval])
    i := data1DLen - 6
    iDayBefore := 0;
    j := dataLen - 500
    PivotSignal = nil
    for j < dataLen && i < data1DLen {
        for i < (data1DLen - 1) && Candles[Client.Interval][j].CloseTime > Candles["1d"][i].CloseTime {
            i++
        }

        iDayBefore = i - 1
        
        PivotPoint = (Candles["1d"][iDayBefore].High + Candles["1d"][iDayBefore].Low + Candles["1d"][iDayBefore].Close) / 3
        S1 = 2*PivotPoint - Candles["1d"][iDayBefore].High
        R1 = 2*PivotPoint - Candles["1d"][iDayBefore].Low

        if Candles[Client.Interval][j].Close > R1 {
            PivotSignal = append(PivotSignal, 1/*, Candles[Client.Interval][j].OpenTime*/)
        } else if Candles[Client.Interval][j].Close < S1 {
            PivotSignal = append(PivotSignal, -1/*, Candles[Client.Interval][j].OpenTime*/)
        } else {
            if len(PivotSignal) > 0 {
                PivotSignal = append(PivotSignal, PivotSignal[len(PivotSignal) - 1]/*, Candles[Client.Interval][j].OpenTime*/)
            } else {
                PivotSignal = append(PivotSignal, 0/*, Candles[Client.Interval][j].OpenTime*/)
            }
        }

        j++
    }
    // S2 = PivotPoint - (R1 - S1)
    // R2 = PivotPoint + (R1 - S1)
    // S3 = data1D.Low - 2 * (data1D.High - PivotPoint)
    // R3 = data1D.High + 2 * (PivotPoint - data1D.Low)
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

func LongLimit() (result float64) {
    result = LastBuyPrice * 1.1

    return
}

func LongStop() (result float64) {
    result = LastBuyPrice * 0.95

    return
}

func ShortLimit() (result float64) {
    result = LastBuyPrice * 0.9

    return
}

func ShortStop() (result float64) {
    result = LastBuyPrice * 1.05

    return
}

func SignalOrderLong() (result bool) {
    var tests []bool
    result = true

    if Sma[len(Sma)-1] < Data[DataLen-1] {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if Rsi[RsiLen-2] < 5 && Rsi[RsiLen-1] >= 5 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if PivotSignal[len(PivotSignal) - 1] == 1 {
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
    result = true

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

func SingalExitLong() (result bool) {
    var tests []bool
    result = false

    if DataExitLow[DataExitLowLen-1] <= LongStop() {
        tests = append(tests, true)
        ExitPrice = LongStop()
    } else {
        tests = append(tests, false)
    }

    if DataExitHigh[DataExitHighLen-1] >= LongLimit() {
        tests = append(tests, true)
        ExitPrice = LongLimit()
    } else {
        tests = append(tests, false)
    }

    for _, test := range tests {
        if test == true {
            result = true
        }
    }

    return result
}

func SignalOrderShort() (result bool) {
    var tests []bool
    result = true

    if Data[DataLen-1] < Sma[len(Sma)-1] {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if Rsi[RsiLen-2] >= 95 && Rsi[RsiLen-1] < 95 {
        tests = append(tests, true)
    } else {
        tests = append(tests, false)
    }

    if PivotSignal[len(PivotSignal) - 1] == -1 {
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
    result = true

    if Rsi[RsiLen-2] < 5 && Rsi[RsiLen-1] >= 5 {
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

func SingalExitShort() (result bool) {
    var tests []bool
    result = false

    if DataExitLow[DataExitLowLen-1] <= ShortLimit() {
        tests = append(tests, true)
        ExitPrice = ShortLimit()
    } else {
        tests = append(tests, false)
    }

    if DataExitHigh[DataExitHighLen-1] >= ShortStop() {
        tests = append(tests, true)
        ExitPrice = ShortStop()
    } else {
        tests = append(tests, false)
    }

    for _, test := range tests {
        if test == true {
            result = true
        }
    }

    return result
}

func GetSignal(isLive bool) (signals []string) {
    Calculate()

    // if the Cryptocurrency value in wallet is significant try to close/exit position
    if math.Abs(SymbolWorth) >= 2 || isLive == false {
        if SymbolWorth > 0 {
            if SingalCloseLong() {
                signals = append(signals, "Close LONG")
            } else if SingalExitLong() {
                signals = append(signals, "Exit LONG")
            }
        } else if SymbolWorth < 0 {
            if SingalCloseShort() {
                signals = append(signals, "Close SHORT")
            } else if SingalExitShort() {
                signals = append(signals, "Exit SHORT")
            }
        }
    }

    if SignalOrderLong() {
        signals = append(signals, "Order LONG")
    } else if SignalOrderShort() {
        signals = append(signals, "Order SHORT")
    }

    if (len(signals) == 0) {
        signals = append(signals, "WAIT")
    }

    for _, signal := range signals {
        response := " [ " + signal + " ] " + float2str(Candles[Client.Interval][len(Candles[Client.Interval]) - 1].Close) + " | " + float2str(Rsi[RsiLen-2]) + " " + float2str(Rsi[RsiLen-1]) + " " + float2str(R1) + " " + float2str(Sma[len(Sma)-1]) + " " + int2str(PivotSignal[len(PivotSignal)-1])
        Response = append(Response, response)
    }

    return
}

func float2str(input float64) (output string) {
    output = strconv.FormatFloat(input, 'f', -1, 64)

    return
}

func int2str(input int64) (output string) {
    output = strconv.FormatInt(input, 10)

    return
}

func Trade(signal string) {
    fmt.Println(GetBaseQuantity(), GetQuoteQuantity())
    Client.Trade(math.Abs(GetBaseQuantity()), GetQuoteQuantity(), signal)
    // pass additional vars:
    // AMOUNT IF MARGIN SELL/BUY
}