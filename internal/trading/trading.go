package trading

import (
	"github.com/hubov/gocryptobot/internal/strategy"
	"time"
	"fmt"
	"strings"
)

func Simulation(startTime, endTime time.Time) {
	if !startTime.IsZero() && !endTime.IsZero() {
		startTimeUnix := startTime.UnixMilli()
		endTimeUnix := endTime.UnixMilli()

		// fmt.Println(startTimeUnix, endTimeUnix)

		strategy.SetTimeframe(startTimeUnix, endTimeUnix)
		strategy.GetData()
		candles := strategy.Candles
		intervalsCount := int(strategy.IntervalsCount)

		// fmt.Println(intervalsCount)
		// fmt.Println(strategy.Client.Interval)

		var intervalIterators = make(map[string]int)
		if len(candles) > 1 {
			for key, _ := range candles {
				if key != strategy.Client.Interval {
					intervalIterators[key] = 1
				}
			}
		}

		// fmt.Println(intervalIterators)

		// fmt.Println(candles)

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
					SimOrder(signal)
				} else {
					// fmt.Println(time.UnixMilli(candles[strategy.Client.Interval][i].OpenTime).UTC(), strategy.GetSignal(), candles[strategy.Client.Interval][i].Open, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])
				}
			}
			i++
		}
	} else {
		strategy.GetSignal()
	}
}

func SimOrder(signal string) {
	command := strings.Split(signal)

    if command[1] == "SHORT" {
        if command[0] == "Close" || command[0] == "Exit" {
            
        } else if (command[0] == "Order") {
            c.OrderMargin("SELL", "MARGIN_BUY")
        }
    } else if command[1] == "LONG" {
        if command[0] == "Close" || command[0] == "Exit" {
            c.OrderMargin("SELL", "NO_SIDE_EFFECT")
        } else if (command[0] == "Order") {
            c.OrderMargin("BUY", "NO_SIDE_EFFECT")
        }
    }
}

func Trade() {
	signals := strategy.GetSignal()
	for _, signal := range signals {
		if (signal != "WAIT") {
			fmt.Println(time.UnixMilli(candles[strategy.Client.Interval][i].OpenTime).UTC(), signal, candles[strategy.Client.Interval][i].Open, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])
			strategy.Trade(signal)
		} else {
			fmt.Println(time.UnixMilli(candles[strategy.Client.Interval][i].OpenTime).UTC(), signal, candles[strategy.Client.Interval][i].Open, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])
		}
	}
}