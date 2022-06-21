# gocryptobot v0.1.0

**goCryptoBot** is a bot that automatically trades cryptocurrencies on the Binance.com exchange using the API.
The current version only allows trading on the `Margin` account.
The bot supports opening long and short positions.

## Prerequisites
### Go
Bot was tested in the version 1.18.2 of Go.
```
% go version
```

### Binance API Key
Binance `API Key` and `API Secret` with restriction set to:
- Enable Reading
- Enable Spot & Margin Trading
- Enable Loan, Repay & Transfer

## Installation
```
% git clone https://github.com/hubov/gocryptobot
% cd gocryptobot
```
	
## Configuration
Your bot configuration needs to be in `config/config.json`
Config file template below:
```
{
	"exchange": "binance",
	"account": "margin",
	"host": "https://api.binance.com",
	"api_key": "",
	"api_secret": "",
	"trade": {
		"base_symbol": "",
		"quote_symbol": "",
		"interval": ""
	}
}
```
You need to provide information regarding your account and the pair intend to trade. Below is an example setup for the `BTC/USDT` pair traded in 15 minute intervals.
```
{
	"exchange": "binance",
	"account": "margin",
	"host": "https://api.binance.com",
	"api_key": "PUT_YOUR_API_KEY_HERE",
	"api_secret": "PUT_YOUR_SECRET_HERE",
	"trade": {
		"base_symbol": "BTC",
		"quote_symbol": "USDT",
		"interval": "15m"
	}
}
```
The intervals available at Binance are the following:
- 1m
- 3m
- 5m
- 15m
- 30m
- 1h
- 2h
- 4h
- 6h
- 8h
- 12h
- 1d
- 3d
- 1w
- ~~1M~~ (monthly interval is currenty unavailable in goCryptoBot)

m -> minutes; h -> hours; d -> days; w -> weeks; M -> Month

## Run

### Live trading
```
% go run bot.go
```

### Simulation
You can test your strategy using the trading simulation.
```
% go run bot.go --sim=1 
```

You can perform the simulation for a specific time period. Date format: `YYYY-MM-DD`, time format: `YYYY-MM-DD HH:ii:ss`.
```
% go run bot.go --sim=1 start="2022-05-01" end="2022-05-31"
```

### Force order
You can manually trigger an open or close position using the following parameters:
- `order short` to place a SHORT order
- `order long` to place a LONG order
- `close short` to close a SHORT position
- `close long` to close a LONG position
- `exit short` to exit a SHORT position
- `exit long` to exit a LONG position

>The forced order is placed on your **real trading account** so it will affect your wallet.
```
% go run bot.go --signal="order long" 
```

### Persistent work
You can run the bot as a Linux service to keep it running even after closing the terminal.

1. 
```
% cd /etc/systemd/system
```

2. Create a file named: `gocryptobot.service`
```
[Unit]
Description=goCryptoBot

[Service]
User=<username>
WorkingDirectory=/home/<username>/gocryptobot
ExecStart=/usr/local/go/bin/go run bot.go
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

3. Reload daemon
```
% sudo systemctl daemon-reload
```

4. Start your service
```
% sudo systemctl start gocryptobot.service
```

5. Check status
```
% sudo systemctl status gocryptobot.service
```

6. Enable new service on every reboot
```
% sudo systemctl enable gocryptobot.service
```

## Mechanics
### Trading
**LONG**
Long positions are always taken using the total value of the quote currency (second in the currency pair). When you close a position, all of the base currency (the first currency in the currency pair) is sold.

>If you are trading the BTC/USDT pair and you have 100 USDT on your account. After meeting the purchase conditions, the bot will place an order for BTC worth 100 USDT. After the sale condition is met, it will sell all of its BTC for USDT.

**SHORT**
To open a short position, a loan is taken with a value corresponding to the quote currency held (the second in the currency pair).

>You trade BTC/USDT with 100 USDT in your account. If the condition for opening a short position is met, the bot will automatically borrow BTC for 100 USDT and open the position.
If the condition to close the short position is met, the bot will buy BTC of the loan amount and automatically pay it back.

### Strategy
The trading strategy is set in `Strategy.go`
There are 8 functions required to implement a strategy and make decisions about opening and closing positions.
`GetData()` is responsible for getting data that the strategy requires
`Calculate()` is responsible for performing calculations on the data obtained from the `GetData()`

**Open position**
`SignalOrderLong()` evaluates whether the conditions for opening a long position are met.
`SignalOrderShort()` corresponds to the function above, but in the context of a short position.

**Close position**
`SignalCloseLong()` evaluates whether the conditions for closing a long position are met.
`SignalCloseShort()` is the equivalent of the above function for short position
The results of the above functions affect the bot's activities only in the intervals set for the bot. For example, once every 15 minutes.

**Exit position** (stop limit/stop loss)
There are two special functions that can trigger a position close if their conditions are met, regardless of the bot's trading interval. They perform the functions of `stop limit` and` stop loss`. They are called up every minute.
`SignalExitLong()` closes a long position if its conditions are met.
`SignalExitShort()` is the short position equivalent of the above function

### Default strategy
The bot has an implemented simple strategy based on:
- Simple Moving Average (SMA)
- Pivot Points (PP)
- Relative Strength Index (RSI)
- Stop limit: 10%
- Stop loss: 5%

You can adjust it according to your needs or completely change it in `Strategy.go`

### Indicators
The bot uses the [Indicator](https://github.com/cinar/indicator) package with a wide selection of trading indicators ready to be used goCryptoBot's strategies. You can also add your own in the function `Calculate()`

## Upgrading the bot
```
git pull origin
```