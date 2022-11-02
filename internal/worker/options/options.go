package options

import genericoptions "github.com/kackerx/crontab/pkg/options"

type Options struct {
	EtcdOptions *genericoptions.EtcdOptions
}

func NewOptions() *Options {
	return &Options{
		EtcdOptions: genericoptions.NewEtcdOptions(),
	}
}
