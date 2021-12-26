package middleware

import (
	"darvik80/go-network/exchange"
	"darvik80/go-network/middleware/route"
	log "github.com/sirupsen/logrus"
)

type Machine interface {
	Send(source exchange.Source, report exchange.Message)
}

type dwsPlc struct {
	devices  []Device
	exchange exchange.Exchange
}

func NewDwsPlcMachine(ex exchange.Exchange, device ...Device) *dwsPlc {
	var m = &dwsPlc{
		devices:  device,
		exchange: ex,
	}
	m.exchange.Subscribe(func(source exchange.Source, msg exchange.StdDwsReport) {
		m.onDwsMessage(source.(Device), msg)
	})

	m.exchange.Subscribe(func(source exchange.Source, msg exchange.StdKeepAliveRequest) {
		m.onKeepAliveMessage(source.(Device), msg)
	})

	return m
}

func (d *dwsPlc) onDwsMessage(source Device, report exchange.StdDwsReport) {
	log.WithFields(log.Fields{"module": "scada-machine"}).Info(source.Name(), " DWSReport")
}

func (d *dwsPlc) onKeepAliveMessage(source Device, msg exchange.StdKeepAliveRequest) {
	log.WithFields(log.Fields{"module": "scada-machine"}).Info(source.Name(), " KeepAlive")
}

func (d *dwsPlc) Send(source exchange.Source, report exchange.Message) {
	d.exchange.Publish(source, report)
}

type scada struct {
	device   Device
	exchange exchange.Exchange
	router   route.SortingRouter
}

func NewScadaMachine(ex exchange.Exchange, router route.SortingRouter, device ...Device) *scada {
	var m = &scada{
		device:   device[0],
		exchange: ex,
		router:   router,
	}

	m.exchange.Subscribe(func(source exchange.Source, msg exchange.StdDwsReport) {
		m.onDwsMessage(source.(Device), msg)
	})

	m.exchange.Subscribe(func(source exchange.Source, msg exchange.StdKeepAliveRequest) {
		m.onKeepAliveMessage(source.(Device), msg)
	})

	return m
}

func (s *scada) Send(source exchange.Source, report exchange.Message) {
	s.exchange.Publish(source, report)
}

func (s *scada) onDwsMessage(source Device, msg exchange.StdDwsReport) {
	log.WithFields(log.Fields{"module": "scada-machine"}).Info(source.Name(), " DWSReport")
	var res = s.router.Advice(route.SortingTask{
		Barcodes: msg.Barcodes,
	})

	var destinations []int
	for _, item := range res {
		destinations = append(destinations, item.Destination)
	}
	source.Send(exchange.StdSortRequest{
		StdBarcodeMessage: exchange.StdBarcodeMessage{
			StdMessage: exchange.StdMessage{
				Id:       msg.Id,
				DeviceId: msg.DeviceId,
				Device:   msg.Device,
			},
			Barcodes: msg.Barcodes,
		},
		Destinations: destinations,
	})
}

func (s *scada) onKeepAliveMessage(source Device, msg exchange.StdKeepAliveRequest) {
	log.WithFields(log.Fields{"module": "scada-machine"}).Info(source.Name(), " KeepAlive")
	source.Send(exchange.StdKeepAliveResponse{
		StdMessage: exchange.StdMessage{
			Id:       msg.Id,
			DeviceId: msg.DeviceId,
			Device:   msg.Device,
		},
	})
}
