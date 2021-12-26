package middleware

import "strings"

type DeviceMode int

const (
	SERVER DeviceMode = iota
	CLIENT
)

func GetDeviceMode(mode string) DeviceMode {
	switch strings.ToUpper(mode) {
	case "SERVER":
		return SERVER
	case "CLIENT":
		return CLIENT
	}

	return CLIENT
}

type DeviceConfig struct {
	Name             string   `yaml:"name"`
	Type             string   `yaml:"type"`
	Address          string   `yaml:"address"`
	Mode             string   `yaml:"mode"`
	Codec            string   `yaml:"codec"`
	AllowedAddresses []string `yaml:"allowed-addresses"`
	Router           *string  `yaml:"router,omitempty"`
}

type LinkConfig struct {
	Id      int            `yaml:"id"`
	Name    string         `yaml:"link"`
	Devices []DeviceConfig `yaml:"devices"`
	Router  string         `yaml:"router"`
}
