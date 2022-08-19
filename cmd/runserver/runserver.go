package main

import (
	"log"
	"math"
	"net"

	db "comp2206-db"

	flag "github.com/spf13/pflag"
	"google.golang.org/grpc"
)

func main() {
	cfg := flag.StringP("config", "c", "./config.json", "server config")
	addr := flag.StringP("addr", "a", "0.0.0.0:15184", "service address")
	flag.Parse()
	serverCfg, err := db.ReadServerConfig(*cfg)
	if err != nil {
		log.Fatal(err)
	}
	server, err := db.NewServer(serverCfg)
	defer server.Close()
	if err != nil {
		log.Fatal(err)
	}
	lis, _ := net.Listen("tcp", *addr)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt), grpc.MaxSendMsgSize(math.MaxInt))
	db.RegisterServerServiceServer(s, server)
	log.Print("server running...")
	log.Fatal(s.Serve(lis))
}
