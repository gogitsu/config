// Copyright 2020 Luca Stasio. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// The gogitsu/config is heavily based on cleanenv by
// (c) 2019 Ilya Kaznacheev
// https://github.com/ilyakaznacheev/cleanenv
// cleanenv copyright notice and permission notice can
// be found in the ATTRIBUTIONS file.

// Package config implements configuration components of gogitsu lib.
package config

import (
	"encoding/json"
	"io"
	"os"

	"github.com/joho/godotenv"

	"github.com/BurntSushi/toml"

	"gopkg.in/yaml.v2"
)

// Parser is the interface to define parsing functionality
// to read from a reader and parse values into a struct.
// A concrete Parser will read from the input "r" reader and
// parse values into the "s" struct using a concrete decoder.
type Parser interface {
	Parse(r io.Reader, s interface{}) error
}

// YAMLParser will parse env from yaml format.
type YAMLParser struct{}

// JSONParser will parse env from json format.
type JSONParser struct{}

// TOMLParser will parse env from toml format.
type TOMLParser struct{}

// ENVParser will parse env from env file.
// It does not fill the struct, but set variables into the env.
type ENVParser struct{}

// Parse parses variables from YAML into the input struct.
func (yp *YAMLParser) Parse(r io.Reader, s interface{}) error {
	return yaml.NewDecoder(r).Decode(s)
}

// Parse parses variables from JSON into the input struct.
func (jp *JSONParser) Parse(r io.Reader, s interface{}) error {
	return json.NewDecoder(r).Decode(s)
}

// Parse parses variables from TOML into the input struct.
// Here we skip returned metadata.
func (tp *TOMLParser) Parse(r io.Reader, s interface{}) error {
	_, err := toml.DecodeReader(r, s)
	return err
}

// Parse parses variables from ENV into the input struct.
// It does not fill the struct, but set variables into the env.
func (ep *ENVParser) Parse(r io.Reader, s interface{}) error {
	vars, err := godotenv.Parse(r)
	if err != nil {
		return err
	}

	for env, val := range vars {
		os.Setenv(env, val)
	}
	return nil
}

// NewParser return a new Parser according to the config file format.
func NewParser(format string) Parser {
	var p Parser
	switch format {
	case "yaml", "yml", ".yaml", ".yml":
		p = &YAMLParser{}
	case "json", ".json":
		p = &JSONParser{}
	case "toml", ".toml":
		p = &TOMLParser{}
	case "env", ".env":
		p = &ENVParser{}
	default:
		panic("'%s' file format not supported")
	}
	return p
}
