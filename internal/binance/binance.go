package binance

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "math"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

type (
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
    ErrorApi struct {
        Code int64 `json:"code"`
        Msg string `json:"msg"`
    }
    ExchangeInfo struct {
        Symbols []Symbol `json:"symbols"`
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
    MarginPair struct {
        Base string `json:"base"`
        IsBuyAllowed bool `json:"isBuyAllowed"`
        IsSellAllowed bool `json:"isSellAllowed"`
        IsMarginTrade bool `json:"isMarginTrade"`
        Quote string `json:"quote"`
        Symbol string `json:"symbol"`
    }
    PriceTicker struct {
        Symbol string `json:"symbol"`
        Price float64 `json:"price,string"`
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
    Symbol struct {
        Symbol string `json:"symbol"`
        BaseAssetPrecision int `json:"baseAssetPrecision"`
        QuoteAssetPrecision int `json:"quoteAssetPrecision"`
        Filters []interface{} `json:"filters"`
    }
    Trade struct {
        BaseSymbol string `json:"base_symbol"`
        QuoteSymbol string `json:"quote_symbol"`
        Interval string
    }
    TradeFills struct {
        Price float64 `json:"price,string"`
        Quantity float64 `json:"qty,string"`
        Commission float64 `json:"commission,string"`
        CommissionAsset string `json:"commissionAsset"`
    }
    TradeHistory struct {
        Commission float64 `json:"commission,string"`
        CommissionAsset string `json:"commissionAsset"`
        IsBestMatch bool `json:"isBestMatch"`
        IsBuyer bool `json:"isBuyer"`
        IsMaker bool `json:"isMaker"`
        Price float64 `json:"price,string"`
        Quantity float64 `json:"qty,string"`
        Symbol string `json:"symbol"`
        Time int64 `json:"time"`
    }
    TradeOrder struct {
        Symbol string `json:"symbol"`
        OrderId int64 `json:"orderId"`
        ClientOrderId string `json:"clientOrderId"`
        Time int64 `json:transactTime`
        OrigQty float64 `json:"origQty,string"`
        ExecutedQty float64 `json:"executedQty,string"`
        Status string `json:"status"`
        Type string `json:"type"`
        Side string `json:"side"`
        Fills []TradeFills
    }
    Wallet struct {
        BaseQuantity float64
        BaseBorrowedQuantity float64
        QuoteQuantity float64
    }
)

var (
    BaseSymbol MarginAsset
    Candles []Candle
    candleStart = "-62135596800000"
    candleEnd = "-62135596800000"
    configuration Configuration
    exchangeInfo ExchangeInfo
    exchange = make(map[string]float64)
    intervals = make(map[string]int)
    LastBuyPrice float64
    QuoteSymbol MarginAsset
    SymbolWorth float64
    TradingPairWallet Wallet
)

func init() {
    SetConfig("", "", "")

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

func (c *Client) ClearData() {
    c.Candles = nil
}

func (c *Client) countIntervals() {
    startInt, _ := strconv.Atoi(c.TimeStart)
    endInt, _ := strconv.Atoi(c.TimeEnd)
    start := time.UnixMilli(int64(startInt))
    end := time.UnixMilli(int64(endInt))
    timeDifference := end.Sub(start)
    c.IntervalsCount = timeDifference.Milliseconds() / int64(intervals[c.Interval])
}

func (c *Client) CreateCandles(data []map[string]string) {
    if len(data) > 0 {
        for i, candle := range data {
            c.Candles = append(c.Candles, Candle{})
            c.Candles[i].Set(str2int(candle["OpenTime"]), str2float(candle["Open"]), str2float(candle["High"]), str2float(candle["Low"]), str2float(candle["Close"]), str2float(candle["Volume"]), str2int(candle["CloseTime"]), str2float(candle["QuoteAssetVolume"]), str2int(candle["TradesNumber"]), str2float(candle["TakerBuyBaseAssetVolume"]), str2float(candle["TakerBuyQuoteAssetVolume"]), str2float(candle["Ignore"]))
        }
    } else {
        panic("Empty data set!")
    }
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

func (c *Client) GetAllMarginPairs() (availablePairs []MarginPair) {
    body, err := c.queryAPI(http.MethodGet, "/sapi/v1/margin/allPairs", nil, true)
    if err = json.Unmarshal(body, &availablePairs); err != nil {
        return
    }

    return
}

func (c *Client) GetCandles(data []map[string]string) (err error) {
    if len(data) >  0 {
        c.CreateCandles(data)
    } else {
        err = c.GetCandlesParams(c.Symbol, c.Interval)
    }

    return
}

func (c *Client) GetCandlesApi(symbol, interval, startTime, endTime, limit string) (resp [][]interface{}, err error) {
    params := make(map[string]string)
    params["limit"] = limit
    params["symbol"] = symbol
    params["interval"] = interval
    if startTime != "-62135596800000" {
        params["startTime"] = startTime
    }
    if endTime != "-62135596800000" {
        params["endTime"] = endTime
    }

    body, err := c.queryAPI(http.MethodGet, "/api/v3/klines", params, false)
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }

    return
}

func (c *Client) GetCandlesParams(symbol, interval string) (err error) {
    var (
        candlesToGet int64
        CandlesArray [][]interface{}
        limit, paramStart string
    )

    if c.IntervalsCount != 0 {
        candlesToGet = c.IntervalsCount
    } else {
        candlesToGet = 500
    }

    paramStart = c.TimeStart

    for candlesToGet > 0 {
        if candlesToGet >= 1000 {
            limit = "1000"
        } else {
            limit = strconv.FormatInt(candlesToGet, 10)
        }

        CandlesArray, err = c.GetCandlesApi(symbol, interval, paramStart, c.TimeEnd, limit)

        candlesLength := len(c.Candles)
        for i, candle := range CandlesArray {
            c.Candles = append(c.Candles, Candle{})
            c.Candles[candlesLength + i].Set(int64(candle[0].(float64)), str2float(candle[1].(string)), str2float(candle[2].(string)), str2float(candle[3].(string)), str2float(candle[4].(string)), str2float(candle[5].(string)), int64(candle[6].(float64)), str2float(candle[7].(string)), int64(candle[8].(float64)), str2float(candle[9].(string)), str2float(candle[10].(string)), str2float(candle[11].(string)))
        }

        if len(CandlesArray) > 0 {
            candlesToGet = candlesToGet - int64(len(CandlesArray))
            paramStart = strconv.FormatInt(c.Candles[len(c.Candles) - 1].CloseTime + 1, 10)
        } else {
            candlesToGet = 0
        }
    }

    return
}

func (c *Client) GetLastTrades() (resp []TradeHistory, err error) {
    params := make(map[string]string)
    params["symbol"] = c.Symbol
    body, err := c.queryAPI(http.MethodGet, "/sapi/v1/margin/myTrades", params, true)
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }

    return
}

func (c *Client) GetOrderPrecision() (resp ExchangeInfo, err error) {
    body, err := c.queryAPI(http.MethodGet, "/api/v3/exchangeInfo?symbol=" + c.Symbol, nil, false)
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }

    return
}

func (c *Client) GetPriceTicker() (resp PriceTicker, err error) {
    params := make(map[string]string)
    params["symbol"] = c.Symbol
    body, err := c.queryAPI(http.MethodGet, "/api/v3/ticker/price", params, false)
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }

    return
}

func (c *Client) GetWallet(isLive bool) {
    var (
        wallet MarginAccount
        trades []TradeHistory
        ticker PriceTicker
        err error
    )

    wg := sync.WaitGroup{}
    wg.Add(1)
    go func() {
        wallet, err = c.MarginAccount()
        if err != nil {
            panic(err)
        }
        wg.Done()
    }()
    wg.Add(1)
    go func() {
        trades, err = c.GetLastTrades()
        if err != nil {
            panic(err)
        }
        wg.Done()
    }()
    wg.Add(1)
    go func() {
        ticker, err = c.GetPriceTicker()
        if err != nil {
            panic(err)
        }
        wg.Done()
    }()
    wg.Add(1)
    go func() {
        exchangeInfo, err = c.GetOrderPrecision()
        for _, filter := range exchangeInfo.Symbols[0].Filters {
            if filter.(map[string]interface{})["filterType"] == "PRICE_FILTER" {
                exchange["minPrice"], _ = strconv.ParseFloat(filter.(map[string]interface{})["minPrice"].(string), 10)
                exchange["maxPrice"], _ = strconv.ParseFloat(filter.(map[string]interface{})["maxPrice"].(string), 10)
                exchange["priceTick"], _ = strconv.ParseFloat(filter.(map[string]interface{})["tickSize"].(string), 10)
            } else if filter.(map[string]interface{})["filterType"] == "LOT_SIZE" {
                exchange["minQty"], _ = strconv.ParseFloat(filter.(map[string]interface{})["minQty"].(string), 10)
                exchange["maxQty"], _ = strconv.ParseFloat(filter.(map[string]interface{})["maxQty"].(string), 10)
                exchange["stepSize"], _ = strconv.ParseFloat(filter.(map[string]interface{})["stepSize"].(string), 10)
            }
        }
        if err != nil {
            panic(err)
        }
        wg.Done()
    }()
    wg.Wait()
    for _, coin := range wallet.UserAssets {
        if (coin.NetAsset != 0) {
            fmt.Println(coin)
        }

        if coin.Asset == configuration.Trade.BaseSymbol {
            BaseSymbol = coin
            TradingPairWallet.BaseQuantity = BaseSymbol.NetAsset
            TradingPairWallet.BaseBorrowedQuantity = BaseSymbol.NetAsset
            if coin.NetAsset != 0 {
                var timestamp int64
                var prices float64
                var quantity float64
                for i := len(trades)-1; i>=0; i-- {
                    if (timestamp == 0) {
                        timestamp = trades[i].Time
                    } else if timestamp != trades[i].Time {
                        break
                    }
                    prices = prices + trades[i].Price * trades[i].Quantity
                    quantity = quantity + trades[i].Quantity
                }
                LastBuyPrice = prices / quantity
            }
            if isLive == true {
                if coin.Borrowed > 0 && coin.Free > 0 {
                    var repay float64
                    if coin.Borrowed < coin.Free {
                        repay = coin.Borrowed
                    } else if coin.Borrowed > coin.Free {
                        repay = coin.Free
                    }
                    fmt.Println(repay)
                    if repay > 0 {
                        c.RepayLoan(repay)
                    }
                }
            }
            SymbolWorth = BaseSymbol.NetAsset * ticker.Price
        } else if coin.Asset == configuration.Trade.QuoteSymbol {
            QuoteSymbol = coin
            TradingPairWallet.QuoteQuantity = QuoteSymbol.NetAsset
        }
    }
}

func (c *Client) MarginAccount() (resp MarginAccount, err error) {
    body, err := c.queryAPI(http.MethodGet, "/sapi/v1/margin/account", nil, true)
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }
    return
}

func (c *Client) MarginBalance() (resp []MarginAsset, err error) {
    account, err := c.MarginAccount()
    for _, asset := range account.UserAssets {
        if asset.NetAsset != 0 {
            resp = append(resp, asset)
        }
    }

    return
}

func (c *Client) OrderMargin(quantity, quoteOrderQty float64, side, sideEffect string) (resp TradeOrder, err error) {
    params := make(map[string]string)
    params["symbol"] = c.Symbol
    params["side"] = side
    params["sideEffectType"] = sideEffect
    params["type"] = "MARKET"
    params["newOrderRespType"] = "FULL"
    if quantity != 0 {
        params["quantity"] = strconv.FormatFloat(quantity, 'f', -1, 64)
    } else {
        if quoteOrderQty != 0 {
            params["quoteOrderQty"] = strconv.FormatFloat(quoteOrderQty, 'f', -1, 64)
        }
    }
    fmt.Println(params)
    // if (Configuration.BuyMax > 0)
    //     params["quantity"] = Configuration.BuyMax

    body, err := c.queryAPI(http.MethodPost, "/sapi/v1/margin/order", params, true)
    fmt.Println(string(body))
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }

    fmt.Println(resp)
    fmt.Println(resp.Fills)
    return
}

func (c *Client) queryAPI(method, endpoint string, params map[string]string, auth bool) (body []byte, err error) {
    refresh := 4
    var res *http.Response
    for refresh > 0 {
        res, err = c.do(method, endpoint, params, auth)
        if err != nil {
            return
        }
        defer res.Body.Close()
        body, err = ioutil.ReadAll(res.Body)
        if err != nil {
            return
        }

        var errorApi ErrorApi
        if err = json.Unmarshal(body, &errorApi); err != nil {
            return
        }
        if errorApi.Code != 0 {
            file, _ := openLogFile("./log/errors.log")
            infoLog := log.New(file, "", log.LstdFlags|log.Lmicroseconds)
            infoLog.Println(endpoint, string(body))
            fmt.Println("API Error", errorApi.Code)
            switch errorApi.Code {
            case -1003: {
                time.Sleep(5 * time.Minute)
                os.Exit(-1003)
            }
            case -1015: {
                time.Sleep(1 * time.Minute)
                os.Exit(-1015)
            }
            default: {
                time.Sleep(10 * time.Second)
                refresh--
            }
            }
        } else {
            refresh = 0
        }
    }

    return
}

func (c *Client) RepayLoan(amount float64) (resp interface{}, err error) {
    params := make(map[string]string)
    params["asset"] = BaseSymbol.Asset
    params["amount"] = float2str(amount)
    body, err := c.queryAPI(http.MethodPost, "/sapi/v1/margin/repay", params, true)
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }

    return
}

func (c *Client) returnCandles() (candles []Candle) {
    return c.Candles
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

func (c *Client) SetTimeframe(start, end int64) {
    c.SetTimeframeOffset(start, end, 500)

    return
}

func (c *Client) SetTimeframeOffset(start, end int64, offset int) {
    start = start - int64(offset * intervals[c.Interval])
    c.TimeStart = strconv.FormatInt(start, 10)
    c.TimeEnd = strconv.FormatInt(end, 10)

    c.countIntervals()

    return
}

func (c *Client) SpotAccount() (resp SpotAccount, err error) {
    body, err := c.queryAPI(http.MethodGet, "/api/v3/account", nil, true)
    if err = json.Unmarshal(body, &resp); err != nil {
        return resp, err
    }
    return
}

func (c *Client) SpotBalance() (resp []SpotAsset, err error) {
    account, err := c.SpotAccount()
    for _, asset := range account.Balances {
        if asset.Free != 0 || asset.Locked != 0 {
            resp = account.Balances
        }
    }

    return
}

func (c *Client) Trade(quantity, quoteOrderQty float64, signal string) {
    command := strings.Split(signal, " ")

    quantity = RoundTradeQuantity(quantity)
    quoteOrderQty = RoundTradeQuantity(quoteOrderQty)

    if command[1] == "SHORT" {
        if command[0] == "Close" || command[0] == "Exit" {
            c.OrderMargin(quantity, 0, "BUY", "AUTO_REPAY")
        } else if (command[0] == "Order") {
            c.OrderMargin(0, quoteOrderQty, "SELL", "MARGIN_BUY")
        }
    } else if command[1] == "LONG" {
        if command[0] == "Close" || command[0] == "Exit" {
            c.OrderMargin(quantity, 0, "SELL", "NO_SIDE_EFFECT")
        } else if (command[0] == "Order") {
            c.OrderMargin(0, quoteOrderQty, "BUY", "NO_SIDE_EFFECT")
        }
    }
}

func GetBaseQuantity() float64 {
    return TradingPairWallet.BaseQuantity
}

func GetQuoteQuantity() float64 {
    return TradingPairWallet.QuoteQuantity
}

func float2str(input float64) (output string) {
    output = strconv.FormatFloat(input, 'f', -1, 64)

    return
}

func openLogFile(path string) (logFile *os.File, err error) {
    logFile, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        return nil, err
    }
    return
}

func RoundTradeQuantity(input float64) (output float64) {
    power := math.Round(1 / exchange["stepSize"])
    output = float64(int(input * power)) / power

    return
}

func SetConfig(base, quote, interval string) {
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

    if base != "" && quote != "" {
        configuration.Trade.BaseSymbol = base
        configuration.Trade.QuoteSymbol = quote
    }
    
    if (interval != "") {
        configuration.Trade.Interval = interval
    }
}

func str2float(input string) (res float64) {
    res, err := strconv.ParseFloat(input, 64);
    if err != nil {
        panic(err)
    }
    return
}

func str2int(input string) (output int64) {
    output, _ = strconv.ParseInt(input, 10, 64)

    return
}