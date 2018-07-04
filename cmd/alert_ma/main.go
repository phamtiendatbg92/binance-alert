package main

import (
	"log"
	"github.com/ducnt114/binance-go"
	"fmt"
	"time"
	"strings"
	"binance-alert/common"
)

const dayDelay = 8
const NUMBER_CANDLE_STICK = 13

var binanceClient = binance_go.NewBinanceClient()

func CurrentTime() int64 {
	return time.Now().Unix() * 1000
}

func CheckMA(currentPrice float64, avgPrice float64, maxDelta float64) bool {
	if currentPrice <= avgPrice {
		return false
	} else {
		var delta = (currentPrice - avgPrice) / avgPrice
		if delta < maxDelta {
			return true
		}
		return false
	}
}
func GetCurrentPrice(symbol string) float64 {
	currentData, _ := binanceClient.GetCandlestickData(symbol, binance_go.Interval1d, (time.Now().Unix()-24*3600)*1000, CurrentTime())
	if len(currentData) > 0 {
		return currentData[0].ClosePrice
	} else {
		return 0
	}
}

func GetMAValue(symbol string, frame int64) float64 {
	startTime := (time.Now().Unix() - (dayDelay+NUMBER_CANDLE_STICK)*frame*60*60) * 1000
	endTime := (time.Now().Unix() - dayDelay*frame*60*60) * 1000
	var interval string
	switch frame {
	case 168: //week
		interval = binance_go.Interval1w
	case 24: // day
		interval = binance_go.Interval1d
	case 4:
		interval = binance_go.Interval4h
	case 1:
		interval = binance_go.Interval1h
	}

	data, err := binanceClient.GetCandlestickData(symbol, interval, startTime, endTime)
	if err != nil {
		log.Println(err)
		return 0
	}
	var totalPrice float64
	for _, d := range data {
		totalPrice += d.ClosePrice
	}
	return totalPrice / 13
}

// check buy point in frame 4h
func FindBuyPoint(symbol string) bool {

	var avgPrice = GetMAValue(symbol, 4)
	var currentPrice = GetCurrentPrice(symbol)

	// current price shows uptrend in 4h frame (5%)
	return CheckMA(currentPrice, avgPrice, 0.05)
}

func main() {

	listSymbol, err := binanceClient.GetListSymbol()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Len symbol: ", len(listSymbol))
	listAlert := make([]string, 0)

	for _, s := range listSymbol {
		if s.QuoteAsset != binance_go.BTCSymbol {
			continue
		}

		var avg1Day = GetMAValue(s.Symbol, 24)
		var currentPrice = GetCurrentPrice(s.Symbol)
		// Make sure current is not higher than average price too much (15%)
		if CheckMA(currentPrice, avg1Day, 0.15) {
			// check frame 4h to find best point to buy
			if FindBuyPoint(s.Symbol) {
				listAlert = append(listAlert, s.Symbol)
				log.Println("Alert to telegram with symbol: ", s.Symbol)
			}
		}
	}

	alertString := strings.Join(listAlert, ",")
	//fmt.Println("MA/EMA alert: %s", alertString)
	common.AlertToTelegram(fmt.Sprintf("Múc: %s", alertString))

	log.Println("Done")

}
