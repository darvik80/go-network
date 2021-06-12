package network

type Transport interface {
	Start(factory HandlerFactory) error
	Shutdown()
}
