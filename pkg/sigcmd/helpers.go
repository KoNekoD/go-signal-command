package sigcmd

import (
	"flag"
	"os"

	"github.com/KoNekoD/go-rabbitmq-consumers/pkg/rmqc"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

func ExecConsumerCommand[T any](consumer rmqc.Consumer[T], logger *zap.Logger) {
	run := func(done chan<- struct{}) {
		defer close(done)
		if err := consumer.Init(); err != nil {
			logger.Fatal("init consumer error", zap.Error(err))
		}
	}

	stop := func() {
		consumer.Stop()
	}

	err := RunSignalCommand(run, stop, logger)
	if err != nil {
		logger.Fatal("run command error", zap.Error(err))
	}
}

func ExecCronCommand(
	def gocron.JobDefinition,
	execute func() error,
	logger *zap.Logger,
	opts ...SignalCommandOpt,
) {
	exec := func() {
		if err := execute(); err != nil {
			logger.Fatal("execute command error", zap.Error(err))
		}
	}

	s, _ := gocron.NewScheduler()
	receivedSignal := make(chan os.Signal, 1)

	run := func(done chan<- struct{}) {
		defer close(done)
		immediately := flag.Bool("immediately", false, "")
		flag.Parse()
		if *immediately {
			exec()
		}

		_, err := s.NewJob(def, gocron.NewTask(exec))
		if err != nil {
			logger.Fatal("new job error", zap.Error(err))
		}

		s.Start()
		<-receivedSignal
	}

	stop := func() {
		if err := s.Shutdown(); err != nil {
			logger.Fatal("shutdown command error", zap.Error(err))
		}
	}

	onSignal := func(sig os.Signal) {
		receivedSignal <- sig
	}

	opts = append(opts, WithWrapOnSignal(onSignal))

	err := RunSignalCommand(run, stop, logger, opts...)
	if err != nil {
		logger.Fatal("run command error", zap.Error(err))
	}
}
