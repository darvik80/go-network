package tcp

import (
	"context"
	"github.com/darvik80/go-network/network"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"time"
)

type client struct {
	host   string
	port   int
	log    log.FieldLogger
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(host string, port int) *client {
	ctx, cancel := context.WithCancel(context.Background())

	return &client{
		host,
		port,
		log.WithFields(
			log.Fields{
				"module": "tcp-client",
				"addr":   host,
				"port":   port,
			}),
		ctx,
		cancel,
	}
}

func (c *client) Start(h func(p network.Pipeline) network.Pipeline) error {
	go func() {
		d := net.Dialer{Timeout: time.Second * 5}
		for {
			conn, err := d.Dial("tcp4", c.host+":"+strconv.Itoa(c.port))
			if err != nil {
				c.log.Warnf(err.Error())
				time.Sleep(time.Second)
			} else {
				c.handleConnection(c.ctx, h, conn)
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

func (s *client) handleConnection(ctx context.Context, handler func(p network.Pipeline) network.Pipeline, c net.Conn) {
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

	channel := network.NewChannelWith(context.Background(), handler(network.NewPipeline()), c)
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
