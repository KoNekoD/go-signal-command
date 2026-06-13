# Go Signal Command

Go helper package for running long-lived commands and gracefully stopping them on OS signals.

## Install

```bash
go get github.com/KoNekoD/go-signal-command
```

## Usage

Example for fiber api:

```go
package main

import (
	"context"
	"github.com/KoNekoD/go-signal-command/pkg/sigcmd"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	const addr = "0.0.0.0:8080"
	
	logger := helpers.SetupObservability("api")

	r := fiber.New(fiber.Config{})

	controllers.InitAllControllers(logger, r)

	srv := &fasthttp.Server{Handler: r.Handler()}

	run := func(done chan<- struct{}) {
		defer close(done)
		if err := srv.ListenAndServe(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to listen", zap.Error(err))
		}
	}

	stop := func() {
		logger.Info("shutdown server with timeout of minute")

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if err := srv.ShutdownWithContext(ctx); err != nil {
			logger.Error("graceful shutdown of server failed", zap.Error(err))
		}
	}

	err := sigcmd.RunSignalCommand(run, stop, logger)
	if err != nil {
		logger.Fatal("run api failed", zap.Error(err))
	}
}
```

### Consumer

```go
package main

import (
	"github.com/KoNekoD/go-signal-command/pkg/sigcmd"
)

func main() {
	logger := helpers.SetupObservability("consume-update")

	consumer := consumers.NewUpdateConsumer(logger)

	sigcmd.ExecConsumerCommand(consumer, logger)
}
```

### Cron

```go
package main

import (
	"os"
	"time"

	"github.com/KoNekoD/go-signal-command/pkg/sigcmd"
	"github.com/go-co-op/gocron/v2"
)

var job = gocron.WeeklyJob(
	1,
	gocron.NewWeekdays(
		time.Sunday,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
	),
	gocron.NewAtTimes(
		gocron.NewAtTime(0, 15, 0),
	),
)

func main() {
	logger := helpers.SetupObservability("planned-update")

	updater := services.NewUpdater(logger)

	execute := func() error { return updater.Update() }

	onSignal := func(sig os.Signal) { updater.Stop() }

	sigcmd.ExecCronCommand(job, execute, logger, sigcmd.WithWrapOnSignal(onSignal))
}
```
