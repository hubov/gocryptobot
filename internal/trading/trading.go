package trading

import (
	"github.com/hubov/gocryptobot/internal/strategy"
	"time"
	// "fmt"
)

func Simulation(startTime, endTime time.Time) {
	if !startTime.IsZero() && !endTime.IsZero() {
		startTimeUnix := startTime.UnixMilli()
		endTimeUnix := endTime.UnixMilli()

		strategy.SetTimeframe(startTimeUnix, endTimeUnix)
		// strategy.SetCandleStart(startTimeUnix)
		// strategy.SetCandleEnd(endTimeUnix)
	}
	strategy.GetSignal()
	// fmt.Println(startTimeUnix, endTimeUnix)
}