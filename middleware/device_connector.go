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
	exchange exchange.Exchange
	links    map[string]*info
}

func NewServerConnector(exchange exchange.Exchange) DeviceConnector {
	return &serverDeviceConnector{
		exchange: exchange,
		links:    make(map[string]*info),
	}
}

func (s *serverDeviceConnector) Connect(device Device) bool {
	if rec, found := s.links[device.Address()]; !found {
		rec = &info{
			channel: tcp.NewServer(device.Address(), codec.NewPipelineFactory(device.Codec())),
			devices: make(map[Device]bool),
		}
		rec.devices[device] = true
		s.links[device.Address()] = rec
	} else {
		rec.devices[device] = true
	}

	return true
}

func (s *serverDeviceConnector) Start() error {
	var ex = s.exchange
	for _, ch := range s.links {
		info := ch
		fn := network.HandlerFactoryFunc(func(ctx context.Context, conn net.Conn) network.Handler {

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
					logger.Warn(device.Name(), " active: ", conn.RemoteAddr())
					device.SetChannel(ctx.Channel())
					ctx.HandleActive()
				},
				func(ctx network.InboundContext, msg network.Message) {
					if device != nil {
						switch m := msg.(type) {
						case exchange.SortReport:
							logger.Info(device.Name(), " Sort, Id: ", m.Id, ", ChuteId: ", m.ChuteId)
						case exchange.DwsReport:
							logger.Info(device.Name(), " DWS Id: ", m.Id)
						}
						(device.(exchange.Exchange)).Send(device, msg)
						ex.Send(device, msg)
					}
				},
				func(ctx network.InactiveContext, err error) {
					logger.Warn(device.Name(), " inactive: ", conn.RemoteAddr())
					device.SetChannel(nil)
					ctx.HandleInactive(err)
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
	exchange exchange.Exchange
	links    map[string]*info
}

func NewClientConnector(exchange exchange.Exchange) DeviceConnector {
	return &clientDeviceConnector{
		exchange: exchange,
		links:    make(map[string]*info),
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
	ex := s.exchange
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

			return network.NewChannelInboundHandlerFunc(
				func(ctx network.ActiveContext) {
					logger.Info(device.Name(), " active: ", conn.RemoteAddr())
					device.SetChannel(ctx.Channel())
					ctx.HandleActive()
				},
				func(ctx network.InboundContext, msg network.Message) {
					if device != nil {
						switch m := msg.(type) {
						case exchange.SortReport:
							logger.Infof("%s, Sort, Id: %d, ChuteId: %d", device.Name(), m.Id, m.ChuteId)
						case exchange.DwsReport:
							logger.Infof("%s, DWS Id: %d", device.Name(), m.Id)
						}
						ex.Send(device, msg)
					}
				},
				func(ctx network.InactiveContext, err error) {
					logger.Info(device.Name(), " inactive: ", conn.RemoteAddr())
					device.SetChannel(nil)
					ctx.HandleInactive(err)
				},
			)
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
