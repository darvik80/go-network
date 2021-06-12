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
		//key := deviceConfig.Address + ":" + strconv.Itoa(deviceConfig.Port)
		//if deviceConfig.Mode == "SERVER" {
		//	if _, found := mid.servers[key]; !found {
		//		log.Info("[mid] create server: ", deviceConfig.Address, ":", deviceConfig.Port)
		//		server := tcp.NewServer(deviceConfig.Address, deviceConfig.Port)
		//		mid.servers[key] = server
		//	}
		//} else if deviceConfig.Mode == "CLIENT" {
		//	if _, found := mid.servers[key]; !found {
		//		log.Info("[mid] create client: ", deviceConfig.Address, ":", deviceConfig.Port)
		//		client := tcp.NewClient(deviceConfig.Address, deviceConfig.Port)
		//		mid.servers[key] = client
		//	}
		//}
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
	//for _, server := range mid.servers {
	//	server.Shutdown()
	//}
	//
	//for _, client := range mid.clients {
	//	client.Shutdown()
	//}
}
