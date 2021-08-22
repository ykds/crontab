package configs

import (
	"crontab/pkg"
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiPort      int             `json:"apiPort"`
	ReadTimeout  int             `json:"readTimeout"`
	WriteTimeout int             `json:"writeTimeout"`
	EtcdConfig   *pkg.EtcdConfig `json:"etcdConfig"`
}

func InitConfig(configFile string) (*Config, error) {
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var C *Config

	err = json.Unmarshal(config, &C)
	if err != nil {
		return nil, err
	}

	return C, nil
}
