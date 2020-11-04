package client

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func ReadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}

type Config struct {
	Host        string `yaml:"host"`
	ExploitDir  string `yaml:"exploit_dir"`
	GrpcAuthKey string `yaml:"grpc_auth_key"`
}
