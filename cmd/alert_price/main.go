package main

import (
	"net/http"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"github.com/go-redis/redis"
	"binance/common"
)

type KItems struct {
	Items []interface{}
}

type KlineItems struct {
	Items []KItems
}

// Get all symbol from redis
// then check alert price
// if threshold price is in range of lowest price - highest price
// then alert to telegram group
func main() {

	redisKeyPrefix := "alert_price:"
	binanceBaseAPIPath := "https://api.binance.com"

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// List keys in redis
	listKey, err := redisClient.Keys(fmt.Sprintf("%s*", redisKeyPrefix)).Result()
	if err != nil {
		fmt.Println("Err: ", err)
		return
	}

	for _, redisKey := range listKey {
		mapVal, err := redisClient.HGetAll(redisKey).Result()
		if err != nil {
			fmt.Println("Err: ", err)
			continue
		}

		symbol := mapVal["symbol"]
		fmt.Println("Symbol: ", symbol)

		price, err := strconv.ParseFloat(mapVal["price"], 64)
		if err != nil {
			fmt.Println("Err: ", err)
			continue
		}
		fmt.Println("Price: ", price)

		isAlerted, err := strconv.ParseBool(mapVal["alerted"])
		if err != nil {
			fmt.Println("Err: ", err)
			continue
		}
		fmt.Println("Alerted: ", isAlerted)

		if isAlerted {
			// return
			continue
		}

		//
		// Query binance API
		//
		getURL := fmt.Sprintf("%s%s?symbol=%s&interval=5m&startTime=%d", binanceBaseAPIPath,
			"/api/v1/klines", symbol, (time.Now().Unix()-10*60)*1000)

		resp, err := http.Get(getURL)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		var listKlines []interface{} = make([]interface{}, 0)
		err = json.Unmarshal(respBytes, &listKlines)
		if err != nil {
			fmt.Println("Err: ", err)
			continue
		}

		for _, kline := range listKlines {
			klines := kline.([]interface{})
			var highPrice, lowPrice float64
			for index, _ := range klines {
				if index == 2 {
					highPrice, err = strconv.ParseFloat(klines[index].(string), 64)
					if err != nil {
						fmt.Println("Err: ", err)
						continue
					}
				} else if index == 3 {
					lowPrice, err = strconv.ParseFloat(klines[index].(string), 64)
					if err != nil {
						fmt.Println("Err: ", err)
						continue
					}
				}
			}
			fmt.Println("Symbol: ", symbol, " High: ", highPrice)
			fmt.Println("Symbol: ", symbol, " Low: ", lowPrice)

			// Check price for alert
			if lowPrice <= price && price <= highPrice {
				// Alert
				common.AlertToTelegram(fmt.Sprintf("%s đã vượt ngưỡng %f", symbol, price))

				// Update redis to mark as alerted
				redisClient.HSet(redisKey, "alerted", "true")
				redisClient.HSet(redisKey, "updated_at", time.Now().Format(time.RFC3339))

				// Complete, return
				continue
			}
		}

		redisClient.HSet(redisKey, "updated_at", time.Now().Format(time.RFC3339))

	}

}
