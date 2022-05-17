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
)

var configuration Configuration

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

func (c *Client) SpotBalance() (resp []SpotAsset, err error) {
    account, err := c.SpotAccount()
    resp = account.Balances

    return
}

func (c *Client) SpotAccount() (resp SpotAccount, err error) {
    res, err := c.do(http.MethodGet, "/api/v3/account", nil)
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
    res, err := c.do(http.MethodGet, "/sapi/v1/margin/account", nil)
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

func QueryAPI(url string) ([]byte) {
    req, _ := http.NewRequest("GET", (configuration.Host + url), nil)
    req.Header.Add("Accept", "application/json")
    req.Header.Add("X-MBX-APIKEY", configuration.Api_key)


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