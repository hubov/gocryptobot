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
		fmt.Println(strategy.IntervalsCount)
		fmt.Println(len(strategy.Candles))
		// strategy.SetCandleStart(startTimeUnix)
		// strategy.SetCandleEnd(endTimeUnix)
	}

	strategy.GetSignal()
	// fmt.Println(startTimeUnix, endTimeUnix)
}