package export

import "time"

type globalConfig struct {
	ClientTimeout *time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

type Config struct {
	Global  globalConfig `json:"global,omitempty" yaml:"global,omitempty"`
	Devices []struct {
		globalConfig
		Password string `json:"password" yaml:"password"`
		URL      string `json:"url" yaml:"url"`
	} `json:"devices" yaml:"devices"`
}
