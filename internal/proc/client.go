package proc

import (
	context "context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/paraskun/extd/internal/proc/proto"
	"github.com/paraskun/extd/pkg/vcas"
	"go.uber.org/zap"
)

type Packet struct {
	Topic  string      `vcas:"name,n" json:"-"`
	Stamp  vcas.Time   `vcas:"time,t" json:"timestamp"`
	Method vcas.Method `vcas:"method,meth,m" json:"-"`
	Value  any         `vcas:"val,value,v" json:"value,omitempty"`

	Units       string `vcas:"units" json:"-"`
	Description string `vcas:"descr" json:"-"`
	Type        string `vcas:"type" json:"-"`
}

type observer = string

type Client struct {
	Conn string
	Log  *zap.Logger

	obs observer
	buf []byte
	pkt Packet
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

func (c *Client) handlePacket(ctx context.Context, pkt *Packet) error {
	if c.obs != "" {
		return nil
	}

	if pkt.Topic == "" {
		return fmt.Errorf("unknown topic")
	}

	switch c.pkt.Method {
	case vcas.PUB:
		return c.publish(ctx, &c.pkt)
	case vcas.SUB:
		return c.subscribe(ctx, c.pkt.Topic)
	case vcas.USB:
		return c.unsubscribe(ctx, c.pkt.Topic)
	case vcas.GET:
		return c.get(ctx, c.pkt.Topic)
	default:
		return fmt.Errorf("unknown method")
	}
}

func (c *Client) publish(ctx context.Context, pkt *Packet) error {
	if pkt.Stamp.Time.IsZero() {
		pkt.Stamp.Time = time.Now()
	}

	pay, err := json.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Publish(ctx, &proto.PublishRequest{
		Conn:    c.Conn,
		Topic:   pkt.Topic,
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

	res, err := c.adapter.StartTimer(ctx, &proto.TimerRequest{
		Conn:     c.Conn,
		Type:     proto.TimerType_KEEPALIVE,
		Interval: 5,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) OnTimerTimeout(ctx context.Context, ttp proto.TimerType) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.obs != "" {
		p := &Packet{Topic: c.obs}

		c.unsubscribe(ctx, c.obs)
		c.send(ctx, p)

		c.obs = ""
	}

	return nil
}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *proto.Message) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.obs != "" {
		if c.obs != msg.GetTopic() {
			return nil
		}

		c.obs = ""

		if err := c.unsubscribe(ctx, msg.GetTopic()); err != nil {
			return fmt.Errorf("unsubscribe: %v", err)
		}
	}

	pkt := &Packet{Topic: msg.GetTopic()}
	err := json.Unmarshal(msg.GetPayload(), pkt)

	if err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	if err := c.send(ctx, pkt); err != nil {
		return fmt.Errorf("send: %v", err)
	}

	return nil
}

func (c *Client) send(ctx context.Context, pkt *Packet) error {
	if pkt.Value == nil {
		pkt.Value = "none"
	}

	pkt.Method = vcas.PUB
	pkt.Description = "-"
	pkt.Units = "-"
	pkt.Type = "rw"

	txt, err := vcas.Marshal(pkt)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.adapter.Send(ctx, &proto.SendBytesRequest{
		Conn:  c.Conn,
		Bytes: append(txt, 10),
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}
