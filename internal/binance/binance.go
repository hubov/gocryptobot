package main

import (
    "os"
    "fmt"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "time"
)

type Configuration struct {
    Host string
    Api_key string
    Api_secret string
    Name string
}

var configuration Configuration

func init() {
    file, err := os.Open("../../config/config.json")
    if err != nil {
        fmt.Println("File reading error", err)
        return
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    configuration = Configuration{}
    err = decoder.Decode(&configuration)
    if err != nil {
      fmt.Println("error:", err)
    }
}

func decodeJSON(input []byte) (map[string]interface{}) {
    var data map[string]interface{}
    err := json.Unmarshal(input, &data)
    if err != nil {
        panic(err)
    }
    return data
}

func queryAPI(url string) ([]byte) {
    req, _ := http.NewRequest("GET", (configuration.Host + url), nil)
    req.Header.Add("Accept": "application/json")
    req.Header.Add("X-MBX-APIKEY", configuration.api_key)

    res, _ := http.DefaultClient.Do(req)
    defer res.Body.Close()
    body, _ := ioutil.ReadAll(res.Body)

    return body
}

func ping() ([]byte) {
    return queryAPI("/api/v3/ping")
}

func serverTime() (int64) {
    result := queryAPI("/api/v3/time")
    data := decodeJSON(result)
    serverTime := data["serverTime"].(float64)
    resultTime := int64(serverTime)

    return resultTime
}

func main() {
    serverTime := serverTime()
    localTime := time.Now().UnixMilli()

    fmt.Println(serverTime, localTime, localTime - serverTime)

}