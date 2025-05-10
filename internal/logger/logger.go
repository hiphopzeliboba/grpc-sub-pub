package logger

import (
	"go.uber.org/zap"
)

// NewLogger creates a new instance of the logger
func NewLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}