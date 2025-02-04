package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/blabtm/emqx-gate/api"
	"github.com/blabtm/emqx-gate/vcas"
)

type client struct {
	conn string
	obs  string
	buf  []byte
	pkt  vcas.Packet
	mux  sync.Mutex
	now  func() time.Time
	cli  gate.ConnectionAdapterClient
}

func newClient(conn string, cli gate.ConnectionAdapterClient) *client {
	return &client{
		conn: conn,
		buf:  make([]byte, 0, 0xff),
		now:  time.Now,
		cli:  cli,
	}
}

func (cli *client) OnReceivedBytes(ctx context.Context, msg []byte) error {
	cli.mux.Lock()
	defer cli.mux.Unlock()

	for _, b := range msg {
		if b != 10 {
			cli.buf = append(cli.buf, b)
			continue
		}

		cli.pkt.Stamp.Time = cli.now()

		if err := cli.pkt.Unmarshal(cli.buf); err != nil {
			return fmt.Errorf("vcas: %w", err)
		}

		if err := cli.handlePacket(ctx, &cli.pkt); err != nil {
			return err
		}

		if cap(cli.buf) > 0xff {
			cli.buf = make([]byte, 0, 0xff)
		} else {
			cli.buf = cli.buf[:0]
		}
	}

	return nil
}

func (cli *client) handlePacket(ctx context.Context, pkt *vcas.Packet) error {
	if cli.obs != "" {
		return nil
	}

	if pkt.Topic == "" {
		return fmt.Errorf("unknown topic")
	}

	switch cli.pkt.Method {
	case vcas.PUB:
		if err := cli.publish(ctx, &cli.pkt); err != nil {
			return fmt.Errorf("pub: %v", err)
		}
	case vcas.SUB:
		if err := cli.subscribe(ctx, cli.pkt.Topic); err != nil {
			return fmt.Errorf("sub: %v", err)
		}
	case vcas.USB:
		if err := cli.unsubscribe(ctx, cli.pkt.Topic); err != nil {
			return fmt.Errorf("usub: %v", err)
		}
	case vcas.GET:
		if err := cli.get(ctx, cli.pkt.Topic); err != nil {
			return fmt.Errorf("get: %v", err)
		}
	default:
		return fmt.Errorf("unknown method")
	}

	return nil
}

func (cli *client) publish(ctx context.Context, pkt *vcas.Packet) error {
	pay, err := json.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("json: %w", err)
	}

	res, err := cli.cli.Publish(ctx, &gate.PublishRequest{
		Conn:    cli.conn,
		Topic:   pkt.Topic,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("cli: %v", res.Message)
	}

	return nil
}

func (cli *client) subscribe(ctx context.Context, top string) error {
	res, err := cli.cli.Subscribe(ctx, &gate.SubscribeRequest{
		Conn:  cli.conn,
		Topic: top,
		Qos:   2,
	})

	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("cli: %v", res.Message)
	}

	return nil
}

func (cli *client) unsubscribe(ctx context.Context, top string) error {
	res, err := cli.cli.Unsubscribe(ctx, &gate.UnsubscribeRequest{
		Conn:  cli.conn,
		Topic: top,
	})

	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("cli: %v", res.Message)
	}

	return nil
}

func (cli *client) get(ctx context.Context, top string) error {
	err := cli.subscribe(ctx, top)

	if err != nil {
		return fmt.Errorf("sub: %w", err)
	}

	cli.obs = top

	time.AfterFunc(5*time.Second, func() {
		cli.mux.Lock()
		defer cli.mux.Unlock()

		if cli.obs != "" {
			cli.pkt.Topic = cli.obs
			cli.pkt.Stamp.Time = cli.now()
			cli.pkt.Value = ""

			_ = cli.unsubscribe(context.Background(), cli.obs)
			_ = cli.send(context.Background(), &cli.pkt)

			cli.obs = ""
		}
	})

	return nil
}

func (cli *client) OnReceivedMessage(ctx context.Context, msg *gate.Message) error {
	cli.mux.Lock()
	defer cli.mux.Unlock()

	if cli.obs != "" {
		if cli.obs != msg.Topic {
			return nil
		}

		cli.obs = ""

		if err := cli.unsubscribe(ctx, msg.Topic); err != nil {
			return fmt.Errorf("usub: %w", err)
		}
	}

	cli.pkt.Topic = msg.Topic
	cli.pkt.Value = ""

	if err := json.Unmarshal(msg.Payload, &cli.pkt); err != nil {
		return fmt.Errorf("json: %w", err)
	}

	if err := cli.send(ctx, &cli.pkt); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (cli *client) send(ctx context.Context, pkt *vcas.Packet) error {
	pkt.Method = vcas.PUB
	pay, err := pkt.Marshal(make([]byte, 0))

	if err != nil {
		return fmt.Errorf("vcas: %w", err)
	}

	res, err := cli.cli.Send(ctx, &gate.SendBytesRequest{
		Conn:  cli.conn,
		Bytes: pay,
	})

	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("cli: %v", res.Message)
	}

	return nil
}
