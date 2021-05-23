package network

type (
	HandlerContext interface {
		Pipeline() Pipeline
		Channel() Channel
		Handler() Handler
		Write(message Message)
		Close(err error)
		Trigger(event Event)
	}

	ActiveContext interface {
		HandlerContext
		HandleActive()
	}

	InboundContext interface {
		HandlerContext
		HandleRead(message Message)
	}

	OutboundContext interface {
		HandlerContext
		HandleWrite(message Message)
	}

	ErrorContext interface {
		HandlerContext
		HandleError(err error)
	}

	InactiveContext interface {
		HandlerContext
		HandleInactive(err error)
	}

	EventContext interface {
		HandlerContext
		HandleEvent(event Event)
	}
)

type handlerContext struct {
	pipeline Pipeline
	handler  Handler
	prev     *handlerContext
	next     *handlerContext
}

func (hc *handlerContext) prevContext() *handlerContext {
	return hc.prev
}

func (hc *handlerContext) nextContext() *handlerContext {
	return hc.next
}

func (hc *handlerContext) Write(message Message) {
	var next = hc

	for {
		if next = next.prevContext(); nil == next {
			break
		}

		if handler, ok := next.Handler().(OutboundHandler); ok {
			handler.HandleWrite(next, message)
			break
		}
	}
}

func (hc *handlerContext) Close(err error) {

	var prev = hc
	for {
		if prev = prev.prevContext(); nil == prev {
			break
		}

		if handler, ok := prev.Handler().(InactiveHandler); ok {
			handler.HandleInactive(prev, err)
			break
		}
	}
}

func (hc *handlerContext) Trigger(event Event) {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler, ok := next.Handler().(EventHandler); ok {
			handler.HandleEvent(next, event)
			break
		}
	}
}

func (hc *handlerContext) Channel() Channel {
	return hc.pipeline.Channel()
}

func (hc *handlerContext) Pipeline() Pipeline {
	return hc.pipeline
}

func (hc *handlerContext) Handler() Handler {
	return hc.handler
}

func (hc *handlerContext) HandleActive() {

	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler, ok := next.Handler().(ActiveHandler); ok {
			handler.HandleActive(next)
			break
		}
	}
}

func (hc *handlerContext) HandleRead(message Message) {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler, ok := next.Handler().(InboundHandler); ok {
			handler.HandleRead(next, message)
			break
		}
	}
}

func (hc *handlerContext) HandleWrite(message Message) {
	var prev = hc

	for {
		if prev = prev.prevContext(); nil == prev {
			break
		}

		if handler, ok := prev.Handler().(OutboundHandler); ok {
			handler.HandleWrite(prev, message)
			break
		}
	}
}

func (hc *handlerContext) HandleError(err error) {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler, ok := next.Handler().(ErrorHandler); ok {
			handler.HandleError(next, err)
			break
		}
	}
}

func (hc *handlerContext) HandleInactive(err error) {
	var next = hc

	for {
		if next = next.prevContext(); nil == next {
			break
		}

		if handler, ok := next.Handler().(InactiveHandler); ok {
			handler.HandleInactive(next, err)
			break
		}
	}
}

func (hc *handlerContext) HandleEvent(event Event) {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler, ok := next.Handler().(EventHandler); ok {
			handler.HandleEvent(next, event)
			break
		}
	}
}
