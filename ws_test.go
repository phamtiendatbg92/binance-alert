package main

import (
	"testing"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"flag"
	"os"
	"os/signal"
	"net/url"
	"binance/entity"
	"encoding/json"
)

func TestWebsocketTicker(t *testing.T) {

	//config := nsq.NewConfig()
	//w, _ := nsq.NewProducer("127.0.0.1:4150", config)
	//defer w.Stop()

	addr := flag.String("addr", "stream.binance.com:9443", "http service address")
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws/linketh@kline_5m"}
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
			log.Printf("recv: %s", message)
			klineItem := &entity.KlineItem{}
			err = json.Unmarshal(message, klineItem)
			if err != nil {
				log.Println("Error: ", err)
				continue
			}

			//if klineItem.Kline.KlineClosed {
			//	err = w.Publish("ws_eth_usdt", message)
			//	if err != nil {
			//		log.Panic("Could not connect")
			//	}
			//	//log.Println("Published message to topic ws_eth_usdt")
			//}
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
