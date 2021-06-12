package middleware

import "darvik80/go-network/exchange"

type Link interface {
	Send(source exchange.Source, report exchange.Report)
}

type link struct {
	devices  []Device
	exchange exchange.Exchange
}

func NewLink(device ...Device) *link {
	return &link{
		devices:  device,
		exchange: exchange.NewChanExchange(64, 4),
	}
}

func (l *link) Send(source exchange.Source, report exchange.Report) {
	l.exchange.Publish(source, report)
}
