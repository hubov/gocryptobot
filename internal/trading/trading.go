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

		
	} else {
		strategy.GetSignal()
	}
}