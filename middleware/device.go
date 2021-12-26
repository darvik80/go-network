package middleware

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/middleware/codec"
	"darvik80/go-network/network"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

type Device interface {
	exchange.Exchange
	Id() int
	Address() string
	AllowedAddresses() []string
	Mode() DeviceMode
	Codec() codec.Codec
	Name() string

	SetChannel(channel network.Channel)
	Send(msg interface{})
}

func NewDevice(id int, cfg DeviceConfig, exchange exchange.Exchange) Device {
	switch strings.ToUpper(cfg.Type) {
	case "DWS":
		return NewDwsDevice(id, cfg, exchange)
	case "PLC":
		return NewPlcDevice(id, cfg, exchange)
	case "SCADA":
		return NewScadaDevice(id, cfg, exchange)
	default:
		return nil
	}
}

type simpleDevice struct {
	exchange.Exchange
	id      int
	cfg     DeviceConfig
	mutex   sync.Mutex
	channel network.Channel
}

func (d *simpleDevice) Id() int {
	return d.id
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

func (d *simpleDevice) Send(msg interface{}) {
	d.mutex.Lock()
	if d.channel != nil {
		d.channel.Write(msg)
	} else {
		log.WithFields(log.Fields{"module": "dev"}).Warn(d.Name(), " drop message, no connection")
	}
	d.mutex.Unlock()
}

type dwsDevice struct {
	simpleDevice
}

func NewDwsDevice(id int, cfg DeviceConfig, ex exchange.Exchange) *dwsDevice {
	log.WithFields(log.Fields{"module": "dws"}).Infof("Register DWS: %s", cfg.Name)

	dev := &dwsDevice{
		simpleDevice{
			Exchange: ex,
			id:       id,
			cfg:      cfg,
		},
	}

	ex.Subscribe(func(source exchange.Source, report exchange.StdDwsReport) {
		dev.onDwsMessage(source.(Device), report)
	})

	return dev
}

func (d *dwsDevice) onDwsMessage(source Device, report exchange.StdDwsReport) {
	log.WithFields(log.Fields{"module": "dws"}).Info(source.Name(), " DWSReport")
	d.Send(report)
}

type plcDevice struct {
	id int
	simpleDevice
}

func NewPlcDevice(id int, cfg DeviceConfig, ex exchange.Exchange) *plcDevice {
	dev := &plcDevice{
		id,
		simpleDevice{
			Exchange: ex,
			id:       id,
			cfg:      cfg,
		},
	}

	log.WithFields(log.Fields{"module": "plc"}).Infof("Register PLC: %s", cfg.Name)

	return dev
}

type scadaDevice struct {
	simpleDevice
}

func NewScadaDevice(id int, cfg DeviceConfig, ex exchange.Exchange) *scadaDevice {
	dev := &scadaDevice{
		simpleDevice{
			Exchange: ex,
			id:       id,
			cfg:      cfg,
		},
	}

	log.WithFields(log.Fields{"module": "scada"}).Infof("Register SCADA: %s", cfg.Name)

	return dev
}
