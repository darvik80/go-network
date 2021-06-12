package middleware

import "darvik80/go-network/middleware/codec"

type DeviceMode int

const (
	SERVER DeviceMode = iota
	CLIENT
)

func GetDeviceMode(mode string) DeviceMode {
	switch mode {
	case "SERVER":
		return SERVER
	case "CLIENT":
		return CLIENT
	}

	return CLIENT
}

type Device interface {
	Address() string
	AllowedAddresses() []string
	Mode() DeviceMode
	Codec() codec.Codec
	Name() string
}

func NewDevice(cfg DeviceConfig) *simpleDevice {
	return &simpleDevice{
		cfg: cfg,
	}
}

type simpleDevice struct {
	cfg DeviceConfig
}

func (d *simpleDevice) Address() string {
	return d.cfg.Address
}

func (d *simpleDevice) AllowedAddresses() []string {
	return d.cfg.AllowedAddresses
}

func (d *simpleDevice) Mode() DeviceMode {
	return GetDeviceMode(d.cfg.Mode)
}

func (d *simpleDevice) Codec() codec.Codec {
	return codec.GetCodec(d.cfg.Codec)
}

func (d *simpleDevice) Name() string {
	return d.cfg.Name
}
