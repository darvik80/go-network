package network

import "github.com/darvik80/go-network/network"

type Transport interface {
	Start(func(p network.Pipeline) network.Pipeline) error
	Shutdown()
}
