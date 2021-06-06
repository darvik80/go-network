package main

import (
	"darvik80/go-network/config"
	"github.com/darvik80/go-network/logging"
	"github.com/darvik80/go-network/network"
	"github.com/darvik80/go-network/network/tcp"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	logging.Setup()

	cfg, err := config.ReadConfig()
	if err != nil {
		log.Error("failed read config ", err)
		return
	}

	log.Info(cfg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info("start app")
	defer log.Info("shutdown app")

	client := tcp.NewClient("0.0.0.0", 5001)
	err = client.Start(func(p network.Pipeline) network.Pipeline {
		p.AddLast(
			network.NewLogger(),
			network.IdleHandler(time.Second*10),
			func(ctx network.OutboundContext, message network.Message) {
				switch m := message.(type) {
				case string:
					ctx.Write([]byte("[" + m + "]\r\n"))
				default:
					log.Warn(ctx.Channel().RemoteAddr().String(), "drop message: ", m)
				}
			},
			func(ctx network.EventContext, event network.Event) {
				switch event.(type) {
				case network.IdleEvent:
					log.Info("read/write idle: ", ctx.Channel().RemoteAddr().String())
					ctx.Write("ALIVE")
				default:
				}
			},
		)

		return p
	})

	if err != nil {
		return
	}
	defer client.Shutdown()

	server := tcp.NewServer("0.0.0.0", 5000)
	err = server.Start(func(p network.Pipeline) network.Pipeline {
		p.AddLast(
			network.NewLogger(),
			network.ReadIdleHandler(time.Second*60),
			network.WriteIdleHandler(time.Second*60),
			func(ctx network.OutboundContext, message network.Message) {
				switch m := message.(type) {
				case string:
					ctx.Write([]byte("[" + m + "]\r\n"))
				default:
					log.Warn(ctx.Channel().RemoteAddr().String(), "drop message: ", m)
				}
			},
			func(ctx network.InboundContext, message network.Message) {
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
			},
			func(ctx network.EventContext, event network.Event) {
				switch event.(type) {
				case network.ReadIdleEvent:
					log.Info("read idle: ", ctx.Channel().RemoteAddr().String())
				case network.WriteIdleEvent:
					log.Info("write idle: ", ctx.Channel().RemoteAddr().String())
					ctx.Write("ALIVE")
				}
			},
		)
		return p
	})

	if err != nil {
		return
	}

	defer server.Shutdown()

	<-sigs
}
