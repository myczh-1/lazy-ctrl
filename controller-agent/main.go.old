package main

import (
	"fmt"
	"log"
	"net"

	"github.com/myczh-1/lazy-ctrl-agent/internal/grpcserver"
	"github.com/myczh-1/lazy-ctrl-agent/internal/handler"
	pb "github.com/myczh-1/lazy-ctrl-agent/proto"
	
	"google.golang.org/grpc"
)

func main() {
	err := handler.LoadCommands("config/commands.json")
	if err != nil {
		log.Fatal("加载 commands.json 失败：", err)
	}

	lis, err := net.Listen("tcp", ":7070")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	controllerServer := grpcserver.NewControllerServer()
	pb.RegisterControllerServiceServer(s, controllerServer)

	fmt.Println("gRPC Agent 正在运行：localhost:7070")
	log.Fatal(s.Serve(lis))
}

