package binance

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
)

type (
    Client struct {
        HttpClient *http.Client
        Host string
        Timeout time.Duration
        Symbol string
        Interval string
        TimeStart string
        TimeEnd string
        IntervalsCount int64
        Candles []Candle
    }
    Configuration struct {
        Host string
        ApiKey string `json:"api_key"`
        ApiSecret string `json:"api_secret"`
        Name string
        Trade Trade
    }
    Trade struct {
        BaseSymbol string `json:"base_symbol"`
        QuoteSymbol string `json:"quote_symbol"`
        Interval string
    }
    SpotAccount struct {
        CanTrade bool `json:"canTrade"`
        TakerCommission int64 `json:"takerCommission"`
        Balances []SpotAsset
    }
    SpotAsset struct {
        Asset string `json:"asset"`
        Free float64 `json:"free,string"`
        Locked float64 `json:"locked,string"`
    }
    MarginAccount struct {
        BorrowEnabled bool `json:"borrowEnabled"`
        MarginLevel float64 `json:"marginLevel,string"`
        TotalAssetOfBtc float64 `json:"totalAssetOfBtc,string"`
        TotalLiabilityOfBtc float64 `json:"totalLiabilityOfBtc,string"`
        TotalNetAssetOfBtc float64 `json:"totalNetAssetOfBtc,string"`
        TradeEnabled bool `json:"tradeEnabled"`
        UserAssets []MarginAsset
    }
    MarginAsset struct {
        Asset string `json:"asset"`
        Borrowed float64 `json:"borrowed,string"`
        Free float64 `json:"free,string"`
        Interest float64 `json:"interest,string"`
        Locked float64 `json:"locked,string"`
        NetAsset float64 `json:"netAsset,string"`
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
    Candle struct {
        OpenTime int64
        Open float64
        High float64
        Low float64
        Close float64
        Volume float64
        CloseTime int64
        QuoteAssetVolume float64
        TradesNumber int64
        TakerBuyBaseAssetVolume float64
        TakerBuyQuoteAssetVolume float64
        Ignore float64
    }
)

var configuration Configuration
var candleStart = "-62135596800000"
var candleEnd = "-62135596800000"
var intervals = make(map[string]int)
var Candles []Candle

func init() {
    file, err := os.Open("config/config.json")
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

    intervals["1m"] = 60000
    intervals["3m"] = 180000
    intervals["5m"] = 300000
    intervals["15m"] = 900000
    intervals["30m"] = 1800000
    intervals["1h"] = 3600000
    intervals["2h"] = 7200000
    intervals["4h"] = 14400000
    intervals["6h"] = 21600000
    intervals["8h"] = 28800000
    intervals["12h"] = 43200000
    intervals["1d"] = 86400000
    intervals["3d"] = 259200000
    intervals["1w"] = 604800000
    // intervals["1M"] = 2629800000
}

func ApiClient(timeout time.Duration) *Client {
    client := &http.Client{
        Timeout: timeout,
    }
    return &Client{
        HttpClient: client,
        Host: configuration.Host,
        Symbol: configuration.Trade.BaseSymbol + configuration.Trade.QuoteSymbol,
        Interval: configuration.Trade.Interval,
        TimeStart: "-62135596800000",
        TimeEnd: "-62135596800000",
        IntervalsCount: 0,
    }
}

func (c *Candle) Set(openTime int64, open, high, low, close, volume float64, closeTime int64, quoteAssetVolume float64, tradesNumber int64, takerBuyBaseAssetVolume, takerBuyQuoteAssetVolume, ignore float64) {
    c.OpenTime = openTime
    c.Open = open
    c.High = high
    c.Low = low
    c.Close = close
    c.Volume = volume
    c.CloseTime = closeTime
    c.QuoteAssetVolume = quoteAssetVolume
    c.TradesNumber = tradesNumber
    c.TakerBuyBaseAssetVolume = takerBuyBaseAssetVolume
    c.TakerBuyQuoteAssetVolume = takerBuyQuoteAssetVolume
    c.Ignore = ignore
}

func (c *Client) do(method, endpoint string, params map[string]string, auth bool) (*http.Response, error) {
    baseURL := fmt.Sprintf("%s%s", c.Host, endpoint)
    req, err := http.NewRequest(method, baseURL, nil)
    if err != nil {
        return nil, err
    }
    if params == nil {
        params = make(map[string]string)
    }
    req.Header.Add("Accept", "application/json")
    if auth == true {
        req.Header.Add("X-MBX-APIKEY", configuration.ApiKey)
        params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
    }
    
    q := req.URL.Query()
    for key, val := range params {
        q.Set(key, val)
    }
    req.URL.RawQuery = q.Encode()

    if auth == true {
        h := hmac.New(sha256.New, []byte(configuration.ApiSecret))
        h.Write([]byte(req.URL.RawQuery))
        signature := hex.EncodeToString(h.Sum(nil))
        req.URL.RawQuery = req.URL.RawQuery + "&signature=" + signature
    }

    return c.HttpClient.Do(req) 
}

func (c *Client) SpotBalance() (resp []SpotAsset, err error) {
    account, err := c.SpotAccount()
    resp = account.Balances

    return
}

func StrToFloat(input string) (res float64) {
    res, err := strconv.ParseFloat(input, 64);
    if err != nil {
        panic(err)
    }
    return
}

func (c *Client) SetTimeframe(start, end int64) {
    c.TimeStart = strconv.FormatInt(start, 10)
    c.TimeEnd = strconv.FormatInt(end, 10)

    c.countIntervals()
    if c.IntervalsCount != 0 && c.IntervalsCount < 500 {
        panic("Count of intervals too short. Is: " + strconv.FormatInt(c.IntervalsCount, 10) + " Needs: 500")
    }

    return
}

func (c *Client) countIntervals() {
    startInt, _ := strconv.Atoi(c.TimeStart)
    endInt, _ := strconv.Atoi(c.TimeEnd)
    start := time.UnixMilli(int64(startInt))
    end := time.UnixMilli(int64(endInt))
    timeDifference := end.Sub(start)
    c.IntervalsCount = timeDifference.Milliseconds() / int64(intervals[c.Interval])
}

func (c *Client) GetCandles() (/*resp []Candle, */err error) {
    /*resp, */err = c.GetCandlesParams(c.Symbol, c.Interval)

    return
}

func (c *Client) GetCandlesParams(symbol, interval string) (/*resp []Candle, */err error) {
    params := make(map[string]string)
    var paramStart, paramEnd string
    var candlesToGet int64
    if c.TimeStart != "-62135596800000" {
        paramStart = c.TimeStart
    }
    if paramEnd != "-62135596800000" {
        paramEnd = c.TimeEnd
    }

    if c.IntervalsCount != 0 {
        candlesToGet = c.IntervalsCount
    } else {
        candlesToGet = 500
    }

    for candlesToGet > 0 {
        if candlesToGet >= 1000 {
            params["limit"] = "1000"
        } else {
            params["limit"] = strconv.FormatInt(candlesToGet, 10)
        }
        params["symbol"] = symbol
        params["interval"] = interval
        if paramStart != "-62135596800000" {
            params["startTime"] = paramStart
        }
        if paramEnd != "-62135596800000" {
            params["endTime"] = paramEnd
        }
        // fmt.Println(params)
        res, err := c.do(http.MethodGet, "/api/v3/klines", params, false)
        if err != nil {
            return err
        }
        defer res.Body.Close()
        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
            return /*resp, */err
        }
        // bodyString := string(body)
        // fmt.Println(bodyString)
        var CandlesArray [][]interface{}
        if err = json.Unmarshal(body, &CandlesArray); err != nil {
            return /*resp, */err
        }
        // fmt.Println(CandlesArray)

        candlesLength := len(c.Candles)
        for i, candle := range CandlesArray {
            // candleInstance = new(Candle)
            //resp = append(resp, Candle{})
            //resp[i].Set(int64(candle[0].(float64)), StrToFloat(candle[1].(string)), StrToFloat(candle[2].(string)), StrToFloat(candle[3].(string)), StrToFloat(candle[4].(string)), StrToFloat(candle[5].(string)), int64(candle[6].(float64)), StrToFloat(candle[7].(string)), int64(candle[8].(float64)), StrToFloat(candle[9].(string)), StrToFloat(candle[10].(string)), StrToFloat(candle[11].(string)))
            c.Candles = append(c.Candles, Candle{})
            c.Candles[candlesLength + i].Set(int64(candle[0].(float64)), StrToFloat(candle[1].(string)), StrToFloat(candle[2].(string)), StrToFloat(candle[3].(string)), StrToFloat(candle[4].(string)), StrToFloat(candle[5].(string)), int64(candle[6].(float64)), StrToFloat(candle[7].(string)), int64(candle[8].(float64)), StrToFloat(candle[9].(string)), StrToFloat(candle[10].(string)), StrToFloat(candle[11].(string)))
            // fmt.Println(c.Candles[candlesLength + i].CloseTime)
        }

        candlesToGet = candlesToGet - int64(len(CandlesArray))
        paramStart = strconv.FormatInt(c.Candles[len(c.Candles) - 1].CloseTime + 1, 10)
    }

    // fmt.Println(len(resp))
    return
}

func (c *Client) returnCandles() (candles []Candle) {
    return c.Candles
}

func (c *Client) SpotAccount() (resp SpotAccount, err error) {
    res, err := c.do(http.MethodGet, "/api/v3/account", nil, true)
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
    return
}

func (c *Client) MarginAccount() (result MarginAccount, err error) {
    var resp MarginAccount
    res, err := c.do(http.MethodGet, "/sapi/v1/margin/account", nil, true)
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
    return
}

func (c *Client) GetWallet() (result []Wallet, err error) {
    var resp []Wallet
    res, err := c.do(http.MethodGet, "/sapi/v1/capital/config/getall", nil, true)
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
    for _, coin := range resp {
        sum := coin.Free + coin.Freeze + coin.Locked + coin.Storage + coin.Withdrawing
        if (sum > 0) {
            result = append(result, coin)
        }
    }
    return
}

func QueryAPI(url string) ([]byte) {
    req, _ := http.NewRequest("GET", (configuration.Host + url), nil)
    req.Header.Add("Accept", "application/json")
    req.Header.Add("X-MBX-APIKEY", configuration.ApiKey)


    res, _ := http.DefaultClient.Do(req)
    defer res.Body.Close()
    body, _ := ioutil.ReadAll(res.Body)

    return body
}

func DecodeJSON(input []byte) (map[string]interface{}) {
    var data map[string]interface{}
    err := json.Unmarshal(input, &data)
    if err != nil {
        panic(err)
    }
    return data
}

func Ping() ([]byte) {
    return QueryAPI("/api/v3/ping")
}

func ServerTime() (int64) {
    result := QueryAPI("/api/v3/time")
    data := DecodeJSON(result)
    serverTime := data["serverTime"].(float64)
    resultTime := int64(serverTime)

    return resultTime
}

func ConnectionDelay() (int64) {
    serverTime := ServerTime()
    localTime := time.Now().UnixMilli()
    diff := localTime - serverTime
    fmt.Println(serverTime, localTime)

    return diff
}