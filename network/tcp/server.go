package tcp

import (
	"context"
	"darvik80/go-network/network"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

type server struct {
	host     string
	listener net.Listener
	log      log.FieldLogger
	ctx      context.Context
}

func NewServer(host string) *server {
	return &server{
		host,
		nil,
		log.WithFields(
			log.Fields{
				"module": "tcp-server",
				"addr":   host,
			}),
		context.Background(),
	}
}

func (s *server) Start(h func(p network.Pipeline) network.Pipeline) error {
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
					s.log.Warnf("can't accept connection, %s", err.Error())
				}
				cancel()
				return
			} else {
				go s.handleConnection(ctx, h, c)
			}
		}
	}()

	s.listener = l
	s.log.Info("server started")
	return nil
}

func (s *server) handleConnection(ctx context.Context, handler func(p network.Pipeline) network.Pipeline, c net.Conn) {
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

func (s *server) Shutdown() {
	s.log.Info("shutdown server")
	if err := s.listener.Close(); err != nil {
		s.log.Warn("failed stop listener: ", err.Error())
	}
}
