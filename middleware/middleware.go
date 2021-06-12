package middleware

import "darvik80/go-network/exchange"

type middleware struct {
	exchange        exchange.Exchange
	links           []Link
	serverConnector DeviceConnector
	clientConnector DeviceConnector
}

func NewMiddleware(config []LinkConfig) (*middleware, error) {
	ex := exchange.NewChanExchange(1024, 8)
	mid := &middleware{
		exchange:        ex,
		serverConnector: NewServerConnector(ex),
		clientConnector: NewClientConnector(ex),
	}

	for _, linkConfig := range config {
		var devices []Device
		linkEx := exchange.NewChanExchange(1024, 8)
		for _, deviceConfig := range linkConfig.Devices {
			device := NewDevice(deviceConfig, linkEx)
			switch device.Mode() {
			case SERVER:
				mid.serverConnector.Connect(device)
			case CLIENT:
				mid.clientConnector.Connect(device)
			}
			devices = append(devices, device)
		}

		mid.links = append(mid.links, NewLink(devices...))
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
