package config

import (
	"strings"
	"testing"
)

var yamlData = `
name: Frank
surname: Zappa
`

const ExpectedCfgName = "Frank"
const ExpectedCfgSurname = "Zappa"

type YAMLCfg struct {
	Name    string `yaml:"name"`
	Surname string `yaml:"surname"`
}

func TestJSONParser(t *testing.T) {
	yamlP := NewParser("yaml")
	cfg := YAMLCfg{}
	yamlP.Parse(strings.NewReader(yamlData), &cfg)

	if cfg.Name != "Frank" {
		t.Fatalf("expected cfg.Name = '%s' is '%s'", ExpectedCfgName, cfg.Name)
	}

	if cfg.Surname != "Zappa" {
		t.Fatalf("expected cfg.Surname = '%s' is '%s'", ExpectedCfgSurname, cfg.Surname)
	}
}
