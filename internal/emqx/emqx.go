package emqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

type server struct {
	Bind string `json:"bind"`
}

type handler struct {
	Addr string `json:"address"`
}

type listener struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Bind string `json:"bind"`
}

type exProtoGatewayUpdateRequest struct {
	Name       string     `json:"name"`
	Timeout    string     `json:"idle_timeout"`
	Mountpoint string     `json:"mountpoint"`
	Enable     bool       `json:"enable"`
	Statistics bool       `json:"enable_stats"`
	Server     server     `json:"server"`
	Handler    handler    `json:"handler"`
	Listeners  []listener `json:"listeners"`
}

type exHookAddServerRequest struct {
	Name      string `json:"name"`
	Enable    bool   `json:"enable"`
	Addr      string `json:"url"`
	Timeout   string `json:"request_timeout,omitempty"`
	Action    string `json:"failed_action,omitempty"`
	Reconnect string `json:"auto_reconnect,omitempty"`
	PoolSize  int    `json:"pool_size,omitempty"`
}

type Client struct {
	BaseUrl string

	conn *http.Client
	addr string
	name string
	pass string
}

func NewClient(url, name, pass string) (*Client, error) {
	addr, err := net.LookupIP("extd")

	if err != nil {
		return nil, fmt.Errorf("dns lookup: %v", err)
	}

	if len(addr) == 0 {
		return nil, fmt.Errorf("dns lookup: no record")
	}

	return &Client{
		BaseUrl: url,

		conn: &http.Client{},
		addr: addr[0].String(),
		name: name,
		pass: pass,
	}, nil
}

func (c *Client) UpdateExProtoGateway(aport, lport, hport int) error {
	pay := exProtoGatewayUpdateRequest{
		Name:       "exproto",
		Enable:     false,
		Statistics: true,
		Timeout:    "300s",
		Handler: handler{
			Addr: fmt.Sprintf("http://%s:%d", c.addr, hport),
		},
		Server: server{
			Bind: strconv.Itoa(aport),
		},
		Listeners: []listener{
			{
				Name: "default",
				Type: "tcp",
				Bind: strconv.Itoa(lport),
			},
		},
	}

	bin, err := json.Marshal(pay)

	if err != nil {
		return fmt.Errorf("marshal request: %v", err)
	}

	url := c.BaseUrl + "/gateways/exproto"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bin))

	if err != nil {
		return fmt.Errorf("create request: %v", err)
	}

	req.SetBasicAuth(c.name, c.pass)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.conn.Do(req)

	if err != nil {
		return fmt.Errorf("request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 204 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("parse response: %v", err)
		}

		return fmt.Errorf(buf.String())
	}

	return nil
}

func (c *Client) AddExHookServer(hport int) error {
	pay := exHookAddServerRequest{
		Name:   "extd",
		Enable: false,
		Addr:   fmt.Sprintf("http://%s:%d", c.addr, hport),
	}

	bin, err := json.Marshal(pay)

	if err != nil {
		return fmt.Errorf("marshal request: %v", err)
	}

	url := c.BaseUrl + "/exhooks"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bin))

	if err != nil {
		return fmt.Errorf("create request: %v", err)
	}

	req.SetBasicAuth(c.name, c.pass)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.conn.Do(req)

	if err != nil {
		return fmt.Errorf("request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("parse response: %v", err)
		}

		return fmt.Errorf(buf.String())
	}

	return nil
}
