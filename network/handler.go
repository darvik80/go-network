package network

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"
)

type (
	Message interface{}
	Event   interface{}

	Handler interface{}

	ActiveHandler interface {
		HandleActive(ctx ActiveContext)
	}

	InboundHandler interface {
		HandleRead(ctx InboundContext, message Message)
	}

	OutboundHandler interface {
		HandleWrite(ctx OutboundContext, message Message)
	}

	ErrorHandler interface {
		HandleError(ctx ErrorContext, err error)
	}

	InactiveHandler interface {
		HandleInactive(ctx InactiveContext, err error)
	}

	EventHandler interface {
		HandleEvent(ctx EventContext, event Event)
	}
)

type CodecHandler interface {
	CodecName() string
	InboundHandler
	OutboundHandler
}

type ChannelHandler interface {
	ActiveHandler
	InboundHandler
	OutboundHandler
	ErrorHandler
	InactiveHandler
}

type ChannelInboundHandler interface {
	ActiveHandler
	InboundHandler
	InactiveHandler
}

type ChannelOutboundHandler interface {
	ActiveHandler
	OutboundHandler
	InactiveHandler
}

type ActiveHandlerFunc func(ctx ActiveContext)

func (fn ActiveHandlerFunc) HandleActive(ctx ActiveContext) { fn(ctx) }

type InboundHandlerFunc func(ctx InboundContext, message Message)

func (fn InboundHandlerFunc) HandleRead(ctx InboundContext, message Message) { fn(ctx, message) }

type OutboundHandlerFunc func(ctx OutboundContext, message Message)

func (fn OutboundHandlerFunc) HandleWrite(ctx OutboundContext, message Message) { fn(ctx, message) }

type ErrorHandlerFunc func(ctx ErrorContext, err error)

func (fn ErrorHandlerFunc) HandleError(ctx ErrorContext, err error) { fn(ctx, err) }

type InactiveHandlerFunc func(ctx InactiveContext, err error)

func (fn InactiveHandlerFunc) HandleInactive(ctx InactiveContext, err error) { fn(ctx, err) }

type EventHandlerFunc func(ctx EventContext, event Event)

func (fn EventHandlerFunc) HandleEvent(ctx EventContext, event Event) { fn(ctx, event) }

type headHandler struct{}

func (*headHandler) HandleActive(ctx ActiveContext) {
	ctx.HandleActive()
}

func (*headHandler) HandleRead(ctx InboundContext, message Message) {
	ctx.HandleRead(message)
}

func (*headHandler) HandleWrite(ctx OutboundContext, message Message) {

	switch m := message.(type) {
	case []byte:
		_ = ctx.Channel().WriteData(m)
	case io.Reader:
		data, _ := ioutil.ReadAll(m)
		_ = ctx.Channel().WriteData(data)
	default:
		panic(fmt.Errorf("unsupported type: %T", m))
	}
}

func (*headHandler) HandleError(ctx ErrorContext, err error) {
	ctx.HandleError(err)
}

func (*headHandler) HandleInactive(ctx InactiveContext, err error) {
	_ = ctx.Channel().Close()
}

type tailHandler struct{}

func (*tailHandler) HandleError(ctx ErrorContext, err error) {
	ctx.Close(err)
}

func (*tailHandler) HandleInactive(ctx InactiveContext, err error) {
	ctx.HandleInactive(err)
}

func (*tailHandler) HandleWrite(ctx OutboundContext, message Message) {
	ctx.HandleWrite(message)
}

type (
	ReadIdleEvent struct{}

	WriteIdleEvent struct{}

	IdleEvent struct{}
)

func ReadIdleHandler(idleTime time.Duration) ChannelInboundHandler {
	return &readIdleHandler{
		idleTime: idleTime,
	}
}

func WriteIdleHandler(idleTime time.Duration) ChannelOutboundHandler {
	return &writeIdleHandler{
		idleTime: idleTime,
	}
}

func IdleHandler(idleTime time.Duration) ChannelHandler {
	return &idleHandler{
		idleTime: idleTime,
	}
}

type readIdleHandler struct {
	mutex        sync.RWMutex
	idleTime     time.Duration
	lastReadTime time.Time
	readTimer    *time.Timer
	handlerCtx   HandlerContext
}

func (r *readIdleHandler) withLock(fn func()) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	fn()
}

func (r *readIdleHandler) withReadLock(fn func()) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	fn()
}

func (r *readIdleHandler) HandleActive(ctx ActiveContext) {
	r.withLock(func() {
		r.handlerCtx = ctx
		r.lastReadTime = time.Now()
		r.readTimer = time.AfterFunc(r.idleTime, r.onReadTimeout)
	})

	ctx.HandleActive()
}

func (r *readIdleHandler) HandleRead(ctx InboundContext, message Message) {
	ctx.HandleRead(message)

	r.withLock(func() {
		r.lastReadTime = time.Now()
		if r.readTimer != nil {
			r.readTimer.Reset(r.idleTime)
		}
	})
}

func (r *readIdleHandler) HandleInactive(ctx InactiveContext, err error) {

	r.withLock(func() {
		r.handlerCtx = nil
		if r.readTimer != nil {
			r.readTimer.Stop()
			r.readTimer = nil
		}
	})

	ctx.HandleInactive(err)
}

func (r *readIdleHandler) onReadTimeout() {

	var expired bool
	var ctx HandlerContext

	r.withReadLock(func() {
		expired = time.Since(r.lastReadTime) >= r.idleTime
		ctx = r.handlerCtx
	})

	if expired && ctx != nil {
		func() {
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelError(err.(error))
				}
			}()

			ctx.Trigger(ReadIdleEvent{})
		}()
	}

	r.withReadLock(func() {
		if r.readTimer != nil {
			r.readTimer.Reset(r.idleTime)
		}
	})
}

type writeIdleHandler struct {
	mutex         sync.RWMutex
	idleTime      time.Duration
	lastWriteTime time.Time
	writeTimer    *time.Timer
	handlerCtx    HandlerContext
}

func (w *writeIdleHandler) withLock(fn func()) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	fn()
}

func (w *writeIdleHandler) withReadLock(fn func()) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	fn()
}

func (w *writeIdleHandler) HandleActive(ctx ActiveContext) {
	w.withLock(func() {
		w.handlerCtx = ctx
		w.lastWriteTime = time.Now()
		w.writeTimer = time.AfterFunc(w.idleTime, w.onWriteTimeout)
	})

	ctx.HandleActive()
}

func (w *writeIdleHandler) HandleWrite(ctx OutboundContext, message Message) {
	w.withLock(func() {
		w.lastWriteTime = time.Now()
		if w.writeTimer != nil {
			w.writeTimer.Reset(w.idleTime)
		}
	})

	ctx.HandleWrite(message)
}

func (w *writeIdleHandler) HandleInactive(ctx InactiveContext, err error) {

	w.withLock(func() {
		w.handlerCtx = nil
		if w.writeTimer != nil {
			w.writeTimer.Stop()
			w.writeTimer = nil
		}
	})

	ctx.HandleInactive(err)
}

func (w *writeIdleHandler) onWriteTimeout() {

	var expired bool
	var ctx HandlerContext

	w.withReadLock(func() {
		expired = time.Since(w.lastWriteTime) >= w.idleTime
		ctx = w.handlerCtx
	})

	if expired && ctx != nil {
		func() {
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelError(err.(error))
				}
			}()

			ctx.Trigger(WriteIdleEvent{})
		}()
	}

	w.withReadLock(func() {
		if w.writeTimer != nil {
			w.writeTimer.Reset(w.idleTime)
		}
	})
}

type idleHandler struct {
	mutex         sync.RWMutex
	idleTime      time.Duration
	lastTime time.Time
	timer    *time.Timer
	handlerCtx    HandlerContext
}

func (i *idleHandler) withLock(fn func()) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	fn()
}

func (i *idleHandler) withReadLock(fn func()) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	fn()
}

func (i *idleHandler) HandleActive(ctx ActiveContext) {
	i.withLock(func() {
		i.handlerCtx = ctx
		i.lastTime = time.Now()
		i.timer = time.AfterFunc(i.idleTime, i.onTimeout)
	})

	ctx.HandleActive()
}

func (i *idleHandler) HandleWrite(ctx OutboundContext, message Message) {
	i.withLock(func() {
		i.lastTime = time.Now()
		if i.timer != nil {
			i.timer.Reset(i.idleTime)
		}
	})

	ctx.HandleWrite(message)
}

func (i *idleHandler) HandleRead(ctx InboundContext, message Message) {
	i.withLock(func() {
		i.lastTime = time.Now()
		if i.timer != nil {
			i.timer.Reset(i.idleTime)
		}
	})

	ctx.HandleRead(message)
}

func (i *idleHandler) HandleError(ctx ErrorContext, err error) {
	ctx.HandleError(err)
}

func (i *idleHandler) HandleInactive(ctx InactiveContext, err error) {

	i.withLock(func() {
		i.handlerCtx = nil
		if i.timer != nil {
			i.timer.Stop()
			i.timer = nil
		}
	})

	ctx.HandleInactive(err)
}

func (i *idleHandler) onTimeout() {
	var expired bool
	var ctx HandlerContext

	i.withReadLock(func() {
		expired = time.Since(i.lastTime) >= i.idleTime
		ctx = i.handlerCtx
	})

	if expired && ctx != nil {
		func() {
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelError(err.(error))
				}
			}()

			ctx.Trigger(IdleEvent{})
		}()
	}

	i.withReadLock(func() {
		if i.timer != nil {
			i.timer.Reset(i.idleTime)
		}
	})
}
