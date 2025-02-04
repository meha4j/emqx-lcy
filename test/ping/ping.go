package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	host string
	src  int
	dst  int
	num  int
)

type event struct {
	Stamp int64  `json:"stamp"`
	Value string `json:"value"`
}

func init() {
	flag.StringVar(&host, "h", "localhost", "")
	flag.IntVar(&src, "src", 20041, "")
	flag.IntVar(&dst, "dst", 1883, "")
	flag.IntVar(&num, "num", 1000, "")
}

func main() {
	flag.Parse()
	slog.Info("pinging", "host", host, "src", src, "dst", dst, "num", num)

	cli := mqtt.NewClient(mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", host, dst)),
	)

	if tok := cli.Connect(); tok.Wait() && tok.Error() != nil {
		log.Fatal(tok.Error())
	}

	est := 0.0
	rec := 0

	wg := sync.WaitGroup{}
	wg.Add(1)

	cli.Subscribe("/test/ping", 0, func(cli mqtt.Client, msg mqtt.Message) {
		r := time.Now().UnixNano()
		e := event{}

		json.Unmarshal(msg.Payload(), &e)

		m, _ := strconv.ParseInt(e.Value, 10, 64)
		est += float64(float64(r-m) / float64(num))
		rec += 1

		if rec == num {
			wg.Done()
		}
	})

	con, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, src))

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < num; i++ {
		t := time.Now().UnixNano()

		if _, err := con.Write([]byte(fmt.Sprintf("name:/test/ping|method:set|val:%d\n", t))); err != nil {
			log.Fatal(err)
		}

    fmt.Printf("\rtotal sent: %d", i + 1)
    time.Sleep(10 * time.Millisecond)
	}

  print("\n")

	wg.Wait()
	fmt.Printf("estimated latency: %f ns.\n", est)
}
