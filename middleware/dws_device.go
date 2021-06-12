package middleware

import "darvik80/go-network/exchange"

type dwsDevice struct {
	simpleDevice
}

func NewDwsDevice(cfg DeviceConfig, exchange exchange.Exchange) *dwsDevice {
	return &dwsDevice{
		simpleDevice{
			Exchange: exchange,
			cfg:      cfg,
		},
	}
}

func (d *dwsDevice) OnMessage(source exchange.Source, report exchange.DwsReport) {

}
