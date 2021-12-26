package handler

import "darvik80/go-network/exchange"

type Handler interface {
	OnDwsMessage(source exchange.Source, report exchange.StdDwsReport)
	OnSortMessage(source exchange.Source, report exchange.StdDwsSortReport)
	OnDwsSortMessage(source exchange.Source, report exchange.StdDwsSortReport)
}
