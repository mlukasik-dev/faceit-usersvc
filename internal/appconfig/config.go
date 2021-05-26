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

var AppConfig Config

func Init(configFile []byte) error {
	if err := config.Load(bytes.NewReader(configFile), &AppConfig, "yaml"); err != nil {
		return err
	}
	return nil
}
