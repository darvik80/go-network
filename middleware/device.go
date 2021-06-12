package middleware

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/middleware/codec"
	"darvik80/go-network/network"
	log "github.com/sirupsen/logrus"
	"sync"
)

type DeviceMode int

const (
	SERVER DeviceMode = iota
	CLIENT
)

func GetDeviceMode(mode string) DeviceMode {
	switch mode {
	case "SERVER":
		return SERVER
	case "CLIENT":
		return CLIENT
	}

	return CLIENT
}

type Device interface {
	Address() string
	AllowedAddresses() []string
	Mode() DeviceMode
	Codec() codec.Codec
	Name() string

	SetChannel(channel network.Channel)
	SendToChannel(msg interface{})
}

type simpleDevice struct {
	exchange.Exchange
	cfg     DeviceConfig
	mutex   sync.Mutex
	channel network.Channel
}

func NewDevice(cfg DeviceConfig, exchange exchange.Exchange) Device {
	switch cfg.Type {
	case "dws":
		return NewDwsDevice(cfg, exchange)
	case "plc":
		return NewPlcDevice(cfg, exchange)
	default:
		return nil
	}
}

func (d *simpleDevice) Address() string {
	return d.cfg.Address
}

func (d *simpleDevice) AllowedAddresses() []string {
	return d.cfg.AllowedAddresses
}

func (d *simpleDevice) Mode() DeviceMode {
	return GetDeviceMode(d.cfg.Mode)
}

func (d *simpleDevice) Codec() codec.Codec {
	return codec.GetCodec(d.cfg.Codec)
}

func (d *simpleDevice) Name() string {
	return d.cfg.Name
}

func (d *simpleDevice) SetChannel(channel network.Channel) {
	d.mutex.Lock()
	d.channel = channel
	d.mutex.Unlock()
}

func (d *simpleDevice) SendToChannel(msg interface{}) {
	d.mutex.Lock()
	if d.channel != nil {
		d.channel.Write(msg)
	} else {
		log.WithFields(log.Fields{"module": "dev"}).Warn(d.Name(), " drop message, no connection")
	}
	d.mutex.Unlock()
}
