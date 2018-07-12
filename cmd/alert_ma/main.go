package main

import (
	"log"
	"github.com/ducnt114/binance-go"
	"time"
	"fmt"
	"strings"
	"math"
	"binance-alert/common"
)

const (
	MA_DAY              = 15
	NUMBER_CANDLE_STICK = 52
	EMA9                = "ema9"
	EMA12               = "ema12"
	EMA26               = "ema26"
	XANH                = "xanh"
	DO                  = "do"
	TRUNG_TINH          = "trungtinh"
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

func GetMAValue(data []float64) float64 {
	var totalPrice float64
	for i := len(data) - 1; i >= len(data)-MA_DAY; i-- {
		totalPrice += data[i]
	}
	return totalPrice / MA_DAY
}

func CalculateEMA(data []float64, emaType string) ([]float64) {

	var startDay int
	var emaWeight float64
	switch emaType {
	case EMA9:
		emaWeight = 0.2
		startDay = 9
	case EMA12:
		emaWeight = float64(2) / float64(13)
		startDay = 12
	case EMA26:
		emaWeight = float64(2) / float64(27)
		startDay = 26
	}
	var ema26 float64
	var emaArray = make([]float64, len(data))
	for i := 0; i < startDay; i++ {
		ema26 += data[i]
		emaArray[i] = 0
	}
	ema26 = ema26 / float64(startDay)

	for i := startDay; i < len(data); i++ {
		ema26 = ema26*(1-emaWeight) + data[i]*emaWeight
		emaArray[i] = ema26
	}
	return emaArray
}

func GetArrayValueFromData(data []*binance_go.CandlestickData) []float64 {
	var array = make([]float64, len(data))
	for i, d := range data {
		array[i] = d.ClosePrice
	}
	return array
}

func GetMarkOfPushSystem(data []*binance_go.CandlestickData) int {

	var ema26 = CalculateEMA(GetArrayValueFromData(data), EMA26)
	var ema12 = CalculateEMA(GetArrayValueFromData(data), EMA12)

	var deltaEMA = make([]float64, len(ema26)-26)
	var index int
	for i := 26; i < len(ema26); i++ {
		deltaEMA[index] = ema12[i] - ema26[i]
		index++
	}
	var ema9 = CalculateEMA(deltaEMA, EMA9)

	var lengthDelta = len(deltaEMA)
	var lengthEMA9 = len(ema9)

	var beforeState = GetPushingSystemState(deltaEMA, lengthDelta-1, ema9, lengthEMA9-1)
	var currentState = GetPushingSystemState(deltaEMA, lengthDelta, ema9, lengthEMA9)
	if (beforeState == DO) {
		if (currentState == TRUNG_TINH) {
			return 2
		} else if (currentState == XANH) {
			return 1
		} else {
			return 0
		}
	} else if (beforeState == TRUNG_TINH) {
		if (currentState == XANH) {
			return 1
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func GetPushingSystemState(deltaEMA [] float64, lengthDelta int, ema9 [] float64, lengthEMA9 int) string {
	if (deltaEMA[lengthDelta-1]-deltaEMA[lengthDelta-2] > 0 && ema9[lengthEMA9-1]-ema9[lengthEMA9-2] > 0) {
		return XANH
	} else if (deltaEMA[lengthDelta-1]-deltaEMA[lengthDelta-2] < 0 && ema9[lengthEMA9-1]-ema9[lengthEMA9-2] < 0) {
		return DO
	} else {
		return TRUNG_TINH
	}
}
func GetData(symbol string, frame string) []*binance_go.CandlestickData {
	var timeFrame int64
	switch frame {
	case binance_go.Interval4h:
		timeFrame = 4
	case binance_go.Interval1d:
		timeFrame = 24

	}

	startTime := (time.Now().Unix() - (NUMBER_CANDLE_STICK)*timeFrame*60*60) * 1000
	data, _ := binanceClient.GetCandlestickData(symbol, frame, startTime, CurrentTime())
	if (len(data) == NUMBER_CANDLE_STICK-1) {
		startTime = (time.Now().Unix() - (NUMBER_CANDLE_STICK+1)*timeFrame*60*60) * 1000
		data, _ = binanceClient.GetCandlestickData(symbol, frame, startTime, CurrentTime())
	}

	return data
}
func FindFalseBreakCandle(symbol string) bool {
	startTime := (time.Now().Unix() - 5*60*60) * 1000
	data, _ := binanceClient.GetCandlestickData(symbol, binance_go.Interval1h, startTime, CurrentTime())
	for _, d := range data {
		var t1 = math.Abs(d.OpenPrice - d.ClosePrice)
		var t2 = math.Abs(d.ClosePrice - d.LowPrice)
		if (((t2 - t1) / t2) > 0.5) {
			return true
		}
	}
	return false
}
func main() {

	listSymbol, err := binanceClient.GetListSymbol()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Len symbol: ", len(listSymbol))
	listAlert := make([]string, 0)

	var highestSymbol = make([]string, 0)
	var highestPoint int
	for _, s := range listSymbol {
		if s.QuoteAsset != binance_go.BTCSymbol {
			continue
		}
		fmt.Println("processing",s.Symbol)
		data1D := GetData(s.Symbol, binance_go.Interval1d)
		data4h := GetData(s.Symbol, binance_go.Interval4h)
		if (len(data1D) != NUMBER_CANDLE_STICK) {
			continue
		}
		if (len(data4h) != NUMBER_CANDLE_STICK) {
			continue
		}

		var currentPrice = GetCurrentPrice(s.Symbol)

		var totalPoint int

		var avg1Day = GetMAValue(GetArrayValueFromData(data1D))
		if currentPrice > avg1Day {
			totalPoint += 1
		}

		var point1 = GetMarkOfPushSystem(data4h)
		var point2 = GetMarkOfPushSystem(data1D)
		if (point1 == 2 && point2 == 2) {
			totalPoint += 2
		}
		totalPoint += point1
		totalPoint += point2

		var haveFalseBreak = FindFalseBreakCandle(s.Symbol)
		if (haveFalseBreak) {
			totalPoint += 1
		}
		if (totalPoint >= 4) {
			listAlert = append(listAlert, fmt.Sprintf("%s: %s", s.Symbol, totalPoint))
		}
		if(totalPoint > highestPoint){
			highestSymbol = highestSymbol[:0]
			highestSymbol = append(highestSymbol, s.Symbol)
			highestPoint = totalPoint
		}else if(totalPoint == highestPoint){
			highestSymbol = append(highestSymbol, s.Symbol)
		}
	}
	if(len(listAlert) > 0){
		alertString := strings.Join(listAlert, "\n")
		common.AlertToTelegram(fmt.Sprintf("Múc: %s", alertString))
	}


	//fmt.Println("%s", alertString)
	//fmt.Println("%v", highestSymbol)
	log.Println("Done")

}
