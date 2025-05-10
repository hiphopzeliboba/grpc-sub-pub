package subpub

import (
	"go.uber.org/zap"
	desc "sub-pub/pkg/subpub_v1"
	"sub-pub/subpub"
)

type Implementation struct {
	desc.UnimplementedPubSubV1Server
	pubsub subpub.SubPub
	logger *zap.Logger
}

func NewImplementation(pubsub subpub.SubPub, logger *zap.Logger) *Implementation {
	return &Implementation{
		pubsub: pubsub,
		logger: logger,
	}
}
