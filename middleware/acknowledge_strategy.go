package middleware

import "darvik80/go-network/exchange"

type AcknowledgeStrategy interface {
	Advice(report exchange.StdSortReport) exchange.StdSortReport
}

type acknowledgeStrategy struct {
	device Device
	report chan exchange.StdSortReport
}

func NewAcknowledgeStrategy(device Device) AcknowledgeStrategy {
	var a = &acknowledgeStrategy{
		device: device,
	}
	device.Subscribe(func(source exchange.Source, msg exchange.StdSortReport) {

	})

	return a
}

func (a *acknowledgeStrategy) Advice(report exchange.StdSortReport) exchange.StdSortReport {
	return exchange.StdSortReport{}
}
