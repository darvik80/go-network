package exchange

type Exchange interface {
	Publish(source Source, msg Message)
	Subscribe(handler Handler)
	Shutdown()
}
