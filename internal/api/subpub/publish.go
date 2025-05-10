package subpub

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	desc "sub-pub/pkg/subpub_v1"
)

// Publish реализует gRPC метод публикации
func (i *Implementation) Publish(ctx context.Context, req *desc.PublishRequest) (*emptypb.Empty, error) {
	logger := i.logger.With(
		zap.String("method", "Publish"),
		zap.String("topic", req.Key),
		zap.String("data", req.Data),
	)

	logger.Info("Received publish request")

	if err := i.pubsub.Publish(req.Key, req.Data); err != nil {
		logger.Error("Failed to publish message",
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "failed to publish: %v", err)
	}

	logger.Info("Message published successfully")
	return &emptypb.Empty{}, nil
}
