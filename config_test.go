package config

import (
	"os"
	"strings"
	"testing"
)

var jsonData = `{"host": "localhost","port": 8080}`

const (
	ExpectedCfgPort        = 8080
	ExpectedCfgHost        = "localhost"
	ExpectedServiceGroup   = "test"
	ExpectedServiceName    = "NAME-FROM-ENV"
	ExpectedServiceVersion = "0.0.1"
)

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
	cfg := JSONConfig{}
	err := c.Parser().Parse(strings.NewReader(jsonData), &cfg)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if cfg.Host != ExpectedCfgHost {
		t.Fatalf("cfg.Host expected '%s' is '%s'", ExpectedCfgHost, cfg.Host)
	}

	if cfg.Port != ExpectedCfgPort {
		t.Fatalf("cfg.Port expected '%d' is '%d'", ExpectedCfgPort, cfg.Port)
	}
}

func TestConfiguratorFromFile(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("SVC_NAME", ExpectedServiceName)
	var config *Cfg = &Cfg{}
	c := NewConfiguratorFor("yaml")
	err := c.Load(config)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if config.Service.Group != ExpectedServiceGroup {
		t.Fatalf("Service.Group expected '%s' is '%s'", ExpectedServiceGroup, config.Service.Group)
	}

	if config.Service.Name != ExpectedServiceName {
		t.Fatalf("Service.Name expected '%s' is '%s'", ExpectedServiceName, config.Service.Name)
	}

	if config.Service.Version != ExpectedServiceVersion {
		t.Fatalf("Service.Version expected '%s' is '%s'", ExpectedServiceVersion, config.Service.Version)
	}
}
