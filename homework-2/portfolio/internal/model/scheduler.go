package model

import (
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/logging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ScheduledExecutor struct {
	delay  time.Duration
	ticker time.Ticker
	stop   chan int
}

func NewScheduledExecutor(delay, tick time.Duration) ScheduledExecutor {
	return ScheduledExecutor{
		delay:  delay,
		ticker: *time.NewTicker(tick),
		stop:   make(chan int),
	}
}

func (exec *ScheduledExecutor) Stop() {
	go func() {
		exec.stop <- 1
	}()
}

func (exec *ScheduledExecutor) close() {
	close(exec.stop)
	exec.ticker.Stop()
}

func (exec ScheduledExecutor) Run(task func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer func() {
			logging.Info("Stopped executing background task")
		}()
		time.Sleep(exec.delay)
		for {
			select {
			case <-exec.ticker.C:
				task()
				break
			case <-exec.stop:
				exec.close()
				return
			case <-sigChan:
				exec.Stop()
			}
		}
	}()
}
