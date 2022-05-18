package Indicators

import (
	"github.com/hubov/gocryptobot/internal/binance"
	"fmt"
)

var candles []Candle

func setCandles(input []Candle) {
	candles =. input
}

func EMA(length int, candlePrice string) (result []float64) {
	// fmt.Println(candles[len(sl)-1])
	result = candles[len(sl)-1]

	return
}