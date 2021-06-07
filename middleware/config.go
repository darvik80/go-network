package middleware

type DeviceConfig struct {
	Address          string   `yaml:"address"`
	Port             int      `yaml:"port"`
	Mode             string   `yaml:"mode"`
	Codec            string   `yaml:"codec"`
	AllowedAddresses []string `yaml:"allowed-addresses"`
}
