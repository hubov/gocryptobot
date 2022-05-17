package main

import (
    "os"
    "fmt"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "time"
    "strconv"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "log"
)

type (
    Client struct {
        httpClient *http.Client
        host string
        Timeout time.Duration
    }
    Configuration struct {
        Host string
        Api_key string
        Api_secret string
        Name string
    }
    Wallet struct {
        Coin string `json:"coin"`
        Name string `json:"name"`
        Free float64 `json:"free,string"`
        Freeze float64 `json:"freeze,string"`
        Locked float64 `json:"locked,string"`
        Storage float64 `json:"storage,string"`
        Withdrawing float64 `json:"withdrawing,string"`
    }
)

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

func ApiClient(timeout time.Duration) *Client {
    client := &http.Client{
        Timeout: timeout,
    }
    return &Client{
        httpClient: client,
        host: configuration.Host,
    }
}

func (c *Client) do(method, endpoint string, params map[string]string) (*http.Response, error) {
    baseURL := fmt.Sprintf("%s%s", c.host, endpoint)
    req, err := http.NewRequest(method, baseURL, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Add("Accept", "application/json")
    req.Header.Add("X-MBX-APIKEY", configuration.Api_key)

    if params == nil {
        params = make(map[string]string)
    }
    params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
    q := req.URL.Query()
    for key, val := range params {
        q.Set(key, val)
    }
    req.URL.RawQuery = q.Encode()

    h := hmac.New(sha256.New, []byte(configuration.Api_secret))
    h.Write([]byte(req.URL.RawQuery))
    signature := hex.EncodeToString(h.Sum(nil))
    req.URL.RawQuery = req.URL.RawQuery + "&signature=" + signature

    return c.httpClient.Do(req) 
}

func (c *Client) GetWallet() (result []Wallet, err error) {
    var resp []Wallet
    res, err := c.do(http.MethodGet, "/sapi/v1/capital/config/getall", nil)
    if err != nil {
        return
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return resp, err
    }
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }
    for i, coin := range resp {
        if (i >= 0) {
            sum := coin.Free + coin.Freeze + coin.Locked + coin.Storage + coin.Withdrawing
            if (sum > 0) {
                result = append(result, coin)
            }
        }
    }
    return
}

func queryAPI(url string) ([]byte) {
    req, _ := http.NewRequest("GET", (configuration.Host + url), nil)
    req.Header.Add("Accept", "application/json")
    req.Header.Add("X-MBX-APIKEY", configuration.Api_key)


    res, _ := http.DefaultClient.Do(req)
    defer res.Body.Close()
    body, _ := ioutil.ReadAll(res.Body)

    return body
}

func decodeJSON(input []byte) (map[string]interface{}) {
    var data map[string]interface{}
    err := json.Unmarshal(input, &data)
    if err != nil {
        panic(err)
    }
    return data
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

func connectionDelay() (int64) {
    serverTime := serverTime()
    localTime := time.Now().UnixMilli()
    diff := localTime - serverTime
    fmt.Println(serverTime, localTime)

    return diff
}

func main() {
    defaultTimeout := time.Second * 10
    client := ApiClient(defaultTimeout)
    wallet, err := client.GetWallet()
    if err != nil {
        log.Fatal(err)
    }
    for i, coin := range wallet {
        if (i >= 0) {
            fmt.Println(coin)
        }
    }
}