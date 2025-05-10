package closer

import (
	"context"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	globalCloser = New(syscall.SIGINT, syscall.SIGTERM)
	logger       *zap.Logger
)

func init() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
}

// Add adds `func() error` callback to the globalCloser
func Add(f ...func() error) {
	globalCloser.Add(f...)
}

// Wait blocks until all closer functions are done
func Wait() <-chan struct{} {
	return globalCloser.done
}

// CloseAll calls all closer functions
func CloseAll() {
	globalCloser.CloseAll()
}

// Closer ...
type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
}

// New returns new Closer, if []os.Signal is specified Closer will automatically call CloseAll when one of signals is received from OS
func New(sig ...os.Signal) *Closer {
	c := &Closer{done: make(chan struct{})}
	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			sig := <-ch
			signal.Stop(ch)
			logger.Info("Received signal to terminate",
				zap.String("signal", sig.String()),
			)
			c.CloseAll()
		}()
	}
	return c
}

// Add func to closer
func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// CloseAll calls all closer functions
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		logger.Info("Starting graceful shutdown")

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Создаем WaitGroup для ожидания завершения всех функций
		var wg sync.WaitGroup
		wg.Add(len(funcs))

		// Создаем канал для сигнализации о завершении
		done := make(chan struct{})

		// Запускаем все функции закрытия
		for _, f := range funcs {
			go func(f func() error) {
				defer wg.Done()
				if err := f(); err != nil {
					logger.Error("Error during shutdown",
						zap.Error(err),
					)
				}
			}(f)
		}

		// Ждем завершения в отдельной горутине
		go func() {
			wg.Wait()
			close(done)
		}()

		// Ждем либо завершения всех функций, либо таймаута
		select {
		case <-done:
			logger.Info("All shutdown functions completed successfully")
		case <-ctx.Done():
			logger.Warn("Shutdown timeout reached, forcing exit")
		}

		logger.Info("Graceful shutdown complete")
		close(c.done)
	})
}
