package net

import (
	"io"
	"io/ioutil"
	"sync"
	"time"
)

type (
	Message interface{}
	Event   interface{}

	Handler interface {}

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

type SimpleChannelHandler = ChannelInboundHandler

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
		//panic(fmt.Errorf("unsupported type: %T", m))
	}
}

func (*headHandler) HandleError(ctx ErrorContext, err error) {
	ctx.HandleError(err)
}

func (*headHandler) HandleInactive(ctx InactiveContext, err error) {
	_ = ctx.Channel().Close()
}

// default: tailHandler
// The final closing operation will be provided when the user registered handler is not processing.
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
	// ReadIdleEvent define a ReadIdleEvent
	ReadIdleEvent struct{}

	// WriteIdleEvent define a WriteIdleEvent
	WriteIdleEvent struct{}
)

// ReadIdleHandler fire ReadIdleEvent after waiting for a reading timeout
func ReadIdleHandler(idleTime time.Duration) ChannelInboundHandler {
	return &readIdleHandler{
		idleTime: idleTime,
	}
}

// WriteIdleHandler fire WriteIdleEvent after waiting for a sending timeout
func WriteIdleHandler(idleTime time.Duration) ChannelOutboundHandler {
	//utils.AssertIf(idleTime < time.Second, "idleTime must be greater than one second")
	return &writeIdleHandler{
		idleTime: idleTime,
	}
}

// readIdleHandler
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
	// cache context.
	r.withLock(func() {
		r.handlerCtx = ctx
		r.lastReadTime = time.Now()
		r.readTimer = time.AfterFunc(r.idleTime, r.onReadTimeout)
	})
	// post the active event.
	ctx.HandleActive()
}

func (r *readIdleHandler) HandleRead(ctx InboundContext, message Message) {
	ctx.HandleRead(message)

	r.withLock(func() {
		// update last read time.
		r.lastReadTime = time.Now()
		// reset timer.
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

	// post the inactive event.
	ctx.HandleInactive(err)
}

func (r *readIdleHandler) onReadTimeout() {

	var expired bool
	var ctx HandlerContext

	r.withReadLock(func() {
		// check if the idle time expires.
		expired = time.Since(r.lastReadTime) >= r.idleTime
		ctx = r.handlerCtx
	})

	if expired && ctx != nil {
		// trigger event.
		func() {
			// capture exception.
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelError(err.(error))
				}
			}()

			// trigger ReadIdleEvent.
			ctx.Trigger(ReadIdleEvent{})
		}()
	}

	// reset timer
	r.withReadLock(func() {
		if r.readTimer != nil {
			r.readTimer.Reset(r.idleTime)
		}
	})
}

// writeIdleHandler
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

	// cache context
	w.withLock(func() {
		w.handlerCtx = ctx
		w.lastWriteTime = time.Now()
		w.writeTimer = time.AfterFunc(w.idleTime, w.onWriteTimeout)
	})

	// post the active event.
	ctx.HandleActive()
}

func (w *writeIdleHandler) HandleWrite(ctx OutboundContext, message Message) {

	// update last write time.
	w.withLock(func() {
		w.lastWriteTime = time.Now()
		// reset timer.
		if w.writeTimer != nil {
			w.writeTimer.Reset(w.idleTime)
		}
	})

	// post write event.
	ctx.HandleWrite(message)
}

func (w *writeIdleHandler) HandleInactive(ctx InactiveContext, err error) {

	w.withLock(func() {
		// reset context
		w.handlerCtx = nil
		// stop the timer.
		if w.writeTimer != nil {
			w.writeTimer.Stop()
			w.writeTimer = nil
		}
	})

	// post the inactive event.
	ctx.HandleInactive(err)
}

func (w *writeIdleHandler) onWriteTimeout() {

	var expired bool
	var ctx HandlerContext

	w.withReadLock(func() {
		// check if the idle time expires.
		expired = time.Since(w.lastWriteTime) >= w.idleTime
		ctx = w.handlerCtx
	})

	// check if the idle time expires
	if expired && ctx != nil {
		// trigger event.
		func() {
			// capture exception
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelError(err.(error))
				}
			}()

			// trigger WriteIdleEvent.
			ctx.Trigger(WriteIdleEvent{})
		}()
	}

	// reset timer.
	w.withReadLock(func() {
		if w.writeTimer != nil {
			w.writeTimer.Reset(w.idleTime)
		}
	})

}