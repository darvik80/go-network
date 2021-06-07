package middleware

import (
	"darvik80/go-network/network"
	"darvik80/go-network/network/tcp"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type middleware struct {
	servers map[string]network.Transport
	clients map[string]network.Transport
}

func NewMiddleware(config []DeviceConfig) (*middleware, error) {
	mid := &middleware{
		servers: make(map[string]network.Transport),
		clients: make(map[string]network.Transport),
	}

	for _, deviceConfig := range config {
		key := deviceConfig.Address + ":" + strconv.Itoa(deviceConfig.Port)
		if deviceConfig.Mode == "SERVER" {
			if _, found := mid.servers[key]; !found {
				log.Info("[mid] create server: ", deviceConfig.Address, ":", deviceConfig.Port)
				server := tcp.NewServer(deviceConfig.Address, deviceConfig.Port)
				mid.servers[key] = server
			}
		} else if deviceConfig.Mode == "CLIENT" {
			if _, found := mid.servers[key]; !found {
				log.Info("[mid] create client: ", deviceConfig.Address, ":", deviceConfig.Port)
				client := tcp.NewClient(deviceConfig.Address, deviceConfig.Port)
				mid.servers[key] = client
			}
		}
	}

	for _, server := range mid.servers {
		// TODO: create server factory
		if err := server.Start(nil); err != nil {
			return nil, err
		}
	}

	for _, client := range mid.clients {
		// TODO: create client factory
		if err := client.Start(nil); err != nil {
			return nil, err
		}
	}

	return mid, nil
}

func (mid *middleware) Shutdown() {
	for _, server := range mid.servers {
		server.Shutdown()
	}

	for _, client := range mid.clients {
		client.Shutdown()
	}
}
