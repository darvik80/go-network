package exchange

type Handler interface{}

func adapter(h Handler) Handler {
	switch item := h.(type) {
	case func(report DwsReport):
		return HandlerDwsReportFunc(item)
	case func(report SortReport):
		return HandlerSortReportFunc(item)
	default:
		return h
	}
}

type Report interface {}

type DwsReport struct {
	Id     int16
	Width  float64
	Weight float64
	Height float64
	Length float64
}

type HandlerDwsReport interface {
	OnMessage(report DwsReport)
}

type HandlerDwsReportFunc func(report DwsReport)

func (fn HandlerDwsReportFunc) OnMessage(report DwsReport) { fn(report) }

type SortReport struct {
	Id      int16
	ChuteId int16
}

type HandlerSortReport interface {
	OnMessage(report SortReport)
}

type HandlerSortReportFunc func(report SortReport)

func (fn HandlerSortReportFunc) OnMessage(report SortReport) { fn(report) }
