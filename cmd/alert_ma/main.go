package main

import (
	"log"
	"github.com/ducnt114/binance-go"
	"time"
	"fmt"
)

const (
	dayDelay            = 8
	NUMBER_CANDLE_STICK = 13
	EMA9                = "ema9"
	EMA12               = "ema12"
	EMA26               = "ema26"
)

var binanceClient = binance_go.NewBinanceClient()

func CurrentTime() int64 {
	return time.Now().Unix() * 1000
}

func CheckMA(currentPrice float64, avgPrice float64, maxDelta float64) bool {
	if currentPrice <= avgPrice || avgPrice == 0 {
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
		//startTime = (time.Now().Unix() - (dayDelay+NUMBER_CANDLE_STICK + 1)*frame*60*60) * 1000
	case 1:
		interval = binance_go.Interval1h
	}

	data, err := binanceClient.GetCandlestickData(symbol, interval, startTime, endTime)
	fmt.Println(len(data))
	if (len(data) == 12) {
		startTime = (time.Now().Unix() - (dayDelay+NUMBER_CANDLE_STICK+1)*frame*60*60) * 1000
		data, err = binanceClient.GetCandlestickData(symbol, interval, startTime, endTime)
	}
	if len(data) != 13 {
		return 0
	}
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

func CalculateEMA(data []float64, emaType string) float64 {

	if (len(data) != 52) {
		return 0
	}
	var startDay int
	var emaWeight float64
	switch emaType {
	case EMA9:
		emaWeight = 0.2
		startDay = 9
	case EMA12:
		emaWeight = 0.15
		startDay = 12
	case EMA26:
		emaWeight = 0.075
		startDay = 26
	}

	var ema26 float64

	for i := 0; i < startDay; i++ {
		ema26 += data[i]
	}
	ema26 = ema26 / float64(startDay)

	for i := startDay; i < len(data); i++ {
		ema26 = ema26*(1-emaWeight) + data[i]*emaWeight
	}
	return ema26
}

func GetArrayValueFromData(data []*binance_go.CandlestickData) []float64 {
	var array = make([]float64, len(data))
	for i, d := range data {
		array[i] = d.ClosePrice
	}
	return array
}
func main() {
	/*
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
			if avg1Day == 0{
				continue
			}
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
		//common.AlertToTelegram(fmt.Sprintf("Múc: %s", alertString))
		fmt.Println("%s",alertString)
		log.Println("Done")
	*/

	startTime := (time.Now().Unix() - (52)*24*60*60) * 1000
	endTime := time.Now().Unix() * 1000

	data, _ := binanceClient.GetCandlestickData("BCCBTC", binance_go.Interval1d, startTime, endTime)
	var ema26 = CalculateEMA(GetArrayValueFromData(data), EMA26)
	var ema12 = CalculateEMA(GetArrayValueFromData(data), EMA12)
	var ema9 = CalculateEMA(GetArrayValueFromData(data), EMA9)
	//var deltaEMA = ema12 - ema26
	//var ema26 = CalculateEMA(GetArrayValueFromData(data), EMA26)
	fmt.Println( ema26, ema12, ema9)
}
