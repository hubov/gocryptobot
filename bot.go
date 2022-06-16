package main

import (
    "github.com/hubov/gocryptobot/internal/trading"
    "fmt"
    "time"
    "flag"
)

func main() {
    isSimulation := flag.Bool("sim", false, "execute a silumation")
    startTime := flag.String("start", "", "start time for simulation")
    endTime := flag.String("end", "", "end time for simulation")

    flag.Parse()

    timeParsedStart, _ := timeParseAny(*startTime)
    timeParsedEnd, _ := timeParseAny(*endTime)

    if (*isSimulation == true) {
        fmt.Println("sim:", *isSimulation)
        if (*startTime != "" || *endTime != "") {
            fmt.Println("start:", timeParsedStart)
            fmt.Println("end:", timeParsedEnd)
        }
    } else {
        fmt.Println("* * * LIVE TRADING * * *")
    }

    simulate := *isSimulation

    // os.Exit(3)

    if simulate == true {
        trading.Simulation(timeParsedStart, timeParsedEnd)
    } else {
        for {
            now := GetWallclockNow()
            wait := 55 - now
            if (wait < 0) {
                wait = wait + 60
            }
            time.Sleep(time.Duration(wait) * time.Second)
            fmt.Println("(Pre) Second executed: ", GetWallclockNow())
            trading.Trade()
            fmt.Println("(Post) Second executed: ", GetWallclockNow())
            time.Sleep(time.Second)
        }
    }
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

func GetWallclockNow() int {
    var t time.Time = time.Now()
    return int(t.Second())
}