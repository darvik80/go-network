package exchange

type Handler interface{}

func adapter(h Handler) Handler {
	switch item := h.(type) {
	case func(source Source, report DwsReport):
		return HandlerDwsReportFunc(item)
	case func(source Source, report SortReport):
		return HandlerSortReportFunc(item)
	default:
		return h
	}
}

type Source interface {}

type Report interface {}

type DwsReport struct {
	Id     int16
	Width  float64
	Weight float64
	Height float64
	Length float64
}

type HandlerDwsReport interface {
	OnMessage(source Source, report DwsReport)
}

type HandlerDwsReportFunc func(source Source, report DwsReport)

func (fn HandlerDwsReportFunc) OnMessage(source Source, report DwsReport) { fn(source, report) }

type SortReport struct {
	Id      int
	ChuteId int
}

type HandlerSortReport interface {
	OnMessage(source Source, report SortReport)
}

type HandlerSortReportFunc func(source Source, report SortReport)

func (fn HandlerSortReportFunc) OnMessage(source Source, report SortReport) { fn(source, report) }
