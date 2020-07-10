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

import "reflect"

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
	parser Parser
}

// NewConfigurator returns a new Configurator instance.
// Here we not set a parser.
func NewConfigurator() *Configurator {
	return &Configurator{}
}

// NewConfiguratorFor returns a new Configurator with
// a configured parser for the input format.
func NewConfiguratorFor(format string) *Configurator {
	return &Configurator{parser: NewParser(format)}
}

// WithFormat configure the internal Parser according to the input format.
func (c *Configurator) WithFormat(format string) *Configurator {
	c.parser = NewParser(format)
	return c
}

// WithParser configure the internal Parser.
func (c *Configurator) WithParser(parser Parser) *Configurator {
	c.parser = parser
	return c
}

// func (c *Configurator) readStructMetadata(cfg interface{}) *metadata {

// }
