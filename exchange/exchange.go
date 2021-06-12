package exchange

type Exchange interface {
	Send(source Source, report Report)
	Subscribe(handler Handler)
}
