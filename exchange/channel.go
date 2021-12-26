package exchange

import "container/list"

type message struct {
	source Source
	msg    Message
}

type channel struct {
	ch       chan message
	handlers *list.List
}

func NewChanExchange(bufSize, routines int) *channel {
	ex := &channel{
		ch:       make(chan message, bufSize),
		handlers: list.New(),
	}

	if routines <= 0 {
		routines = 1
	}
	for idx := 0; idx < routines; idx++ {
		go ex.process()
	}

	return ex
}

func (e *channel) process() {
	for {
		select {
		case msg, ok := <-e.ch:
			if ok {
				switch m := msg.msg.(type) {
				case StdDwsReport:
					for h := e.handlers.Front(); h != nil; h = h.Next() {
						switch handler := h.Value.(type) {
						case StdHandlerDwsReport:
							handler.OnMessage(msg.source, m)
						}
					}
				case StdDwsSortReport:
					for h := e.handlers.Front(); h != nil; h = h.Next() {
						switch handler := h.Value.(type) {
						case StdHandlerDwsSortReport:
							handler.OnMessage(msg.source, m)
						}
					}
				case StdHeartbeat:
					for h := e.handlers.Front(); h != nil; h = h.Next() {
						switch handler := h.Value.(type) {
						case StdHandlerHeartbeat:
							handler.OnMessage(msg.source, m)
						}
					}
				case StdKeepAliveRequest:
					for h := e.handlers.Front(); h != nil; h = h.Next() {
						switch handler := h.Value.(type) {
						case StdHandlerKeepAliveRequest:
							handler.OnMessage(msg.source, m)
						}
					}
				case StdKeepAliveResponse:
					for h := e.handlers.Front(); h != nil; h = h.Next() {
						switch handler := h.Value.(type) {
						case StdHandlerKeepAliveResponse:
							handler.OnMessage(msg.source, m)
						}
					}
				}
			}
		}
	}
}

func (e *channel) Publish(source Source, msg Message) {
	e.ch <- message{source, msg}
}

func (e *channel) Subscribe(handler Handler) {
	e.handlers.PushBack(adapter(handler))
}

func (e *channel) Shutdown() {
	close(e.ch)
}
