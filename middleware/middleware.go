package middleware

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/middleware/route"
)

type middleware struct {
	exchange        exchange.Exchange
	machines        []Machine
	serverConnector DeviceConnector
	clientConnector DeviceConnector
}

type Options struct {
	router   route.SortingRouter
	links    []LinkConfig
	devices  []DeviceConfig
	exchange exchange.Exchange
	handlers []exchange.Handler
}

type Option func(*Options)

func WithRouter(router route.SortingRouter) Option {
	return func(o *Options) {
		o.router = router
	}
}

func WithLinks(links []LinkConfig) Option {
	return func(o *Options) {
		o.links = links
	}
}

func WithDevices(devices []DeviceConfig) Option {
	return func(o *Options) {
		o.devices = devices
	}
}

func WithExchange(ex exchange.Exchange) Option {
	return func(o *Options) {
		o.exchange = ex
	}
}

func WithSubscriber(sub exchange.Handler) Option {
	return func(o *Options) {
		o.handlers = append(o.handlers, sub)
	}
}

func NewMiddleware(opts ...Option) (*middleware, error) {

	var options = &Options{}
	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *House as the argument
		opt(options)
	}

	mid := &middleware{
		exchange:        options.exchange,
		serverConnector: NewServerConnector(options.exchange),
		clientConnector: NewClientConnector(options.exchange),
	}
	for _, sub := range options.handlers {
		mid.exchange.Subscribe(sub)
	}

	for _, linkConfig := range options.links {
		var devices []Device
		linkEx := exchange.NewChanExchange(1024, 8)
		for _, deviceConfig := range linkConfig.Devices {
			device := NewDevice(linkConfig.Id, deviceConfig, linkEx)
			switch device.Mode() {
			case SERVER:
				mid.serverConnector.Connect(device)
			case CLIENT:
				mid.clientConnector.Connect(device)
			}
			devices = append(devices, device)
		}

		mid.machines = append(mid.machines, NewDwsPlcMachine(mid.exchange, devices...))
	}

	for _, deviceConfig := range options.devices {
		linkEx := exchange.NewChanExchange(1024, 8)
		device := NewDevice(0, deviceConfig, linkEx)
		switch device.Mode() {
		case SERVER:
			mid.serverConnector.Connect(device)
		case CLIENT:
			mid.clientConnector.Connect(device)
		}
		mid.machines = append(mid.machines, NewScadaMachine(mid.exchange, options.router, device))
	}

	if err := mid.serverConnector.Start(); err != nil {
		return nil, err
	}

	if err := mid.clientConnector.Start(); err != nil {
		return nil, err
	}

	return mid, nil
}

func (mid *middleware) Shutdown() {
	mid.serverConnector.Shutdown()
	mid.clientConnector.Shutdown()
}
