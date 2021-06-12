package exchange

type Exchange interface {
	Publish(source Source, report Report)
	Subscribe(handler Handler)
}
