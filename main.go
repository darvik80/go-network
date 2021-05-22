package main

import (
	"github.com/darvik80/go-network/logging"
	"github.com/darvik80/go-network/net"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	logging.Setup()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Start app")

	server := net.NewServer("0.0.0.0", 5000)
	err := server.Start(func(p net.Pipeline) net.Pipeline {
		p.AddLast(
			net.NewLogger(),
			net.ReadIdleHandler(time.Second*60),
			net.WriteIdleHandler(time.Second*60),
			net.OutboundHandlerFunc(func(ctx net.OutboundContext, message net.Message) {
				switch m := message.(type) {
				case string:
					ctx.Write([]byte("[" + m + "]\r\n"))
				default:
					log.Warn(ctx.Channel().RemoteAddr().String(), "drop message: ", m)
				}
			}),
			net.InboundHandlerFunc(func(ctx net.InboundContext, message net.Message) {
				str := strings.ToLower(strings.Trim(string(message.([]byte)), " \r\n"))
				log.Info("handle: ", str)
				ctx.Write("Ping")
				switch str {
				case "quit":
					ctx.Write("Bye...")
					ctx.Close(nil)
				case "ping":
					ctx.Write("pong")
				default:
					ctx.Write(str)
				}
			}),
			net.EventHandlerFunc(func(ctx net.EventContext, event net.Event) {
				switch event.(type) {
				case net.ReadIdleEvent:
					log.Info("read idle: ", ctx.Channel().RemoteAddr().String())
				case net.WriteIdleEvent:
					log.Info("write idle: ", ctx.Channel().RemoteAddr().String())
					ctx.Write("ALIVE")
				default:
				}
			}),
		)
		return p
	})

	if err != nil {
		return
	}

	<-sigs

	log.Info("Shutdown app")
}
