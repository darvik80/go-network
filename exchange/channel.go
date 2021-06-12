package exchange

import "container/list"

type message struct {
	source Source
	report Report
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
				switch m := msg.report.(type) {
				case DwsReport:
					for h := e.handlers.Front(); h != nil; h = h.Next() {
						switch handler := h.Value.(type) {
						case HandlerDwsReport:
							handler.OnMessage(msg.source, m)
						}
					}
				case SortReport:
					for h := e.handlers.Front(); h != nil; h = h.Next() {
						switch handler := h.Value.(type) {
						case HandlerSortReport:
							handler.OnMessage(msg.source, m)
						}
					}
				}
			}
		}
	}
}

func (e *channel) Publish(source Source, msg Report) {
	e.ch <- message{ source, msg }
}

func (e *channel) Subscribe(handler Handler) {
	e.handlers.PushBack(adapter(handler))
}

func (e *channel) Shutdown() {
	close(e.ch)
}
