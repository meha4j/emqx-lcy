package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/meha4j/emqx-lcy/internal/proc"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("EMQX_LCY_PROC_PORT")

	if port == "" {
		log.Println("Environment variable \"EMQX_LCY_PROC_PORT\" does not presented. Using default one (9111).")

		port = "9111"
	}

	net, err := net.Listen("tcp", fmt.Sprintf(":%v", port))

	if err != nil {
		panic(err)
	}

	rpc := grpc.NewServer()

	proc.RegisterConnectionUnaryHandlerServer(rpc, proc.NewService())
	log.Printf("Listening on %v.\n", net.Addr())
	rpc.Serve(net)
}
