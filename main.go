package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/meha4j/extd/internal/proc"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("EMQX_EXT_PROC_PORT")
	addr := os.Getenv("EMQX_EXT_PROC_CLI_ADDR")

	if port == "" {
		port = "9111"
	}

	net, err := net.Listen("tcp", fmt.Sprintf(":%v", port))

	if err != nil {
		panic(err)
	}

	defer net.Close()

	rpc := grpc.NewServer()
	srv, err := proc.NewService(addr)

	if err != nil {
		panic(err)
	}

	proc.RegisterConnectionUnaryHandlerServer(rpc, srv)
	// auth.RegisterHookProviderServer(rpc, auth.NewService())

	log.Printf("Listening on %v.\n", net.Addr())

	rpc.Serve(net)
}
