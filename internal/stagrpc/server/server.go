package stagrpc

import (
	"context"
	pb "stork/api/protobuf-spec"
)

type AgentServer struct {
}

func (s *AgentServer) GetVersion(ctx context.Context, params *pb.VersionParams) (*pb.StorkVersion, error) {
	return &pb.StorkVersion{VersionText: "1.2"}, nil
}

