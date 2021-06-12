package middleware

type DeviceConfig struct {
	Name             string   `yaml:"name"`
	Type             string   `yaml:"type"`
	Address          string   `yaml:"address"`
	Mode             string   `yaml:"mode"`
	Codec            string   `yaml:"codec"`
	AllowedAddresses []string `yaml:"allowed-addresses"`
}

type LinkConfig struct {
	Name    string         `yaml:"link"`
	Devices []DeviceConfig `yaml:"devices"`
}
