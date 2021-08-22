package configs

import (
	"crontab/pkg"
	"encoding/json"
	"io/ioutil"
)

type MongodbConfig struct {
	Uri            string `json:"uri"`
	ConnectTimeout int    `json:"connect_timeout"`
	DB             string `json:"db"`
	Collection     string `json:"collection"`
}

type Config struct {
	EtcdConfig    *pkg.EtcdConfig `json:"etcdConfig"`
	MongodbConfig *MongodbConfig  `json:"mongodb"`
}

func InitConfig(configFile string) (*Config, error) {
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
