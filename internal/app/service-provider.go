package app

import (
	"go.uber.org/zap"
	"sub-pub/internal/api/subpub"
	"sub-pub/internal/config"
	sp "sub-pub/subpub"
)

type serviceProvider struct {
	// pgConfig   config.PGConfig
	grpcConfig config.GRPCConfig
	apiImpl    *subpub.Implementation
	pkg        sp.SubPub
	logger     *zap.Logger
}

func newServiceProvider() *serviceProvider {
	sp := &serviceProvider{}
	sp.initSubPub()
	return sp
}

func (s *serviceProvider) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *serviceProvider) initSubPub() {
	s.pkg = sp.NewSubPub()
}

func (s *serviceProvider) GRPCConfig() config.GRPCConfig {
	if s.grpcConfig == nil {
		cfg, err := config.NewGRPCConfig()
		if err != nil {
			s.logger.Fatal("Failed to get grpc config",
				zap.Error(err),
			)
		}

		s.grpcConfig = cfg
	}

	return s.grpcConfig
}

func (s *serviceProvider) SubPubImpl() *subpub.Implementation {
	if s.apiImpl == nil {
		s.apiImpl = subpub.NewImplementation(s.pkg, s.logger)
	}
	return s.apiImpl
}
