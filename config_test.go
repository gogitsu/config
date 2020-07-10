package config

import (
	"strings"
	"testing"
)

var jsonData = `{"host": "localhost","port": 8080}`

const ExpectedCfgHost = "127.0.0.1"
const ExpectedCfgPort = 8080

type JSONConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func TestConfigurator(t *testing.T) {
	c := NewConfigurator().WithFormat("json").WithFileNamePrefix("CFG")
	t.Logf("c.paths: %+v, c.fileName: '%s'", c.paths, c.FileName())
	cfg := JSONConfig{}
	err := c.Parser().Parse(strings.NewReader(jsonData), &cfg)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf("CONFIGURATION: %+v\n", cfg)
}
