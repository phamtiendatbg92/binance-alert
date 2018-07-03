package main

import (
	"log"
	"time"
	"flag"
	"os"
	"os/signal"
	"net/url"
	"binance/entity"
	"encoding/json"
	"github.com/gorilla/websocket"
	"strconv"
	"fmt"
	"binance/common"
)

func main() {

	var lastTimeAlert int64 = 0
	var deltaTime int64 = 60

	addr := flag.String("addr", "stream.binance.com:9443", "http service address")
	flag.Parse()
	log.SetFlags(0)

	symbol := "adxeth"       //os.Args[1]
	thresholdStr := "0.0014" //os.Args[2]

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		fmt.Errorf("Threshold not valid")
		return
	}
	fmt.Println("Alert with threshold: ", threshold)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: fmt.Sprintf("/ws/%s@kline_1m", symbol)}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Printf("connected")
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			//log.Printf("recv: %s", message)
			klineItem := &entity.KlineItem{}
			err = json.Unmarshal(message, klineItem)
			if err != nil {
				log.Println("Error: ", err)
				continue
			}

			lowPrice, err := strconv.ParseFloat(klineItem.Kline.LowPrice, 64)
			if err != nil {
				fmt.Println("Error when parse float for lowPrice, detail: ", err)
				continue
			}
			fmt.Println("Low price: ", lowPrice)
			if lowPrice < threshold {
				currentTime := time.Now().Unix()
				if currentTime-lastTimeAlert > deltaTime {
					common.AlertToTelegram(fmt.Sprintf("%s price is less than %v, current: %v", symbol, threshold, lowPrice))
				}
				lastTimeAlert = currentTime
			}

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}
