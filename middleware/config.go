package middleware

type DeviceConfig struct {
	Name             string   `yaml:"name""`
	Address          string   `yaml:"address"`
	Mode             string   `yaml:"mode"`
	Codec            string   `yaml:"codec"`
	AllowedAddresses []string `yaml:"allowed-addresses"`
}
