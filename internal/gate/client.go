package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/blabtm/extd/internal/api/gate"
	"github.com/blabtm/extd/vcas"
)

type Client struct {
	Conn string

	obs string
	buf []byte
	pkt vcas.Packet
	mux sync.Mutex

	cli gate.ConnectionAdapterClient
}

func NewClient(conn string, cli gate.ConnectionAdapterClient) *Client {
	return &Client{
		Conn: conn,

		cli: cli,
		buf: make([]byte, 0, 0xff),
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	for _, b := range msg {
		if b != 10 {
			c.buf = append(c.buf, b)
			continue
		}

		c.pkt.Stamp.Time = time.Now()

		if err := c.pkt.Unmarshal(c.buf); err != nil {
			return fmt.Errorf("vcas: %v", err)
		}

		if err := c.handlePacket(ctx, &c.pkt); err != nil {
			return err
		}

		if cap(c.buf) > 0xff {
			c.buf = make([]byte, 0, 0xff)
		} else {
			c.buf = c.buf[:0]
		}
	}

	return nil
}

func (c *Client) handlePacket(ctx context.Context, pkt *vcas.Packet) error {
	if c.obs != "" {
		return nil
	}

	if pkt.Topic == "" {
		return fmt.Errorf("unknown topic")
	}

	switch c.pkt.Method {
	case vcas.PUB:
		if err := c.publish(ctx, &c.pkt); err != nil {
			return fmt.Errorf("pub: %v", err)
		}
	case vcas.SUB:
		if err := c.subscribe(ctx, c.pkt.Topic); err != nil {
			return fmt.Errorf("sub: %v", err)
		}
	case vcas.USB:
		if err := c.unsubscribe(ctx, c.pkt.Topic); err != nil {
			return fmt.Errorf("usub: %v", err)
		}
	case vcas.GET:
		if err := c.get(ctx, c.pkt.Topic); err != nil {
			return fmt.Errorf("get: %v", err)
		}
	default:
		return fmt.Errorf("unknown method")
	}

	return nil
}

func (c *Client) publish(ctx context.Context, pkt *vcas.Packet) error {
	pay, err := json.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.cli.Publish(ctx, &gate.PublishRequest{
		Conn:    c.Conn,
		Topic:   pkt.Topic,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		return fmt.Errorf("cli: %v", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("req: %v", res.Message)
	}

	return nil
}

func (c *Client) subscribe(ctx context.Context, top string) error {
	res, err := c.cli.Subscribe(ctx, &gate.SubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
		Qos:   2,
	})

	if err != nil {
		return fmt.Errorf("cli: %v", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("req: %v", res.Message)
	}

	return nil
}

func (c *Client) unsubscribe(ctx context.Context, top string) error {
	res, err := c.cli.Unsubscribe(ctx, &gate.UnsubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
	})

	if err != nil {
		return fmt.Errorf("cli: %v", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("req: %v", res.Message)
	}

	return nil
}

func (c *Client) get(ctx context.Context, top string) error {
	err := c.subscribe(ctx, top)

	if err != nil {
		return fmt.Errorf("sub: %v", err)
	}

	c.obs = top

	time.AfterFunc(5*time.Second, func() {
		c.mux.Lock()
		defer c.mux.Unlock()

		if c.obs != "" {
			c.pkt.Topic = c.obs
			c.pkt.Stamp.Time = time.Now()
			c.pkt.Value = ""

			c.unsubscribe(context.Background(), c.obs)
			c.send(context.Background(), &c.pkt)

			c.obs = ""
		}
	})

	return nil
}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *gate.Message) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.obs != "" {
		if c.obs != msg.Topic {
			return nil
		}

		c.obs = ""

		if err := c.unsubscribe(ctx, msg.Topic); err != nil {
			return fmt.Errorf("usb: %v", err)
		}
	}

	c.pkt.Topic = msg.Topic
  c.pkt.Value = ""

	if err := json.Unmarshal(msg.Payload, &c.pkt); err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	if err := c.send(ctx, &c.pkt); err != nil {
		return fmt.Errorf("send: %v", err)
	}

	return nil
}

func (c *Client) send(ctx context.Context, pkt *vcas.Packet) error {
	pkt.Method = vcas.PUB
	pay, err := pkt.Marshal(make([]byte, 0))

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.cli.Send(ctx, &gate.SendBytesRequest{
		Conn:  c.Conn,
		Bytes: pay,
	})

	if err != nil {
		return fmt.Errorf("cli: %v", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("req: %v", res.Message)
	}

	return nil
}
