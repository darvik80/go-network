package middleware

import (
	"context"
	"darvik80/go-network/exchange"
	"darvik80/go-network/middleware/codec"
	"darvik80/go-network/network"
	"darvik80/go-network/network/tcp"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
)

var logger = log.WithFields(log.Fields{"module": "server-connector"})

type DeviceConnector interface {
	Connect(device Device) bool
	Start() error
	Shutdown()
}

type info struct {
	channel network.Transport
	devices map[Device]bool
}

type serverDeviceConnector struct {
	mu sync.Mutex
	links map[string]*info
}

func NewServerConnector() DeviceConnector {
	return &serverDeviceConnector{
		links: make(map[string]*info),
	}
}

func (s *serverDeviceConnector) Connect(device Device) bool {
	if rec, found := s.links[device.Address()]; !found {
		rec = &info{
			channel: tcp.NewServer(device.Address(), codec.NewPipelineFactory(device.Codec())),
			devices: make(map[Device]bool),
		}
		rec.devices[device] = false
		s.links[device.Address()] = rec
	} else {
		rec.devices[device] = true
	}

	return true
}

func (s *serverDeviceConnector) Start() error {
	var mutex = &s.mu
	for _, ch := range s.links {
		info := ch
		fn := network.HandlerFactoryFunc(func(ctx context.Context, conn net.Conn) network.Handler {
			mutex.Lock()
			defer mutex.Unlock()

			var device Device
			for dev := range info.devices {
				if nil == dev.AllowedAddresses() {
					device = dev
					break
				} else {
					raddr := conn.RemoteAddr().String()
					for _, addr := range dev.AllowedAddresses() {
						if strings.HasPrefix(raddr, addr) {
							device = dev
							break
						}
					}
				}
			}

			if device == nil {
				logger.Warn("not allowed: ", conn.RemoteAddr())
				return nil
			}

			return network.NewChannelInboundHandlerFunc(
				func(ctx network.ActiveContext) {
					mutex.Lock()
					defer mutex.Unlock()

					if active, found := info.devices[device]; found && active {
						logger.Warn(device.Name(), " already connected: ", conn.RemoteAddr())
						ctx.Close(nil)
					} else {
						info.devices[device] = true
						logger.Warn(device.Name(), " active: ", conn.RemoteAddr())
					}
				},
				func(ctx network.InboundContext, msg network.Message) {
					if device != nil {
						switch m := msg.(type) {
						case exchange.SortReport:
							logger.Info(device.Name(), " Sort, Id: ", m.Id, ", ChuteId: ", m.ChuteId)
						case exchange.DwsReport:
							logger.Info(device.Name(), " DWS Id: ", m.Id)
						}
					}
				},
				func(ctx network.InactiveContext, err error) {
					mutex.Lock()
					defer mutex.Unlock()

					info.devices[device] = false
					logger.Warn(device.Name(), " inactive: ", conn.RemoteAddr())
				},
			)
		})

		if err := ch.channel.Start(fn); err != nil {
			return err
		}
	}

	return nil
}

func (s *serverDeviceConnector) Shutdown() {
	for _, ch := range s.links {
		ch.channel.Shutdown()
	}
}

type clientDeviceConnector struct {
	links map[string]*info
}

func NewClientConnector() DeviceConnector {
	return &clientDeviceConnector{
		links: make(map[string]*info),
	}
}

func (s *clientDeviceConnector) Connect(device Device) bool {
	if rec, found := s.links[device.Address()]; !found {
		rec = &info{
			channel: tcp.NewClient(device.Address(), codec.NewPipelineFactory(device.Codec())),
			devices: make(map[Device]bool),
		}
		rec.devices[device] = true
		s.links[device.Address()] = rec
	} else {
		rec.devices[device] = true
	}

	return true
}

func (s *clientDeviceConnector) Start() error {
	for _, ch := range s.links {
		info := ch
		fn := network.HandlerFactoryFunc(func(ctx context.Context, conn net.Conn) network.Handler {
			var device Device
			for dev, _ := range info.devices {
				device = dev
				break
			}

			if device == nil {
				logger.Warn("not allowed connection: ", conn.RemoteAddr())
				return nil
			}

			return func(ctx network.InboundContext, msg network.Message) {
				if device != nil {
					switch m := msg.(type) {
					case exchange.SortReport:
						logger.Infof("%s, Sort, Id: %d, ChuteId: %d", device.Name(), m.Id, m.ChuteId)
					case exchange.DwsReport:
						logger.Infof("%s, DWS Id: %d", device.Name(), m.Id)
					}
				}
			}
		})

		if err := ch.channel.Start(fn); err != nil {
			return err
		}
	}

	return nil
}

func (s *clientDeviceConnector) Shutdown() {
	for _, ch := range s.links {
		ch.channel.Shutdown()
	}
}
