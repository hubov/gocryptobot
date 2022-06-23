package trading

import (
	// "encoding/csv"
	"github.com/hubov/gocryptobot/internal/strategy"
	"time"
	"fmt"
	"io"
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

var  (
	SimTradingHistory []SimTrading
	SimWallet Wallet
	symbol string
)

func lastFileLine(fileHandle *os.File) string {
	line := ""
    var cursor int64 = -1

    stat, _ := fileHandle.Stat()
    filesize := stat.Size()
    if filesize > 0 {
		for { 
		    cursor -= 1
		    fileHandle.Seek(cursor, io.SeekEnd)

		    char := make([]byte, 1)
		    fileHandle.Read(char)

		    if cursor != -1 && (char[0] == 10 || char[0] == 13) {
		        break
		    }
		    line = fmt.Sprintf("%s%s", string(char), line)
		    if cursor <= -filesize {
		        break
		    }
		}
	}

    return line
}

func SetSymbol(tradingSymbol string) {
	symbol = tradingSymbol
}

func Simulation(startTime, endTime time.Time, base, quote, interval string, tradeLog bool) {
	SimWallet.BaseQuantity = 0
	SimWallet.QuoteQuantity = 1000
	symbol = base + quote

	strategy.SetConfig(base, quote, interval)

	if !startTime.IsZero() && !endTime.IsZero() {
		startTimeUnix := startTime.UnixMilli()
		endTimeUnix := endTime.UnixMilli()

		strategy.SetTimeframe(startTimeUnix, endTimeUnix)
		strategy.GetData(false)
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

		var i int = 501
		for i < intervalsCount {
			strategy.Update[strategy.Client.Interval] = candles[strategy.Client.Interval][0:i]
			for key, value := range intervalIterators {
				fmt.Println(key, value, i)
				for candles[key][value].OpenTime < candles[strategy.Client.Interval][i].OpenTime {
					value++
					intervalIterators[key] = value
				}
				strategy.Update[key] = candles[key][0:intervalIterators[key]]
			}

			strategy.SetData(strategy.Update)
			signals := strategy.GetSignal(false)
			for k, signal := range signals {
				if (signal != "WAIT") {
					fmt.Println(time.UnixMilli(candles[strategy.Client.Interval][i].OpenTime).UTC(), strategy.Response[k])
					SimOrder(signal, candles[strategy.Client.Interval][i].Open, candles[strategy.Client.Interval][i].OpenTime, tradeLog)
				}
			}
			i++
		}

		fmt.Println(SimWallet)
	} else {
		strategy.GetSignal(false)
	}
}

func SimOrder(signal string, price float64, tradeTime int64, tradeLog bool) {
	command := strings.Split(signal, " ")
	var quantity float64

    if command[1] == "SHORT" {
        if command[0] == "Exit" {
        	quantity = math.Abs(SimWallet.BaseQuantity)

        	SimWallet.BaseQuantity = SimWallet.BaseQuantity + quantity
        	price = strategy.ExitPrice
        	SimWallet.QuoteQuantity = SimWallet.QuoteQuantity - quantity * price

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
            price = strategy.ExitPrice
            SimWallet.QuoteQuantity = SimWallet.QuoteQuantity + quantity * price

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

	if tradeLog == true {
		TradeLog(command[0], command[1], quantity, price)
	}
}

func Trade() {
	var tradeTime bool

	strategy.GetData(true)
	signals := strategy.GetSignal(true)
	for _, signal := range signals {
		if (time.Now().Minute() == 14 || time.Now().Minute() == 29 || time.Now().Minute() == 44 || time.Now().Minute() == 59) {
			signal = "* " + signal + " *"
			tradeTime = true
		} else {
			tradeTime = false
		}
		
		fmt.Println(time.Now().UTC(), strategy.Response[len(strategy.Response) - 1])

		file, _ := openLogFile("./log/live-trading.log")
		infoLog := log.New(file, "", log.LstdFlags|log.Lmicroseconds)
		infoLog.Println(strategy.Response[len(strategy.Response) - 1])

		command := strings.Split(signal, " ")

		if command[0] == "Exit" {
			strategy.Trade(signal)
		} else if (tradeTime == true) {
			strategy.Trade(signal)
		}
	}
}

func TradeLog(order, orderType string, amount, price float64) {
	tradesFile, err := os.OpenFile("scans/trades/" + symbol + ".csv", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	line := lastFileLine(tradesFile)
	fmt.Println("tradelog:", line)
}

func TriggerTrade(signal string) {
	strategy.GetData(true)
	strategy.GetSignal(true)

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