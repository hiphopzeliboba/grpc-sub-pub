package subpub

import (
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	desc "sub-pub/pkg/subpub_v1"
)

// Subscribe реализует gRPC метод подписки (стрим)
func (i *Implementation) Subscribe(req *desc.SubscribeRequest, stream desc.PubSubV1_SubscribeServer) error {
	logger := i.logger.With(
		zap.String("method", "Subscribe"),
		zap.String("topic", req.Key),
	)

	logger.Info("Received subscribe request")

	// Создаем обработчик
	handler := func(msg interface{}) {
		data, ok := msg.(string)
		if !ok {
			logger.Warn("Received message of invalid type",
				zap.Any("message", msg),
			)
			return
		}

		err := stream.Send(&desc.Event{Data: data})
		if err != nil {
			logger.Error("Failed to send event",
				zap.Error(err),
				zap.String("data", data),
			)
			return
		}

		logger.Info("Event sent successfully",
			zap.String("data", data),
		)
	}

	// Подписываемся
	logger.Info("Creating subscription")
	sub, err := i.pubsub.Subscribe(req.Key, handler)
	if err != nil {
		logger.Error("Failed to create subscription",
			zap.Error(err),
		)
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	logger.Info("Subscription created successfully")
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			logger.Error("Failed to unsubscribe",
				zap.Error(err),
			)
		}
		logger.Info("Unsubscribed")
	}()

	// Блокируем до отмены контекста
	<-stream.Context().Done()
	logger.Info("Stream context done")

	return nil
}
