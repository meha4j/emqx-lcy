package proc

import (
	context "context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/meha4j/extd/internal/proc/proto"
	"github.com/meha4j/extd/pkg/vcas"
	"go.uber.org/zap"
)

type Packet struct {
	Topic  string      `vcas:"name,n"`
	Stamp  vcas.Time   `vcas:"time,t" json:"timestamp"`
	Method vcas.Method `vcas:"method,meth,m"`
	Value  any         `vcas:"val,value,v" json:"value"`

	Units       string `vcas:"units"`
	Description string `vcas:"descr"`
	Type        string `vcas:"type"`
}

type Observer struct {
	string
}

type Client struct {
	Conn string
	Log  *zap.Logger

	buf []byte
	pkt Packet
	obs Observer
	mut sync.Mutex

	adapter proto.ConnectionAdapterClient
}

func NewClient(conn string, adapter proto.ConnectionAdapterClient, log *zap.Logger) *Client {
	return &Client{
		Conn: conn,
		Log:  log,

		buf: make([]byte, 0, 0xff),

		adapter: adapter,
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, b := range msg {
		if b != 10 {
			c.buf = append(c.buf, b)
			continue
		}

		if err := vcas.Unmarshal(c.buf, &c.pkt); err != nil {
			return fmt.Errorf("unmarshal: %v", err)
		}

		if err := c.handlePacket(ctx); err != nil {
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

func (c *Client) handlePacket(ctx context.Context) error {
	if c.obs.string != "" {
		return nil
	}

	switch c.pkt.Method {
	case vcas.PUB:
		return c.publish(ctx)
	case vcas.SUB:
		return c.subscribe(ctx, c.pkt.Topic)
	case vcas.USB:
		return c.unsubscribe(ctx, c.pkt.Topic)
	case vcas.GET:
		return c.get(ctx, c.pkt.Topic)
	default:
		return fmt.Errorf("method not found")
	}
}

func (c *Client) OnTimerTimeout(ctx context.Context, ttp proto.TimerType) error { return nil }

func (c *Client) OnReceivedMessage(ctx context.Context, msg *proto.Message) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.obs.string != "" {
		if c.obs.string != msg.GetTopic() {
			return nil
		}

		c.obs.string = ""

		if err := c.unsubscribe(ctx, msg.GetTopic()); err != nil {
			return fmt.Errorf("unsubscribe: %v", err)
		}
	}

	c.pkt.Topic = msg.GetTopic()
	err := json.Unmarshal(msg.GetPayload(), &c.pkt)

	if err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	if err := c.send(ctx); err != nil {
		return fmt.Errorf("send: %v", err)
	}

	return nil
}

func (c *Client) send(ctx context.Context) error {
	txt, err := vcas.Marshal(c.pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Send(ctx, &proto.SendBytesRequest{
		Conn:  c.Conn,
		Bytes: txt,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) publish(ctx context.Context) error {
	pay, err := json.Marshal(c.pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Publish(ctx, &proto.PublishRequest{
		Conn:    c.Conn,
		Topic:   c.pkt.Topic,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) subscribe(ctx context.Context, top string) error {
	res, err := c.adapter.Subscribe(ctx, &proto.SubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
		Qos:   2,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) unsubscribe(ctx context.Context, top string) error {
	res, err := c.adapter.Unsubscribe(ctx, &proto.UnsubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) get(ctx context.Context, top string) error {
	err := c.subscribe(ctx, top)

	if err != nil {
		return fmt.Errorf("subscribe: %v", err)
	}

	time.AfterFunc(5*time.Second, func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		if c.obs.string != "" {
			c.obs.string = ""

			c.unsubscribe(ctx, top)
			c.empty(top)
			c.send(ctx)
		}
	})

	return nil
}

func (c *Client) empty(top string) {
	c.pkt.Topic = top
	c.pkt.Stamp.Time = time.Now()
	c.pkt.Method = vcas.PUB

	c.pkt.Description = "-"
	c.pkt.Value = "none"
	c.pkt.Units = "-"
	c.pkt.Type = "rw"
}
