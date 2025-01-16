package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
  dst string
	top string
	frq int
)

func init() {
  flag.StringVar(&dst, "dst", "tcp://localhost:1883", "destination address")
	flag.StringVar(&top, "top", "test", "topic to ping")
	flag.IntVar(&frq, "frq", 1, "ping frequency")
}

func main() {
  flag.Parse()
  slog.Info("pining", "addr", dst, "top", top, "freq", frq)

  cli := mqtt.NewClient(mqtt.NewClientOptions().AddBroker(dst))

  if tok := cli.Connect(); tok.Wait() && tok.Error() != nil {
    panic(tok.Error())
  }

  tck := time.NewTicker(time.Duration(int(time.Second) / frq))
  val := 0.0

  for {
    t := <-tck.C
    stamp := t.UnixMilli()
    val += 0.1

    cli.Publish(top, 0, false, fmt.Sprintf("{\"stamp\":%d,\"value\":%v}", stamp, math.Sin(float64(val))))
  }
}
