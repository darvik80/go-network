package middleware

type middleware struct {
	links []Link
	serverConnector DeviceConnector
	clientConnector DeviceConnector
}

func NewMiddleware(config []DeviceConfig) (*middleware, error) {
	mid := &middleware{
		serverConnector : NewServerConnector(),
		clientConnector: NewClientConnector(),
	}

	for _, deviceConfig := range config {
		device := NewDevice(deviceConfig)
		switch device.Mode() {
		case SERVER:
			mid.serverConnector.Connect(device)
		case CLIENT:
			mid.clientConnector.Connect(device)
		}
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
