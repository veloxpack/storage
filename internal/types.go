package internal

import (
	"context"

	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StorageService interface {
	Save(context.Context, *pb.SaveRequest) (*emptypb.Empty, error)
	Stat(context.Context, *pb.StatRequest) (*pb.StatResponse, error)
	Open(*pb.OpenRequest, grpc.ServerStreamingServer[pb.Chunk]) error
	Delete(context.Context, *pb.DeleteRequest) (*emptypb.Empty, error)
}
