package main

import (
	"github.com/go-redis/redis"
	"fmt"
	"testing"
	"strconv"
	"time"
)

func TestExampleNewClient(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>

	// HSET
	client.HSet("alert_price:TRIGETH", "symbol", "TRIGETH")
	client.HSet("alert_price:TRIGETH", "price", "0.001841")
	client.HSet("alert_price:TRIGETH", "alerted", "false")

	client.HSet("alert_price:XRPETH", "symbol", "XRPETH")
	client.HSet("alert_price:XRPETH", "price", "0.00122")
	client.HSet("alert_price:XRPETH", "alerted", "false")

	// List keys
	listKey, err := client.Keys("alert_price:*").Result()
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range listKey {
		mapVal, err := client.HGetAll(key).Result()
		if err != nil {
			t.Fatal(err)
		}
		price, err := strconv.ParseFloat(mapVal["price"], 64)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Price: ", price)

		isAlert, err := strconv.ParseBool(mapVal["alerted"])
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Alert: ", isAlert)
	}
}

func TestFormatDatetime(t *testing.T) {
	fmt.Println(time.Now().Format(time.RFC3339))
}
