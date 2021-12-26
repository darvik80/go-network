package exchange

type Handler interface{}

func adapter(h Handler) Handler {
	switch item := h.(type) {
	case func(source Source, report StdDwsReport):
		return StdHandlerDwsReportFunc(item)
	case func(source Source, report StdSortReport):
		return StdHandlerSortReportFunc(item)
	case func(source Source, report StdDwsSortReport):
		return StdHandlerDwsSortReportFunc(item)
	case func(source Source, msg StdKeepAliveRequest):
		return StdHandlerKeepAliveRequestFunc(item)
	case func(source Source, msg StdKeepAliveResponse):
		return StdHandlerKeepAliveResponseFunc(item)
	case func(source Source, msg StdHeartbeat):
		return StdHandlerHeartbeatFunc(item)
	default:
		return h
	}
}

type Source interface {
	Name() string
}

type Message interface{}

type StdHandlerDwsReport interface {
	OnMessage(source Source, report StdDwsReport)
}

type StdHandlerDwsReportFunc func(source Source, report StdDwsReport)

func (fn StdHandlerDwsReportFunc) OnMessage(source Source, report StdDwsReport) { fn(source, report) }

type StdHandlerSortRequest interface {
	OnMessage(source Source, report StdSortRequest)
}

type StdHandlerSortRequestFunc func(source Source, report StdSortRequest)

func (fn StdHandlerSortRequestFunc) OnMessage(source Source, report StdSortRequest) {
	fn(source, report)
}

type StdHandlerSortReport interface {
	OnMessage(source Source, report StdDwsSortReport)
}

type StdHandlerSortReportFunc func(source Source, report StdSortReport)

func (fn StdHandlerSortReportFunc) OnMessage(source Source, report StdSortReport) {
	fn(source, report)
}

type StdHandlerDwsSortReport interface {
	OnMessage(source Source, report StdDwsSortReport)
}

type StdHandlerDwsSortReportFunc func(source Source, report StdDwsSortReport)

func (fn StdHandlerDwsSortReportFunc) OnMessage(source Source, report StdDwsSortReport) {
	fn(source, report)
}

type StdHandlerKeepAliveRequest interface {
	OnMessage(source Source, report StdKeepAliveRequest)
}

type StdHandlerKeepAliveRequestFunc func(source Source, msg StdKeepAliveRequest)

func (fn StdHandlerKeepAliveRequestFunc) OnMessage(source Source, report StdKeepAliveRequest) {
	fn(source, report)
}

type StdHandlerKeepAliveResponse interface {
	OnMessage(source Source, report StdKeepAliveResponse)
}

type StdHandlerKeepAliveResponseFunc func(source Source, msg StdKeepAliveResponse)

func (fn StdHandlerKeepAliveResponseFunc) OnMessage(source Source, msg StdKeepAliveResponse) {
	fn(source, msg)
}

type StdHandlerHeartbeat interface {
	OnMessage(source Source, report StdHeartbeat)
}

type StdHandlerHeartbeatFunc func(source Source, msg StdHeartbeat)

func (fn StdHandlerHeartbeatFunc) OnMessage(source Source, msg StdHeartbeat) {
	fn(source, msg)
}
