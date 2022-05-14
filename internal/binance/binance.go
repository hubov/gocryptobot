package main

import (
    "os"
    "fmt"
    "net/http"
    "io/ioutil"
    "encoding/json"
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

func main() {
    url := configuration.Host + "/api/v3/ping"
    fmt.Println("Hello, " + url)

    req, _ := http.NewRequest("GET", url, nil)
    res, _ := http.DefaultClient.Do(req)
    defer res.Body.Close()
    body, _ := ioutil.ReadAll(res.Body)
    fmt.Println(string(body))

}