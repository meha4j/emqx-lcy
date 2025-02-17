package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
)

type Request struct {
	Name   string `json:"name"`
	Server struct {
		Bind string `json:"bind"`
	} `json:"server"`
	Handler struct {
		Addr string `json:"address"`
	} `json:"handler"`
}

func getEnv(env string) string {
	value, ok := os.LookupEnv(env)

	if !ok {
		panic(fmt.Errorf("env not set: %v", env))
	}

	return value
}

func register(ip net.IP) error {
	aPort := getEnv("EMQX_ADAPTER_PORT")
	eHost := getEnv("EMQX_HOST")
	ePort := getEnv("EMQX_PORT")
	eUser := getEnv("EMQX_USER")
	ePass := getEnv("EMQX_PASS")
	gPort := getEnv("PORT")

	cli := &http.Client{}
	url := fmt.Sprintf("http://%s:%s/api/v5/gateways/exproto", eHost, ePort)
	pay, err := json.Marshal(&Request{
		Name: "exproto",
		Server: struct {
			Bind string "json:\"bind\""
		}{
			Bind: aPort,
		},
		Handler: struct {
			Addr string "json:\"address\""
		}{
			Addr: fmt.Sprintf("http://%s:%s", ip.String(), gPort),
		},
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(pay))

	if err != nil {
		return err
	}

	req.SetBasicAuth(eUser, ePass)
	req.Header.Add("Content-Type", "application/json")

	res, err := cli.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 204 {
		msg, err := io.ReadAll(res.Body)

		if err != nil {
			return err
		}

		return errors.New(string(msg))
	}

	return nil
}

func main() {
	_, network, err := net.ParseCIDR(getEnv("NETWORK"))

	if err != nil {
		panic(err)
	}

	addrs, err := net.InterfaceAddrs()

	if err != nil {
		panic(err)
	}

	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && network.Contains(ip.IP) {
			if err := register(ip.IP); err != nil {
				panic(err)
			}

			cmd := exec.Command("haproxy", "-f", "/usr/local/etc/haproxy/haproxy.cfg")

			if err := cmd.Run(); err != nil {
				panic(err)
			}

			return
		}
	}

	panic("host does not belongs to specified network")
}
