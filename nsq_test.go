package main

import (
	"testing"
	"log"
	"github.com/nsqio/go-nsq"
	"net"
	"time"
	"os"
)

func TestNSQProducer(t *testing.T) {
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer("127.0.0.1:4150", config)

	err := w.Publish("test", []byte("test"))
	if err != nil {
		log.Panic("Could not connect")
	}

	w.Stop()
}

func TestNSQConsumer(t *testing.T) {
	config := nsq.NewConfig()
	laddr := "127.0.0.1"
	// so that the test can simulate binding consumer to specified address
	config.LocalAddr, _ = net.ResolveTCPAddr("tcp", laddr+":0")
	// so that the test can simulate reaching max requeues and a call to LogFailedMessage
	config.DefaultRequeueDelay = 0
	// so that the test wont timeout from backing off
	config.MaxBackoffDuration = time.Millisecond * 50

	topicName := "ws_eth_usdt"
	q, _ := nsq.NewConsumer(topicName, "dump_eth_usdt", config)
	q.SetLogger(log.New(os.Stderr, "", log.Flags()), nsq.LogLevelDebug)

	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		log.Println(string(m.Body))
		return nil
	}))

	addr := "127.0.0.1:4150"
	err := q.ConnectToNSQD(addr)
	if err != nil {
		t.Fatal(err)
	}

	stats := q.Stats()
	if stats.Connections == 0 {
		t.Fatal("stats report 0 connections (should be > 0)")
	}

	<-q.StopChan

}
