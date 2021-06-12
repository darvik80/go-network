package tcp

import (
	"context"
	"darvik80/go-network/network"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type client struct {
	host   string
	log    log.FieldLogger
	ctx    context.Context
	cancel context.CancelFunc
	factory network.PipelineFactory
}

func NewClient(host string, factory network.PipelineFactory) *client {
	ctx, cancel := context.WithCancel(context.Background())

	return &client{
		host,
		log.WithFields(
			log.Fields{
				"module": "tcp-client",
				"addr":   host,
			}),
		ctx,
		cancel,
		factory,
	}
}

func (c *client) Start(factory network.HandlerFactory) error {
	go func() {
		d := net.Dialer{Timeout: time.Second * 5}
		for {
			conn, err := d.Dial("tcp4", c.host)
			if err != nil {
				c.log.Warnf(err.Error())
				time.Sleep(time.Second)
			} else {
				c.handleConnection(c.ctx, factory, conn)
				select {
				case <-c.ctx.Done():
					return
				}
			}
		}
	}()

	c.log.Info("client started")
	return nil
}

func (s *client) handleConnection(ctx context.Context, hf network.HandlerFactory , c net.Conn) {
	chData := make(chan []byte)
	chErr := make(chan error)
	go func(chData chan []byte, chErr chan error) {
		for {
			data := make([]byte, 1024)
			n, err := c.Read(data)
			if err == nil {
				chData <- data[:n]
			} else {
				chErr <- err
				if err == io.EOF {
					return
				}
			}
		}
	}(chData, chErr)

	p := s.factory.Create(network.NewPipeline()).AddLast(hf.Create(ctx, c))
	channel := network.NewChannelWith(context.Background(), p, c)
	defer channel.Close()

	for {
		select {
		case <-ctx.Done():
			channel.Pipeline().FireChannelError(c.Close())
			return
		case data := <-chData:
			channel.Pipeline().FireChannelRead(data)
		case err := <-chErr:
			channel.Pipeline().FireChannelError(err)
			return
		}
	}
}

func (c *client) Shutdown() {
	c.log.Info("shutdown client")
	c.cancel()
}
