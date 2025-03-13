package helper

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
)

// setup watcher to invoke cancel on user interrupt
func Init_userInterrupt(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	slog.Info("Registering for Interrupt Signal")
	wg.Add(1)
	go func() {
		slog.Info("Signal Handler Starting...")
		defer func() {
			slog.Info("Signal Handler Exiting...")
			wg.Done()
		}()
		var sig os.Signal
		for {
			select {
			case sig = <-c:
				slog.Info("Received Signal. Aborting.", "signal", sig)
				if sig == os.Interrupt {
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
