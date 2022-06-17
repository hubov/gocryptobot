package trading

import (
	"github.com/hubov/gocryptobot/internal/strategy"
	"time"
	"fmt"
	"strings"
	"math"
	"os"
	"log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type (
	Wallet struct {
		BaseQuantity float64
		QuoteQuantity float64
	}
	SimTrading struct {
		Operation string
		Price float64
		Quantity float64
		TradeTime time.Time
	}
)

var SimTradingHistory []SimTrading
var SimWallet Wallet

func Simulation(startTime, endTime time.Time) {
	SimWallet.BaseQuantity = 0
	SimWallet.QuoteQuantity = 1000

	if !startTime.IsZero() && !endTime.IsZero() {
		startTimeUnix := startTime.UnixMilli()
		endTimeUnix := endTime.UnixMilli()

		strategy.SetTimeframe(startTimeUnix, endTimeUnix)
		strategy.GetData()
		candles := strategy.Candles
		intervalsCount := int(strategy.IntervalsCount)

		var intervalIterators = make(map[string]int)
		if len(candles) > 1 {
			for key, _ := range candles {
				if key != strategy.Client.Interval {
					intervalIterators[key] = 1
				}
			}
		}

		var i int
		i = 501
		for i < intervalsCount {
			strategy.Update[strategy.Client.Interval] = candles[strategy.Client.Interval][0:i]
			for key, value := range intervalIterators {
				for candles[key][value].CloseTime < candles[strategy.Client.Interval][i].CloseTime {
					value++
					intervalIterators[key] = value
				}
				strategy.Update[key] = candles[key][0:intervalIterators[key]]
			}

			strategy.SetData(strategy.Update)
			signals := strategy.GetSignal()
			for _, signal := range signals {
				if (signal != "WAIT") {
					fmt.Println(time.UnixMilli(candles[strategy.Client.Interval][i].OpenTime).UTC(), signal, candles[strategy.Client.Interval][i].Open, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])
					SimOrder(signal, candles[strategy.Client.Interval][i].Open, candles[strategy.Client.Interval][i].OpenTime)
				}
			}
			i++
		}

		fmt.Println(SimWallet)
	} else {
		strategy.GetSignal()
	}
}

func SimOrder(signal string, price float64, tradeTime int64) {
	command := strings.Split(signal, " ")
	var quantity float64

    if command[1] == "SHORT" {
        if command[0] == "Exit" {
        	quantity = math.Abs(SimWallet.BaseQuantity)

        	SimWallet.BaseQuantity = SimWallet.BaseQuantity + quantity
        	SimWallet.QuoteQuantity = SimWallet.QuoteQuantity - quantity * strategy.ExitPrice

        	change := math.Round(((strategy.LastBuyPrice - price) / strategy.LastBuyPrice * 100) * 100) / 100
        	fmt.Println(change,  "%")

        	strategy.LastBuyPrice = 0
        } else if command[0] == "Close" {
        	quantity = math.Abs(SimWallet.BaseQuantity)

        	SimWallet.BaseQuantity = SimWallet.BaseQuantity + quantity
        	SimWallet.QuoteQuantity = SimWallet.QuoteQuantity - quantity * price

        	change := math.Round(((strategy.LastBuyPrice - price) / strategy.LastBuyPrice * 100) * 100) / 100
        	fmt.Println(change,  "%")

        	strategy.LastBuyPrice = 0
        } else if (command[0] == "Order") {
        	quantity = SimWallet.QuoteQuantity / price

            SimWallet.BaseQuantity = SimWallet.BaseQuantity - quantity
            SimWallet.QuoteQuantity = SimWallet.QuoteQuantity * 2

            strategy.LastBuyPrice = price
        }
        strategy.SymbolWorth = SimWallet.BaseQuantity / price
    } else if command[1] == "LONG" {
        if command[0] == "Exit" {
            quantity = SimWallet.BaseQuantity

            SimWallet.BaseQuantity = 0
            SimWallet.QuoteQuantity = SimWallet.QuoteQuantity + quantity * strategy.ExitPrice

            change := math.Round(((price - strategy.LastBuyPrice) / strategy.LastBuyPrice * 100) * 100) / 100
        	fmt.Println(change,  "%")

            strategy.LastBuyPrice = 0
        } else if command[0] == "Close" {
        	quantity = SimWallet.BaseQuantity

            SimWallet.BaseQuantity = 0
            SimWallet.QuoteQuantity = SimWallet.QuoteQuantity + quantity * price

            change := math.Round(((price - strategy.LastBuyPrice) / strategy.LastBuyPrice * 100) * 100) / 100
        	fmt.Println(change,  "%")

            strategy.LastBuyPrice = 0
        } else if (command[0] == "Order") {
            quantity = SimWallet.QuoteQuantity / price

            SimWallet.BaseQuantity = SimWallet.BaseQuantity + quantity
            SimWallet.QuoteQuantity = 0

            strategy.LastBuyPrice = price
        }

        strategy.SymbolWorth = SimWallet.BaseQuantity / price
    }

    row := SimTrading{
		Operation: signal,
		Price: price,
		Quantity: quantity,
		TradeTime: time.UnixMilli(tradeTime).UTC(),
	}
	SimTradingHistory = append(SimTradingHistory, row)
}

func Trade() {
	var tradeTime bool

	strategy.GetData()
	signals := strategy.GetSignal()
	for _, signal := range signals {
		if (time.Now().Minute() == 14 || time.Now().Minute() == 29 || time.Now().Minute() == 44 || time.Now().Minute() == 59) {
			signal = "* " + signal + " *"
			tradeTime = true
		} else {
			tradeTime = false
		}
		
		fmt.Println(time.Now().UTC(), "[", signal, "]", strategy.Candles[strategy.Client.Interval][len(strategy.Candles[strategy.Client.Interval]) - 1].Close, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])

		file, _ := openLogFile("./log/live-trading.log")
		infoLog := log.New(file, "[" + signal + "] ", log.LstdFlags|log.Lmicroseconds)
		infoLog.Println(strategy.Candles[strategy.Client.Interval][len(strategy.Candles[strategy.Client.Interval]) - 1].Close, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])

		if (tradeTime == true) {
			// strategy.Trade(signal)
		}
	}
}

func TriggerTrade(signal string) {
	strategy.GetData()
	strategy.GetSignal()

	command := strings.Split(signal, " ")

	titler :=  cases.Title(language.English)
	upper :=  cases.Upper(language.English)
	command[0] = titler.String(command[0])
	command[1] = upper.String(command[1])

	fmt.Println("TRADE!!!")
	strategy.Trade(command[0] + " " + command[1])
}

func openLogFile(path string) (logFile *os.File, err error) {
	logFile, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        return nil, err
    }
    return
}