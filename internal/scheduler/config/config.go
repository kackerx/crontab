package config

import (
    "github.com/pkg/errors"
    "github.com/spf13/viper"
)

type Config struct {
    Mongodb `mapstructure:"mongodb"`
}

type Mongodb struct {
    Uri     string `mapstructure:"uri"`
    Timeout int64  `mapstructure:"timeout"`
}

func NewConfig() (*Config, error) {
    v := viper.New()
    //v.SetConfigType("yaml")
    v.SetConfigFile("configs/cron-worker.yaml")
    if err := v.ReadInConfig(); err != nil {
        return nil, errors.Wrap(err, "读取配置文件失败")
    }

    var config Config
    err := v.Unmarshal(&config)
    if err != nil {
        return nil, errors.Wrap(err, "配置反序列化失败")
    }

    return &config, nil
}
