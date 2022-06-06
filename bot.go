package main

import (
    "github.com/hubov/gocryptobot/internal/trading"
    "fmt"
    "time"
    "flag"
    // "os"
)

func main() {
    isSimulation := flag.Bool("sim", false, "execute a silumation")
    startTime := flag.String("start", "", "start time for simulation")
    endTime := flag.String("end", "", "end time for simulation")

    flag.Parse()

    timeParsedStart, _ := timeParseAny(*startTime)
    timeParsedEnd, _ := timeParseAny(*endTime)

    fmt.Println("sim:", *isSimulation)
    fmt.Println("start:", timeParsedStart)
    fmt.Println("end:", timeParsedEnd)

    simulate := *isSimulation

    // os.Exit(3)

    if simulate == true {
        trading.Simulation(timeParsedStart, timeParsedEnd)
    }
    // for {
    //     fmt.Println(strategy.GetSignal())
    //     time.Sleep(60 * time.Second)
    // }
}

func timeParseAny(dateTime string) (result time.Time, err error) {
    formats := [2]string{"2006-01-02", "2006-01-02 15:04:05"}
    for _, format := range formats {
        result, err = time.Parse(format, dateTime)
        if err == nil {
            break
        }
    }

    return
}