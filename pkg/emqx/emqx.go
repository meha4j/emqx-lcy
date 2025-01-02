package emqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type Server struct {
	Bind string `json:"bind"`
}

type Handler struct {
	Addr string `json:"address"`
}

type Listener struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Bind string `json:"bind"`
}

type ExProtoGatewayUpdateRequest struct {
	Name       string     `json:"name"`
	Timeout    string     `json:"idle_timeout"`
	Mountpoint string     `json:"mountpoint"`
	Enable     bool       `json:"enable"`
	Statistics bool       `json:"enable_stats"`
	Server     Server     `json:"server"`
	Handler    Handler    `json:"handler"`
	Listeners  []Listener `json:"listeners"`
}

type ExHookServerUpdateRequest struct {
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
	Addr    string

	conn *http.Client
	user string
	pass string
}

func NewClient(url, user, pass string) (*Client, error) {
	return &Client{
		BaseUrl: url,
		Addr:    "localhost",

		conn: &http.Client{},
		user: user,
		pass: pass,
	}, nil
}

func (c *Client) LookupAddress(host string) error {
	addr, err := net.LookupIP(host)

	if err != nil {
		return fmt.Errorf("dns: %v", err)
	}

	if len(addr) == 0 {
		return fmt.Errorf("no record")
	}

	c.Addr = addr[0].String()

	return nil
}

func (c *Client) UpdateExProtoGateway(pay *ExProtoGatewayUpdateRequest) error {
	bin, err := json.Marshal(pay)

	if err != nil {
		return fmt.Errorf("marshal payload: %v", err)
	}

	url := c.BaseUrl + "/gateways/exproto"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bin))

	if err != nil {
		return fmt.Errorf("create request: %v", err)
	}

	req.SetBasicAuth(c.user, c.pass)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.conn.Do(req)

	if err != nil {
		return fmt.Errorf("exec request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 204 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("parse response: %v", err)
		}

		return fmt.Errorf("%v", buf.String())
	}

	return nil
}

func (c *Client) UpdateExHookServer(pay *ExHookServerUpdateRequest) error {
	ok, err := c.CheckExHookServer(pay.Name)

	if err != nil {
		return fmt.Errorf("check server: %v", err)
	}

	bin, err := json.Marshal(pay)

	if err != nil {
		return fmt.Errorf("marshal payload: %v", err)
	}

	url := c.BaseUrl + "/exhooks"
	met := http.MethodPost

	if ok {
		url += "/" + pay.Name
		met = http.MethodPut
	}

	req, err := http.NewRequest(met, url, bytes.NewReader(bin))

	if err != nil {
		return fmt.Errorf("create request: %v", err)
	}

	req.SetBasicAuth(c.user, c.pass)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.conn.Do(req)

	if err != nil {
		return fmt.Errorf("exec request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("parse response: %v", err)
		}

		return fmt.Errorf("%v", buf.String())
	}

	return nil
}

func (c *Client) CheckExHookServer(name string) (bool, error) {
	url := c.BaseUrl + "/exhooks/" + name
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return false, fmt.Errorf("create request: %v", err)
	}

	req.SetBasicAuth(c.user, c.pass)

	res, err := c.conn.Do(req)

	if err != nil {
		return false, fmt.Errorf("request exec: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}
