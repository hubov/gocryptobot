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

func main() {
	base := flag.String("base", "", "filter by base currency")
	quote := flag.String("quote", "", "filter by quote currency")
	update := flag.Bool("update", false, "update candles?")
	interval := "15m"
	var intervalMilli int64 = 900000
	var apiCallsPerMinute = 1100 // API calls limit per minute

	flag.Parse()

scriptStart := time.Now()

	baseFilter := *base
	quoteFilter := *quote
	updateCandles := *update

	timeout := time.Second * 10
	client := binance.ApiClient(timeout)
	availablePairs := client.GetAllMarginPairs() // Weight: 1 (IP)

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
	    		var line string
				candlesFile, err := os.OpenFile("scans/candles/" + pair.Symbol + ".csv", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	    		if err != nil {
	    			log.Fatal(err)
	    		}

	    		if (updateCandles == true) {
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
	    				err = client.GetCandlesParams(pair.Symbol, interval) // Weight: 1 (IP)
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

					candlesFile, err = os.OpenFile("scans/candles/" + pair.Symbol + ".csv", os.O_RDONLY, 0755)
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

		    		fmt.Println(time.UnixMilli(times[0]).UTC(), time.UnixMilli(times[1]).UTC())

		    		trading.Simulation(time.UnixMilli(times[0]), time.UnixMilli(times[1]), pair.Base, pair.Quote, interval, true, records)
	    		}

	    		tradesFile, err = os.OpenFile("scans/trades/" + pair.Symbol + ".csv", os.O_RDONLY, 0755)
				if err != nil {
	    			log.Fatal(err)
	    		}
	    		tradesReader := csv.NewReader(tradesFile)

	    		for {
	    			trade, err := tradesReader.Read()

	    			
	    		}

os.Exit(11)
	   //  		candlesFile.Seek(0, io.SeekStart)
	   //  		candlesReader := csv.NewReader(candlesFile)
	   //  		for {
	   //  			record, err := candlesReader.Read()
	   //  			if err == io.EOF {
	   //  				break
	   //  			}
				// 	if err != nil {
				// 		log.Fatal(err)
				// 	}
				
				// 	for value := range record {
				// 		fmt.Printf("%s\n", record[value])
				// 	}
				// }
	        	i++
	   //      	os.Exit(1)
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

func str2float(input string) (output float64) {
	output, _ = strconv.ParseFloat(input, 64)

	return
}

func str2int(input string) (output int64) {
	output, _ = strconv.ParseInt(input, 10, 64)

	return
}