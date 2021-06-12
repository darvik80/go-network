package config

import (
	"darvik80/go-network/middleware"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"strings"
)

const DefaultGroup = "DEFAULT_GROUP"
const DefaultDataId = "config.yml"

type Config struct {
	RocketMQ struct {
		Nameserver string `yaml:"name-server"`
	} `yaml:"rocket-mq"`
	Links []middleware.LinkConfig `yaml:"links"`
}

func readConfig(r io.Reader, cfg *Config) error {
	decoder := yaml.NewDecoder(r)
	err := decoder.Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}

func ReadFileConfig(file string, cfg *Config) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return readConfig(f, cfg)
}

func ReadCloudConfig(name string, cfg *Config) error {
	//create clientConfig
	clientConfig := constant.ClientConfig{
		AppName:             "bifrost",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
		Username:            "nacos",
		Password:            "nacos",
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "nacos",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	if err != nil {
		return err
	}

	content, err := configClient.GetConfig(
		vo.ConfigParam{
			DataId: name,
			Group:  DefaultGroup,
		},
	)

	if err != nil {
		return err
	}

	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: name,
		Group:  DefaultGroup,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
		},
	})

	return readConfig(strings.NewReader(content), cfg)
}

func ReadConfig() (*Config, error) {
	cfg := &Config{}
	err := ReadFileConfig(DefaultDataId, cfg)
	if err != nil {
		return nil, err
	}

	err = ReadCloudConfig(DefaultDataId, cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
