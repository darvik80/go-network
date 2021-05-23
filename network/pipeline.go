package network

import (
	"fmt"
)

type Pipeline interface {
	AttachChannel(channel Channel)

	AddFirst(handlers ...Handler) Pipeline

	AddLast(handlers ...Handler) Pipeline

	FireChannelActive()
	FireChannelRead(message Message)
	FireChannelWrite(message Message)
	FireChannelError(err error)
	FireChannelInactive(err error)
	FireChannelEvent(event Event)
	Channel() Channel
}

func NewPipeline() Pipeline {

	p := &pipeline{}

	p.head = &handlerContext{
		pipeline: p,
		handler:  new(headHandler),
	}

	p.tail = &handlerContext{
		pipeline: p,
		handler:  new(tailHandler),
	}

	p.head.next = p.tail
	p.tail.prev = p.head

	return p
}

type pipeline struct {
	head    *handlerContext
	tail    *handlerContext
	channel Channel
}

func (p* pipeline) adapter(h Handler) Handler {
	switch item := h.(type) {
	case func(ctx ActiveContext):
		return ActiveHandlerFunc(item)
	case func(ctx InactiveContext, err error):
		return InactiveHandlerFunc(item)
	case func(ctx OutboundContext, message Message):
		return OutboundHandlerFunc(item)
	case func(ctx InboundContext, message Message):
		return InboundHandlerFunc(item)
	case func(ctx ErrorContext, err error):
		return ErrorHandlerFunc(item)
	case func(ctx EventContext, event Event):
		return EventHandlerFunc(item)
	default:
		return h
	}
}

func (p *pipeline) AddFirst(handlers ...Handler) Pipeline {
	for _, h := range handlers {
		p.addFirst(p.adapter(h))
	}

	return p
}

func (p *pipeline) AddLast(handlers ...Handler) Pipeline {
	for _, h := range handlers {
		p.addLast(p.adapter(h))
	}
	return p
}

func (p *pipeline) addFirst(handler Handler) {

	// checking handler.
	checkHandler(handler)

	oldNext := p.head.next
	p.head.next = &handlerContext{
		pipeline: p,
		handler:  handler,
		prev:     p.head,
		next:     oldNext,
	}

	oldNext.prev = p.head.next
}

func (p *pipeline) addLast(handler Handler) {
	checkHandler(handler)

	oldPrev := p.tail.prev
	p.tail.prev = &handlerContext{
		pipeline: p,
		handler:  handler,
		prev:     oldPrev,
		next:     p.tail,
	}

	oldPrev.next = p.tail.prev
}

func (p *pipeline) Channel() Channel {
	return p.channel
}

func (p *pipeline) AttachChannel(channel Channel) {
	p.channel = channel
}

func (p *pipeline) FireChannelActive() {
	p.head.HandleActive()
}

func (p *pipeline) FireChannelRead(message Message) {
	p.head.HandleRead(message)
}

func (p *pipeline) FireChannelWrite(message Message) {
	p.tail.HandleWrite(message)
}

func (p *pipeline) FireChannelError(err error) {
	p.head.HandleError(err)
}

func (p *pipeline) FireChannelInactive(err error) {
	p.tail.HandleInactive(err)
}

func (p *pipeline) FireChannelEvent(event Event) {
	p.head.HandleEvent(event)
}

func checkHandler(handlers ...Handler) {
	for _, h := range handlers {
		switch h.(type) {
		case ActiveHandler:
		case InboundHandler:
		case OutboundHandler:
		case ErrorHandler:
		case InactiveHandler:
		case EventHandler:
		default:
			panic(fmt.Errorf("unrecognized Handler: %T", h))
		}
	}
}
