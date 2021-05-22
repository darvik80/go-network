package net

import (
	"context"
	"net"
	"sync/atomic"
)

type Channel interface {
	Close() error

	IsActive() bool

	Write(Message) bool

	WriteData([]byte) error

	Context() context.Context
	Pipeline() Pipeline

	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}

func NewChannelWith(ctx context.Context, pipeline Pipeline, transport net.Conn) Channel {
	childCtx, cancel := context.WithCancel(ctx)
	ch :=  &channel{
		ctx:       childCtx,
		cancel:    cancel,
		pipeline:  pipeline,
		transport: transport,
	}

	pipeline.AttachChannel(ch)
	defer pipeline.FireChannelActive()

	return ch
}

type channel struct {
	ctx        context.Context
	cancel     context.CancelFunc
	transport  net.Conn
	pipeline   Pipeline
	closed     int32
}

func (c *channel) Write(message Message) bool {

	select {
	case <-c.ctx.Done():
		return false
	default:
		c.pipeline.FireChannelWrite(message)
		return true
	}
}

func (c *channel) WriteData(data []byte) error {
	_, err := c.transport.Write(data)
	return err
}

func (c *channel) RemoteAddr() net.Addr {
	return c.transport.RemoteAddr()
}

func (c *channel) LocalAddr() net.Addr {
	return c.transport.LocalAddr()
}

func (c *channel) Close() error {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.cancel()
		return c.transport.Close()
	}
	return nil
}

func (c *channel) IsActive() bool {
	return 0 == atomic.LoadInt32(&c.closed)
}

func (c *channel) Transport() net.Conn {
	return c.transport
}

func (c *channel) Pipeline() Pipeline {
	return c.pipeline
}

func (c *channel) Context() context.Context {
	return c.ctx
}

func (c *channel) invokeMethod(fn func()) {

	defer func() {
		if err := recover(); nil != err && 0 == atomic.LoadInt32(&c.closed) {
			c.pipeline.FireChannelError(err.(error))
		}
	}()

	fn()
}

func (c *channel) postCloseEvent(err error) {
	if 0 == atomic.LoadInt32(&c.closed) {
		c.pipeline.FireChannelInactive(err)
	}
}
