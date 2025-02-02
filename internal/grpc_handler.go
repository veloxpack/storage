package internal

import (
	"context"

	pb "github.com/mediaprodcast/proto/genproto/go/storage/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type grpcHandler struct {
	pb.StorageServiceServer
	svc StorageService
}

func NewGRPCHandler(s *grpc.Server, svc StorageService) {
	pb.RegisterStorageServiceServer(s, &grpcHandler{svc: svc})
}

func (h *grpcHandler) Save(ctx context.Context, p *pb.SaveRequest) (*emptypb.Empty, error) {
	return h.svc.Save(ctx, p)
}

func (h *grpcHandler) Stat(ctx context.Context, p *pb.StatRequest) (*pb.StatResponse, error) {
	return h.svc.Stat(ctx, p)
}

func (h *grpcHandler) Open(p *pb.OpenRequest, stream pb.StorageService_OpenServer) error {
	return h.svc.Open(p, stream)
}

func (h *grpcHandler) Delete(ctx context.Context, p *pb.DeleteRequest) (*emptypb.Empty, error) {
	return h.svc.Delete(ctx, p)
}
