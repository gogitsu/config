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
	"os"
	"reflect"
)

var defaultPaths = []string{"../config", ".", "./config"}

const defaultFileNamePrefix = "config"
const defaultFileType = "yaml"

// metadata is the struct where all the meta information
// about a configuration field are stored.
type metadata struct {
	env   []string
	field struct {
		name     string
		value    reflect.Value
		defValue *string
	}
	layout      *string
	separator   string
	description string
	required    bool
}

// Configurator is the main struct to access configuration functionalities.
type Configurator struct {
	env            string
	parser         Parser
	paths          []string
	fileNamePrefix string
	fileType       string
}

// NewConfigurator returns a new Configurator instance.
// Here we not set a parser.
func NewConfigurator() *Configurator {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	fileNamePrefix := defaultFileNamePrefix + "-" + env

	return &Configurator{
		env:            env,
		paths:          defaultPaths,
		fileNamePrefix: fileNamePrefix,
		fileType:       defaultFileType,
	}
}

// NewConfiguratorFor returns a new Configurator with
// a configured parser for the input format.
func NewConfiguratorFor(format string) *Configurator {
	c := NewConfigurator().WithFormat(format)
	// c.parser = NewParser(format)
	return c
}

// WithFormat configure the internal Parser according to the input format.
func (c *Configurator) WithFormat(format string) *Configurator {
	c.fileType = format
	c.parser = NewParser(format)
	return c
}

// WithFileNamePrefix .
func (c *Configurator) WithFileNamePrefix(prefix string) *Configurator {
	c.fileNamePrefix = prefix + "-" + c.env
	return c
}

// WithParser configure the internal Parser.
func (c *Configurator) WithParser(parser Parser) *Configurator {
	c.parser = parser
	return c
}

// Parser returns the internal parser instance.
func (c *Configurator) Parser() Parser {
	return c.parser
}

// FileName returns the configuration filename.
func (c *Configurator) FileName() string {
	return c.fileNamePrefix + "." + c.fileType
}

// func (c *Configurator) readStructMetadata(cfg interface{}) *metadata {

// }
