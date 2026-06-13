package sigcmd

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type SignalCommandOptions struct {
	onSignal func(os.Signal)
}

type SignalCommandOpt func(*SignalCommandOptions)

func WithWrapOnSignal(v func(os.Signal)) SignalCommandOpt {
	return func(o *SignalCommandOptions) {
		oldOnSignal := o.onSignal

		o.onSignal = func(sig os.Signal) {
			v(sig)

			if oldOnSignal != nil {
				oldOnSignal(sig)
			}
		}
	}
}

func RunSignalCommand(run func(done chan<- struct{}), stop func(), logger *zap.Logger, opts ...SignalCommandOpt) error {
	o := &SignalCommandOptions{}

	signals := []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM}

	o.onSignal = func(sig os.Signal) {
		logger.Info("stop signal received, program will stop", zap.String("sig", sig.String()))
	}

	for _, opt := range opts {
		opt(o)
	}

	logger.Info("started")

	done := make(chan struct{})
	go run(done)

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, signals...)

	select {
	case <-done: // in case of normal exit from running function
	case sig := <-sigCh: // in case of signal received
		o.onSignal(sig)
		stop()
		<-done
	}

	logger.Info("exited")

	return nil
}
