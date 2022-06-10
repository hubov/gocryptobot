package trading

import (
	"github.com/hubov/gocryptobot/internal/strategy"
	"time"
	"fmt"
)

func Simulation(startTime, endTime time.Time) {
	if !startTime.IsZero() && !endTime.IsZero() {
		startTimeUnix := startTime.UnixMilli()
		endTimeUnix := endTime.UnixMilli()

		fmt.Println(startTimeUnix, endTimeUnix)

		strategy.SetTimeframe(startTimeUnix, endTimeUnix)
		strategy.GetData()
		candles := strategy.Candles
		intervalsCount := int(strategy.IntervalsCount)

		fmt.Println(intervalsCount)
		fmt.Println(strategy.Client.Interval)

		var intervalIterators = make(map[string]int)
		if len(candles) > 1 {
			for key, _ := range candles {
				if key != strategy.Client.Interval {
					intervalIterators[key] = 1
				}
			}
		}

		fmt.Println(intervalIterators)

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
			// fmt.Println(i, intervalIterators)

			strategy.SetData(strategy.Update)
			if (strategy.GetSignal() != "WAIT") {
				fmt.Println(time.UnixMilli(candles[strategy.Client.Interval][i].OpenTime).UTC(), strategy.GetSignal(), candles[strategy.Client.Interval][i].Open, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])
			} else {
				// fmt.Println(time.UnixMilli(candles[strategy.Client.Interval][i].OpenTime).UTC(), strategy.GetSignal(), candles[strategy.Client.Interval][i].Open, "|", strategy.Rsi[strategy.RsiLen-2], strategy.Rsi[strategy.RsiLen-1], strategy.R1, strategy.Sma[len(strategy.Sma)-1], strategy.Data[strategy.DataLen-1])
			}
			i++
		}
	} else {
		strategy.GetSignal()
	}
}