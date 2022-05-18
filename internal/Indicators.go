package indicators

import (
	"github.com/hubov/gocryptobot/internal/binance"
	"fmt"
)

var candles *[]binance.Candle

// type (
// 	Candle struct {
//         OpenTime int64
//         Open float64
//         High float64
//         Low float64
//         Close float64
//         Volume float64
//         CloseTime int64
//         QuoteAssetVolume float64
//         TradesNumber int64
//         TakerBuyBaseAssetVolume float64
//         TakerBuyQuoteAssetVolume float64
//         Ignore float64
//     }
// )

// func (c *binance.Candle) Set(openTime int64, open, high, low, close, volume float64, closeTime int64, quoteAssetVolume float64, tradesNumber int64, takerBuyBaseAssetVolume, takerBuyQuoteAssetVolume, ignore float64) {
//     c.OpenTime = openTime
//     c.Open = open
//     c.High = high
//     c.Low = low
//     c.Close = close
//     c.Volume = volume
//     c.CloseTime = closeTime
//     c.QuoteAssetVolume = quoteAssetVolume
//     c.TradesNumber = tradesNumber
//     c.TakerBuyBaseAssetVolume = takerBuyBaseAssetVolume
//     c.TakerBuyQuoteAssetVolume = takerBuyQuoteAssetVolume
//     c.Ignore = ignore
// }

func SetCandles(input *[]binance.Candle) {
	candles = input
}

func EMA(length int, candlePrice string) (result []float64) {
	fmt.Println(candles[len(candles)-1])
	// result = candles[len(candles)-1]

	return
}