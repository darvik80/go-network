package middleware

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

type Codec int

const (
	CodecUnknown Codec = iota
	CodecSswDws
	CodecSswPlc
)

func GetCodec(codec string) Codec {
	switch codec {
	case "SSW_DWS":
		return CodecSswDws
	case "SSW_PLC":
		return CodecSswPlc
	}

	return CodecUnknown
}

type Device interface {
	Address() string
	Mode() DeviceMode
	Codec() Codec
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

func (d *simpleDevice) Mode() DeviceMode {
	return GetDeviceMode(d.cfg.Mode)
}

func (d *simpleDevice) Codec() Codec {
	return GetCodec(d.cfg.Codec)
}

func (d *simpleDevice) Name() string {
	return d.cfg.Name
}
