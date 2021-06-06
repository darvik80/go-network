package exchange

import "container/list"

type channel struct {
	ch       chan interface{}
	handlers *list.List
}

func NewChanExchange() *channel {
	ex := &channel{
		ch:       make(chan interface{}, 128),
		handlers: list.New(),
	}

	go ex.process()

	return ex
}

func (e *channel) process() {
	for {
		select {
		case msg, ok := <-e.ch:
			if !ok {
				return
			}
			switch m := msg.(type) {
			case DwsReport:
				for h := e.handlers.Front(); h != nil; h = h.Next() {
					switch handler := h.Value.(type) {
					case HandlerDwsReport:
						handler.OnMessage(m)
					}
				}
			case SortReport:
				for h := e.handlers.Front(); h != nil; h = h.Next() {
					switch handler := h.Value.(type) {
					case HandlerSortReport:
						handler.OnMessage(m)
					}
				}
			}
		}
	}
}

func (e *channel) Send(msg interface{}) {
	e.ch <- msg
}

func (e *channel) Subscribe(handler interface{}) {
	e.handlers.PushBack(adapter(handler))
}

func (e *channel) Shutdown() {
	close(e.ch)
}