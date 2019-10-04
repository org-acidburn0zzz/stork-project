package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	pb "stork/api/protobuf-spec"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:10123", opts...)
	if err != nil {
		panic(err.Error())
	}

	defer conn.Close()

	client := pb.NewStorkAgentClient(conn)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	ver, err := client.GetVersion(ctx, &pb.VersionParams{})
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("version returned is %s", ver.VersionText)
	}
}
