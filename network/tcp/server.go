package tcp

import (
	"context"
	"darvik80/go-network/network"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

var logger = log.WithFields(
	log.Fields{
		"module": "tcp-server",
	})

type server struct {
	host     string
	listener net.Listener
	ctx      context.Context
	factory  network.PipelineFactory
}

func NewServer(host string, factory network.PipelineFactory) *server {
	return &server{
		host,
		nil,
		context.Background(),
		factory,
	}
}

func (s *server) Start(factory network.HandlerFactory) error {
	l, err := net.Listen("tcp4", s.host)
	if err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithCancel(s.ctx)
		for {
			c, err := l.Accept()
			if err != nil {
				if err != net.ErrClosed {
					logger.Warnf("%s, can't accept connection, %s", s.host, err.Error())
				}
				cancel()
				return
			} else {
				handler := factory.Create(ctx, c)
				if handler == nil {
					c.Close()
				} else {
					go s.handleConnection(ctx, handler, c)
				}
			}
		}
	}()

	s.listener = l
	log.WithFields(log.Fields{"network": "tcp-server"}).Info(s.host, ", started")
	return nil
}

func (s *server) handleConnection(ctx context.Context, handler network.Handler, c net.Conn) {
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

	p := s.factory.Create(network.NewPipeline()).AddLast(handler)
	channel := network.NewChannelWith(ctx, p, c)
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

func (s *server) Shutdown() {
	log.WithFields(log.Fields{"network": "tcp-server"}).Info(s.host, ", shutdown")
	if err := s.listener.Close(); err != nil {
		log.WithFields(log.Fields{"network": "tcp-server"}).Warn(s.host, ", failed stop listener: ", err.Error())
	}
}
