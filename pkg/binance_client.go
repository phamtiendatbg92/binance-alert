package pkg

import (
	"fmt"
	"net/http"
	"log"
)

const (
	baseEndpoint = "https://api.binance.com"
)

type binanceClient struct {
	BaseEndpoint string
}

func NewBinanceClient() *binanceClient {
	return &binanceClient{
		BaseEndpoint: baseEndpoint,
	}
}

func (b *binanceClient) GetListSymbol() ([]string, error) {
	res := make([]string, 0)
	getURL := fmt.Sprintf("%s/api/v1/exchangeInfo", b.BaseEndpoint)

	resp, err := http.Get(getURL)
	if err != nil {
		log.Println(err)
		return res, err
	}
	defer resp.Body.Close()


}
