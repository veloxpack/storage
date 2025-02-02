package internal

import (
	"context"

	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type service struct{}

func NewService() *service {
	return &service{}
}

// Save stores a file or data into the service
func (s *service) Save(ctx context.Context, req *pb.SaveRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// Stat returns information about a file
func (s *service) Stat(ctx context.Context, req *pb.StatRequest) (*pb.StatResponse, error) {

	// Create a response with some basic file info
	return &pb.StatResponse{
		Stat: &pb.Stat{},
	}, nil
}

// Open streams the file data (this is a simple example, implement actual streaming logic)
func (s *service) Open(req *pb.OpenRequest, stream pb.StorageService_OpenServer) error {
	return nil
}

// Delete removes a file from pb
func (s *service) Delete(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
