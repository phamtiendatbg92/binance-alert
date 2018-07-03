package main

import (
	"github.com/ducnt114/binance-go"
	"log"
	"time"
	"github.com/ducnt114/binance-alert/common"
	"fmt"
	"strings"
)

func main() {

	binanceClient := binance_go.NewBinanceClient()

	listSymbol, err := binanceClient.GetListSymbol()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Len symbol: ", len(listSymbol))

	startTime := (time.Now().Unix() - 13*24*60*60) * 1000
	endTime := time.Now().Unix() * 1000

	listAlert := make([]string, 0)

	for _, s := range listSymbol {
		if s.QuoteAsset != binance_go.BTCSymbol {
			continue
		}

		data, err := binanceClient.GetCandlestickData(s.Symbol, binance_go.Interval1d, startTime, endTime)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("Symbol: ", s.Symbol, " Len data: ", len(data))
		var totalPrice float64
		var lastPrice float64
		for _, d := range data {
			//log.Printf("Time: %f ClosePrice: %f\n", d.CloseTime, d.ClosePrice)
			totalPrice += d.ClosePrice
			lastPrice = d.ClosePrice
		}

		avgMAPrice := totalPrice / 13.0

		if lastPrice > avgMAPrice {
			listAlert = append(listAlert, s.Symbol)
			//common.AlertToTelegram(fmt.Sprintf("[%s] - MA(13): %.2f - EMA(13): %.2f", s.Symbol, avgMAPrice, 0.0))
			//common.AlertToTelegram(fmt.Sprintf("[%s] - MA(13): %.6f", s.Symbol, avgMAPrice))
			log.Println("Alert to telegram with symbol: ", s.Symbol)
		}
	}

	log.Println("List symbol to alert: ", len(listAlert))

	alertString := strings.Join(listAlert, ",")
	common.AlertToTelegram(fmt.Sprintf("MA/EMA alert: %s", alertString))

	log.Println("Done")

}
