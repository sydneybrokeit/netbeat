// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period time.Duration `config:"period"`
	Sets []ConfigSet `config:"sets"`
}

type ConfigSet struct {
	Hosts        []string            `config:"hosts"`
	Port         uint16              `config:"port"`
	Community    string              `config:"community"`
	User         string              `config:"user"`
	AuthPassword string              `config:"authpass"`
	PrivPassword string              `config:"privpass"`
	Version      string              `config:"version"`
	OIDs         []map[string]string `config:"oids"`
}

var DefaultConfig = Config{
	Period: 1 * time.Second,
}
