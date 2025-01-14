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

type packet struct {
	Topic  string      `vcas:"name,n" json:"-"`
	Stamp  vcas.Time   `vcas:"time,t" json:"stamp"`
	Method vcas.Method `vcas:"method,meth,m" json:"-"`
	Value  any         `vcas:"val,value,v" json:"value,omitempty"`
	Units  any         `vcas:"units" json:"-"`
	Descr  any         `vcas:"descr" json:"-"`
	Type   any         `vcas:"type" json:"-"`
}

type Client struct {
	Conn string

	obs string
	buf []byte
	pkt packet
	mux sync.Mutex

	cli gate.ConnectionAdapterClient
}

func NewClient(conn string, cli gate.ConnectionAdapterClient) *Client {
	buf := make([]byte, 0, 0xff)

	return &Client{
		Conn: conn,

		cli: cli,
		buf: buf,
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
		c.pkt.Value = nil

		if err := vcas.Unmarshal(c.buf, &c.pkt); err != nil {
			return fmt.Errorf("unmarshal: %v", err)
		}

		if err := c.handlePacket(ctx, &c.pkt); err != nil {
			return fmt.Errorf("handle: %v", err)
		}

		if cap(c.buf) > 0xff {
			c.buf = make([]byte, 0, 0xff)
		} else {
			c.buf = c.buf[:0]
		}
	}

	return nil
}

func (c *Client) handlePacket(ctx context.Context, pkt *packet) error {
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
			return fmt.Errorf("sube: %v", err)
		}
	case vcas.USB:
		if err := c.unsubscribe(ctx, c.pkt.Topic); err != nil {
			return fmt.Errorf("usb: %v", err)
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

func (c *Client) publish(ctx context.Context, pkt *packet) error {
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
			c.pkt.Value = nil

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
	c.pkt.Value = nil

	if err := json.Unmarshal(msg.Payload, &c.pkt); err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	if err := c.send(ctx, &c.pkt); err != nil {
		return fmt.Errorf("send: %v", err)
	}

	return nil
}

func (c *Client) send(ctx context.Context, pkt *packet) error {
	pkt.Method = vcas.PUB
	txt, err := vcas.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.cli.Send(ctx, &gate.SendBytesRequest{
		Conn:  c.Conn,
		Bytes: append(txt, 10),
	})

	if err != nil {
		return fmt.Errorf("cli: %v", err)
	}

	if res.Code != gate.ResultCode_SUCCESS {
		return fmt.Errorf("req: %v", res.Message)
	}

	return nil
}
