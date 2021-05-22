package net

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
)

type server struct {
	host     string
	port     int
	listener net.Listener
	log      *log.Entry
	ctx      context.Context
}

func NewServer(host string, port int) *server {
	return &server{
		host,
		port,
		nil,
		log.WithFields(
			log.Fields{
				"module": "tcp-server",
				"server": host,
				"port":   port,
			}),
		context.Background(),
	}
}

func (s *server) Start(h func(p Pipeline) Pipeline) error {
	l, err := net.Listen("tcp4", s.host+":"+strconv.Itoa(s.port))
	if err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithCancel(s.ctx)
		for {
			c, err := l.Accept()
			if err != nil {
				s.log.Warnf("can't accept connection, %e", err)
				cancel()
				return
			} else {
				s.handleConnection(ctx, h, c)
			}
		}
	}()

	s.listener = l
	s.log.Info("server started")
	return nil
}

func (s *server) handleConnection(ctx context.Context, handler func(p Pipeline) Pipeline, c net.Conn) {
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

	channel := NewChannelWith(context.Background(), handler(NewPipeline()), c)
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

func (s *server) Shutdown() error {
	return s.listener.Close()
}
