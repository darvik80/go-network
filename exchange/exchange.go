package exchange

type Exchange interface {
	Send(report Report)
	Subscribe(handler Handler)
}
