package middleware

import "darvik80/go-network/exchange"

type simple struct {
	mode     DeviceMode
	exchange exchange.Exchange
}
