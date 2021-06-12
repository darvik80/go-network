package middleware

import (
	"darvik80/go-network/network"
)

type DeviceConnector interface {
	Connect(device Device) bool
	Start() error
	Shutdown()
}

type channelInfo struct {
	transport network.Transport
	factory   network.PipelineFactoryFunc
}

type serverDeviceConnector struct {
	channels map[string]*channelInfo
	devices  map[Device]bool
}

func NewServerConnector() DeviceConnector {
	return &serverDeviceConnector{
		channels: make(map[string]*channelInfo),
		devices:  make(map[Device]bool),
	}
}

func (s *serverDeviceConnector) Connect(device Device) bool {
	if _, found := s.channels[device.Address()]; !found {
		s.channels[device.Address()] = &channelInfo{
			nil, //tcp.NewServer(device.Address()),
			nil, //codec.NewPipeline(device.Codec()),
		}
		s.devices[device] = true
	}

	return true
}

func (s *serverDeviceConnector) Start() error {
	//for _, ch := range s.channels {
	//	ch.transport.Start(ch.factory().)
	//}

	return nil
}

func (s *serverDeviceConnector) Shutdown() {

}

type clientDeviceConnector struct {
	channels map[string]network.Transport
	devices  map[string]Device
}

func NewClientConnector() DeviceConnector {
	return &clientDeviceConnector{
		channels: make(map[string]network.Transport),
		devices:  make(map[string]Device),
	}
}

func (s *clientDeviceConnector) Connect(device Device) bool {
	return false
}

func (s *clientDeviceConnector) Start() error {
	return nil
}

func (s *clientDeviceConnector) Shutdown() {

}
