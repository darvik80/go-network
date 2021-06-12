package middleware

import (
	"darvik80/go-network/exchange"
	log "github.com/sirupsen/logrus"
)

type plcDevice struct {
	simpleDevice
}

func NewPlcDevice(cfg DeviceConfig, ex exchange.Exchange) *plcDevice {
	dev := &plcDevice{
		simpleDevice{
			Exchange: ex,
			cfg:      cfg,
		},
	}

	ex.Subscribe(func(source exchange.Source, report exchange.DwsReport) {
		dev.onDwsMessage(source.(Device), report)
	})

	ex.Subscribe(func(source exchange.Source, report exchange.SortReport) {
		dev.onPlcMessage(source.(Device), report)
	})

	return dev
}

func (d *plcDevice) onDwsMessage(source Device, report exchange.DwsReport) {
	log.WithFields(log.Fields{"module": "plc"}).Info(source.Name(), " DWSReport")
	d.Publish(d, exchange.SortReport{Id: report.Id, ChuteId: 10})
}

func (d *plcDevice) onPlcMessage(source Device, report exchange.SortReport) {
	log.WithFields(log.Fields{"module": "plc"}).Info(source.Name(), " PlcReport")
	d.send(report)
}
