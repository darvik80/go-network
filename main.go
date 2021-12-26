package main

import (
	"darvik80/go-network/application"
	"darvik80/go-network/middleware"
	"darvik80/go-network/middleware/handler"
	"darvik80/go-network/middleware/route"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	var st = time.Now()
	runtime.GOMAXPROCS(runtime.NumCPU())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	app, err := application.NewApp()
	if err != nil {
		return
	}
	defer app.Shutdown()

	err = app.Setup(
		func(app application.Application) (func(), error) {
			mid, err := middleware.NewMiddleware(
				middleware.WithExchange(app.Exchange()),
				middleware.WithRouter(route.NewPortCodeSortingRouter(app.Db())),
				middleware.WithLinks(app.Config().Links),
				middleware.WithDevices(app.Config().Devices),
				middleware.WithSubscriber(handler.NewEventBusHandler(
					app.Config().Eventbus.Topic, app.Bus(), app.Exchange()),
				),
			)
			if err != nil {
				return nil, err
			}

			return func() {
				mid.Shutdown()
			}, nil
		},
	)
	if err != nil {
		return
	}

	log.Infof("[app] start app, %dms", time.Now().Sub(st).Milliseconds())

	<-sigs
}
