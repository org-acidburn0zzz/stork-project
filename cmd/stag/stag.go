package main

import (
	"net"
	"stork/internal/stagrpc/server"
	"google.golang.org/grpc"
	pb "stork/api/protobuf-spec"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:10123")
	if err != nil {
		panic(err.Error())
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterStorkAgentServer(grpcServer, &stagrpc.AgentServer{})

	grpcServer.Serve(lis)
}
