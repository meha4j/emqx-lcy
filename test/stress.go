//go:build integration

package test

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type event struct {
	Stamp int64  `json:"stamp"`
	Value string `json:"value"`
}

func BenchmarkLatency(b *testing.B) {
	b.StopTimer()

	cli := mqtt.NewClient(mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883"),
	)

	if tok := cli.Connect(); tok.Wait() && tok.Error() != nil {
		b.Fatal(tok.Error())
	}

	est := 0.0
	rec := 0
	wg := sync.WaitGroup{}

	wg.Add(1)
	cli.Subscribe("/test/ping", 0, func(cli mqtt.Client, msg mqtt.Message) {
		t := time.Now().UnixNano()
		e := event{}

		json.Unmarshal(msg.Payload(), &e)

		m, _ := strconv.ParseInt(e.Value, 10, 64)
		est += float64(float64(t-m) / float64(b.N))
		rec += 1

		if rec == b.N {
			wg.Done()
		}
	})

	con, err := net.Dial("tcp", "localhost:20041")

	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		t := time.Now().UnixNano()

		if _, err := con.Write([]byte(fmt.Sprintf("name:/test/ping|method:set|val:%d\n", t))); err != nil {
			b.Fatal(err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	b.Logf("estimated lateny: %fns", est)
}
