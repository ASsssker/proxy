package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ASsssker/proxy/internal/apps/requester"
	"github.com/ASsssker/proxy/internal/config"
)

func main() {
	cfg := config.MustLoad()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	app := requester.MustNewRequester(ctx, cfg)
	go app.MustRun(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	app.Stop(ctx)
}
