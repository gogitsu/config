package config

import (
	"os"
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

type Cfg struct {
	Service struct {
		Group   string
		Name    string `env:"SVC_NAME"`
		Version string
	}
}

func TestConfigurator(t *testing.T) {
	c := NewConfigurator().WithFormat("json")
	t.Logf("c.paths: %+v, c.fileName: '%s'", c.paths, c.FileName())
	cfg := JSONConfig{}
	err := c.Parser().Parse(strings.NewReader(jsonData), &cfg)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("CONFIGURATION: %+v\n", cfg)
}

func TestConfiguratorFromFile(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("SVC_NAME", "service-name-from-env")
	var config *Cfg = &Cfg{}
	c := NewConfiguratorFor("yaml")
	err := c.Load(config)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("CONFIGURATION: %+v\n", config)
}
