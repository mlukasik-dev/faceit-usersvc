package appconfig

import (
	"bytes"

	"github.com/gopher-lib/config"
)

type Config struct {
	Port    string
	Mongodb struct {
		URI string
	}
}

// AppConfig contains application configuration.
var AppConfig Config

// Init takes configuration file content in yaml format,
// parses it and initilizes appconfig.AppConfig struct.
func Init(cfg []byte) error {
	return config.Load(bytes.NewReader(cfg), &AppConfig, "yaml")
}
