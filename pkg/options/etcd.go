package options

import (
	"github.com/spf13/viper"
)

type EtcdOptions struct {
	Endpoints []string `mapstructure:"endpoints"`
	Timeout   int      `mapstructure:"timeout"`
}

func NewEtcdOptions() *EtcdOptions {
	//return &EtcdOptions{
	//	Endpoints: []string{"0.0.0.0:2379"},
	//	Timeout:   3333,
	//}
	v := viper.New()
	v.SetConfigFile("configs/cron-worker.yaml")

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	var eo EtcdOptions
	if err := v.UnmarshalKey("etcd", &eo); err != nil {
		panic(err)
	}

	return &eo
}
