package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/hubov/gocryptobot/internal/binance"
	"github.com/hubov/gocryptobot/internal/trading"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	intervalMilli int64 = 900000
	apiCallsPerMinute = 1100 // API calls limit per minute
	dayMilli int64 = 86400000
	candles = make(map[int64][]string)
)


func main() {
	base := flag.String("base", "", "filter by base currency")
	quote := flag.String("quote", "", "filter by quote currency")
	update := flag.Bool("update", false, "update candles?")
	interval := "15m"

	flag.Parse()

scriptStart := time.Now()

	baseFilter := *base
	quoteFilter := *quote
	updateCandles := *update

	timeout := time.Second * 10
	client := binance.ApiClient(timeout)
	availablePairs := client.GetAllMarginPairs() // Weight: 1 (IP)


// api call
	i := 0
	apiCallsLimit := apiCallsPerMinute
	apiCallsUpdate := time.Now()
    for _, pair := range availablePairs {
    	nowApi := time.Now()
    	if nowApi.Add(time.Duration(-1) * time.Minute).After(apiCallsUpdate) {
    		apiCallsUpdate = time.Now()
    		apiCallsLimit = apiCallsPerMinute
    	}

    	if apiCallsLimit > 0 {
	    	if baseFilter == "" && quoteFilter == "" || baseFilter != "" && pair.Base == baseFilter && quoteFilter != "" && pair.Quote == quoteFilter || (baseFilter != "" && quoteFilter == "" && pair.Base == baseFilter || quoteFilter != "" && baseFilter == "" && pair.Quote == quoteFilter) {
	    		fmt.Println(pair)

				if (updateCandles == true) {
					apiCallsLimit = saveCandles(pair.Symbol, interval, client, intervalMilli, apiCallsLimit)
				}

				records := readCandles(pair.Symbol)

	    		nowFileName := time.Now().UTC().Format("20060102150405")
	    		trading.Simulation(time.UnixMilli(str2int(records[0]["OpenTime"])), time.UnixMilli(str2int(records[len(records) - 1]["CloseTime"])), pair.Base, pair.Quote, interval, true, records, nowFileName)

	    		trades := readTrades(pair.Symbol, nowFileName)

	    		var budget float64 = 1000
	    		fmt.Println("BUDGET", budget)
	    		for _, trade := range trades {
	    			budget *= 1+(trade-0.0075)
	    			fmt.Println(budget, (trade*100))
	    		}

os.Exit(11)
	        	i++
	    	}
	    } else {
	    	time.Sleep(time.Until(apiCallsUpdate.Add(61 * time.Second)))
	    }
    }
    fmt.Println("Count:", i)

scriptEnd := time.Now()
fmt.Println("Script time:", scriptEnd.Sub(scriptStart))
os.Exit(3)
}

func fileExists(filepath string) bool {
    info, err := os.Stat(filepath)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func float2str(input float64) (output string) {
    output = strconv.FormatFloat(input, 'f', -1, 64)

    return
}

func getFileLine(fileHandle *os.File, reversed bool) string {
	line := ""
    var cursor int64

    if reversed == true {
    	cursor = -1
    } else {
    	cursor = -1
    }

    stat, _ := fileHandle.Stat()
    filesize := stat.Size()
    if filesize > 0 {
		for { 
			if reversed == true {
		    	cursor -= 1
		    	fileHandle.Seek(cursor, io.SeekEnd)
			} else {
				cursor += 1
				fileHandle.Seek(cursor, io.SeekStart)
			}

		    char := make([]byte, 1)
		    fileHandle.Read(char)

		    if cursor != -1 && (char[0] == 10 || char[0] == 13) {
		        break
		    }
		    if reversed == true {
		    	line = fmt.Sprintf("%s%s", string(char), line)
		    } else {
		    	line = fmt.Sprintf("%s%s", line, string(char))
		    }
		    if cursor <= -filesize || cursor >= filesize {
		        break
		    }
		}
	}

    return line
}

func int2str(input int64) (output string) {
    output = strconv.FormatInt(input, 10)

    return
}

func readCandles(symbol string) []map[string]string {
	candlesFile, err := os.OpenFile("scans/candles/" + symbol + ".csv", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	candlesReader := csv.NewReader(candlesFile)
	var (
		record []string
		records []map[string]string
	)
	for {
		record, err = candlesReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		row := make(map[string]string)

		row["OpenTime"] = record[0]
		row["Open"] = record[2]
		row["High"] = record[3]
		row["Low"] = record[4]
		row["Close"] = record[5]
		row["Volume"] = record[6]
		row["CloseTime"] = record[1]
		row["QuoteAssetVolume"] = record[7]
		row["TradesNumber"] = record[8]
		row["TakerBuyBaseAssetVolume"] = record[9]
		row["TakerBuyQuoteAssetVolume"] = record[10]
		row["Ignore"] = record[11]

		records = append(records, row)
	}
	candlesFile.Close()
	fmt.Println(len(records))

	fmt.Println(time.UnixMilli(str2int(records[0]["OpenTime"])).UTC(), time.UnixMilli(str2int(records[len(records) - 1]["CloseTime"])).UTC())

	return records
}

func readTrades(symbol, nowFileName string) (tradeResults []float64) {
	tradesFile, err := os.OpenFile("scans/trades/" + symbol + "-" + nowFileName + ".csv", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	tradesReader := csv.NewReader(tradesFile)

	var (
		wins, losses []int64
		lastPrice, currentPrice float64
		lastDate int64
		change float64
	)

	for {
		trade, err := tradesReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(trade)

		currentPrice = str2float(trade[4])
		if trade[1] == "Order" {
			lastPrice = currentPrice
			lastDate = str2int(trade[0])
		} else if lastPrice != -1  && lastDate > 0 {
			if trade[2] == "SHORT" {
				if lastPrice > currentPrice {
					wins = append(wins, lastDate)
					change = (currentPrice - lastPrice) / lastPrice * -1
					change -= 0.001
				} else {
					losses = append(losses, lastDate)
					change = (currentPrice - lastPrice) / lastPrice * -1
					change -= 0.001
				}
			} else {
				if lastPrice < currentPrice {
					wins = append(wins, lastDate)
					change = (currentPrice - lastPrice) / lastPrice
				} else {
					losses = append(losses, lastDate)
					change = (currentPrice - lastPrice) / lastPrice * -1
				}
			}
			tradeResults = append(tradeResults, change)
			lastPrice = -1
		}
	}
	tradesFile.Close()

	fmt.Println("wins", wins)
	fmt.Println(len(wins))
	fmt.Println("losses", losses)
	fmt.Println(len(losses))

	return
}

func saveCandles(symbol, interval string, client *binance.Client, intervalMilli int64, apiCallsLimit int) int {
	var line string
	candlesFile, err := os.OpenFile("scans/candles/" + symbol + ".csv", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	line = getFileLine(candlesFile, true)
	now := time.Now().UTC()
	var timeStart int64
	var start time.Time
	var fileLastUpdate int64 = 0

	if (line != "") {
		lineFields := strings.Split(line, ",")
		fileLastUpdate, _ := strconv.ParseInt(lineFields[1], 10, 64)
		timeStart = fileLastUpdate + 1
		start = time.UnixMilli(timeStart)
	} else {
		start = now.Add(time.Duration(-4392) * time.Hour)
		timeStart = start.UnixMilli()
	}

	timeDifference := now.Sub(start)
	candles2Update := timeDifference.Milliseconds() / intervalMilli
	fmt.Println(time.UnixMilli(timeStart).UTC(), timeDifference, candles2Update)

	if candles2Update > 0 {
		// candlesFile.Seek(0, io.SeekStart)

		fmt.Println(time.UnixMilli(timeStart).UTC(), timeStart)

		client.SetTimeframeOffset(timeStart, now.UnixMilli(), 0)
		client.ClearData()
		err = client.GetCandlesParams(symbol, interval) // Weight: 1 (IP)
		apiCallsLimit -= int(math.Ceil(float64(candles2Update) / 1000))

		candlesWriter := csv.NewWriter(candlesFile)
		// defer candlesWriter.Flush()
		// defer candlesFile.Close()

		for _, candle := range client.Candles {
			row := []string{int2str(candle.OpenTime), int2str(candle.CloseTime), float2str(candle.Open), float2str(candle.High), float2str(candle.Low), float2str(candle.Close), float2str(candle.Volume), float2str(candle.QuoteAssetVolume), int2str(candle.TradesNumber), float2str(candle.TakerBuyBaseAssetVolume), float2str(candle.TakerBuyQuoteAssetVolume), float2str(candle.Ignore)}
			if (candle.OpenTime > fileLastUpdate) {
				if err := candlesWriter.Write(row); err != nil {
					log.Fatalln("error writing record to file", err)
				}
			}
		}
		candlesWriter.Flush()
		fmt.Println(candlesWriter.Error())
	}

	var (
		lines []string = nil
		times []int64 = nil
	)
	lines = append(lines, getFileLine(candlesFile, false))
	lines = append(lines, getFileLine(candlesFile, true))
	candlesFile.Close()

	fmt.Println(lines)

	if len(lines) == 2 {
		for _, line := range lines {
			lineFields := strings.Split(line, ",")
			lineTime, _ := strconv.ParseInt(lineFields[0], 10, 64)
			times = append(times, lineTime)
		}
	} else {
		panic("Something  went wrong.")
	}

	return apiCallsLimit
}

func sliceData(candles map[int64][]string, timeFrom, timeTo int64, periodType string) (result []float64) {
	timeFrom -= 500 * intervalMilli

	fmt.Println(timeFrom, timeTo)

	var property string
    i := timeFrom
    for i < timeTo {
        switch (periodType) {
        case "close": 
            property = candles[i][5]
        case "open":
            property = candles[i][2]
        case "low":
            property = candles[i][4]
        case "high": 
            property = candles[i][3]
        case "volume":
        	property = candles[i][6]
    	case "quoteassetvolume":
    		property = candles[i][7]
    	case "tradesnumber":
    		property = candles[i][8]
    	case "takerbuybaseassetvolume":
    		property = candles[i][9]
    	case "takerbuyquoteassetvolume":
    		property = candles[i][10]
    	case "ignore":
    		property = candles[i][11]
    	}

        result = append(result, str2float(property))
        i += intervalMilli
    }

    return
}

func str2float(input string) (output float64) {
	output, _ = strconv.ParseFloat(input, 64)

	return
}

func str2int(input string) (output int64) {
	output, _ = strconv.ParseInt(input, 10, 64)

	return
}