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
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var defaultPaths = []string{"../config", ".", "./config"}

const (
	// DefaultSeparator is a default list and map separator character
	DefaultSeparator      = ","
	defaultFileType       = "yaml"
	defaultFileNamePrefix = "config"
)

const (
	// TagEnv is the Name of the environment variable or a list of names
	TagEnv = "env"
	// TagEnvLayout is the Value parsing layout (for types like time.Time)
	TagEnvLayout = "env-layout"
	// TagEnvDefault is the Default value
	TagEnvDefault = "env-default"
	// TagEnvSeparator is the Custom list and map separator
	TagEnvSeparator = "env-separator"
	// TagEnvDescription is the Environment variable description
	TagEnvDescription = "env-description"
	// TagEnvRequired is the Flag to mark a field as required
	TagEnvRequired = "env-required"
)

// Setter is an interface for a custom value setter.
//
// To implement a custom value setter you need to add a SetValue function to your type that will receive a string raw value:
//
// 	type MyField string
//
// 	func (f *MyField) SetValue(s string) error {
// 		if s == "" {
// 			return fmt.Errorf("field value can't be empty")
// 		}
// 		*f = MyField("my field is: " + s)
// 		return nil
// 	}
type Setter interface {
	SetValue(string) error
}

// Configurator is the main struct to access configuration functionalities.
type Configurator struct {
	env            string
	envPrefix      string
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
		envPrefix:      "",
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

// WithEnvPrefix configure the env vars prefix.
func (c *Configurator) WithEnvPrefix(envPrefix string) *Configurator {
	if !strings.HasSuffix(envPrefix, "_") {
		envPrefix = envPrefix + "_"
	}
	c.envPrefix = envPrefix
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

// WithPath adds a path into the serach paths.
func (c *Configurator) WithPath(path string) *Configurator {
	c.paths = append(c.paths, path)
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

// Load reads configuration from default file into the cfg structure.
func (c *Configurator) Load(cfg interface{}) error {
	var err error
	for _, p := range c.paths {
		err = c.LoadFromFile(p+"/"+c.FileName(), cfg)
		if err == nil {
			return nil
		}
	}
	return err
}

// LoadFromFile reads configuration from the specified file into the cfg structure.
func (c *Configurator) LoadFromFile(path string, cfg interface{}) error {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	err = c.Parser().Parse(f, cfg)
	if err != nil {
		if e, ok := err.(*os.PathError); ok {
			return e
		}
		return fmt.Errorf("config file parsing error: %s", err.Error())
	}

	return c.readEnvVars(cfg)
}

func (c *Configurator) readStructMetadata(cfgRoot interface{}) ([]metadata, error) {
	cfgStack := []interface{}{cfgRoot}
	metas := make([]metadata, 0)

	for i := 0; i < len(cfgStack); i++ {
		s := reflect.ValueOf(cfgStack[i])

		// unwrap pointer
		if s.Kind() == reflect.Ptr {
			s = s.Elem()
		}

		// process only structures
		if s.Kind() != reflect.Struct {
			return nil, fmt.Errorf("wrong type %v", s.Kind())
		}
		sTypeInfo := s.Type()

		// read tags
		for idx := 0; idx < s.NumField(); idx++ {
			fType := sTypeInfo.Field(idx)
			var (
				defValue  *string
				layout    *string
				separator string
			)

			// process nested structure (except of time.Time)
			if fld := s.Field(idx); fld.Kind() == reflect.Struct {
				// add structure to parsing stack
				if fld.Type() != reflect.TypeOf(time.Time{}) {
					cfgStack = append(cfgStack, fld.Addr().Interface())
					continue
				}
				// process time.Time
				if l, ok := fType.Tag.Lookup(TagEnvLayout); ok {
					layout = &l
				}
			}

			// check if the field value can be changed
			if !s.Field(idx).CanSet() {
				continue
			}

			if def, ok := fType.Tag.Lookup(TagEnvDefault); ok {
				defValue = &def
			}

			if sep, ok := fType.Tag.Lookup(TagEnvSeparator); ok {
				separator = sep
			} else {
				separator = DefaultSeparator
			}

			_, required := fType.Tag.Lookup(TagEnvRequired)
			envList := make([]string, 0)

			if envs, ok := fType.Tag.Lookup(TagEnv); ok && len(envs) != 0 {
				envList = strings.Split(envs, DefaultSeparator)
			}

			metas = append(metas, metadata{
				env:         envList,
				fieldName:   s.Type().Field(idx).Name,
				fieldValue:  s.Field(idx),
				defValue:    defValue,
				layout:      layout,
				separator:   separator,
				description: fType.Tag.Get(TagEnvDescription),
				required:    required,
			})
		}
	}

	return metas, nil
}

func (c *Configurator) readEnvVars(cfg interface{}) error {
	metaInfo, err := c.readStructMetadata(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, meta := range metaInfo {
		var rawValue *string

		for _, env := range meta.env {
			if value, ok := os.LookupEnv(c.envPrefix + env); ok {
				rawValue = &value
				break
			}
		}

		if rawValue == nil && meta.required && meta.isFieldValueZero() {
			err := fmt.Errorf("field %q is required but the value is not provided",
				meta.fieldName)
			return err
		}

		if rawValue == nil && meta.isFieldValueZero() {
			rawValue = meta.defValue
		}

		if rawValue == nil {
			continue
		}

		if err := c.parseValue(meta.fieldValue, *rawValue, meta.separator, meta.layout); err != nil {
			return err
		}
	}

	return nil
}

// parseValue parses value into the corresponding field.
// In case of maps and slices it uses provided separator to split raw value string
func (c *Configurator) parseValue(field reflect.Value, value, sep string, layout *string) error {
	// TODO: simplify recursion

	if field.CanInterface() {
		if cs, ok := field.Interface().(Setter); ok {
			return cs.SetValue(value)
		} else if csp, ok := field.Addr().Interface().(Setter); ok {
			return csp.SetValue(value)
		}
	}

	valueType := field.Type()

	switch valueType.Kind() {
	// parse string value
	case reflect.String:
		field.SetString(value)

	// parse boolean value
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	// parse integer (or time) value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Kind() == reflect.Int64 && valueType.PkgPath() == "time" && valueType.Name() == "Duration" {
			// try to parse time
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))

		} else {
			// parse regular integer
			number, err := strconv.ParseInt(value, 0, valueType.Bits())
			if err != nil {
				return err
			}
			field.SetInt(number)
		}

	// parse unsigned integer value
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, err := strconv.ParseUint(value, 0, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetUint(number)

	// parse floating point value
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(value, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(number)

	// parse sliced value
	case reflect.Slice:
		sliceValue, err := c.parseSlice(valueType, value, sep, layout)
		if err != nil {
			return err
		}

		field.Set(*sliceValue)

	// parse mapped value
	case reflect.Map:
		mapValue, err := c.parseMap(valueType, value, sep, layout)
		if err != nil {
			return err
		}

		field.Set(*mapValue)

	case reflect.Struct:
		// process time.Time only
		if valueType.PkgPath() == "time" && valueType.Name() == "Time" {

			var l string
			if layout != nil {
				l = *layout
			} else {
				l = time.RFC3339
			}
			val, err := time.Parse(l, value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(val))
		}

	default:
		return fmt.Errorf("unsupported type %s.%s", valueType.PkgPath(), valueType.Name())
	}

	return nil
}

// parseSlice parses value into a slice of given type
func (c *Configurator) parseSlice(valueType reflect.Type, value string, sep string, layout *string) (*reflect.Value, error) {
	sliceValue := reflect.MakeSlice(valueType, 0, 0)
	if valueType.Elem().Kind() == reflect.Uint8 {
		sliceValue = reflect.ValueOf([]byte(value))
	} else if len(strings.TrimSpace(value)) != 0 {
		values := strings.Split(value, sep)
		sliceValue = reflect.MakeSlice(valueType, len(values), len(values))

		for i, val := range values {
			if err := c.parseValue(sliceValue.Index(i), val, sep, layout); err != nil {
				return nil, err
			}
		}
	}
	return &sliceValue, nil
}

// parseMap parses value into a map of given type
func (c *Configurator) parseMap(valueType reflect.Type, value string, sep string, layout *string) (*reflect.Value, error) {
	mapValue := reflect.MakeMap(valueType)
	if len(strings.TrimSpace(value)) != 0 {
		pairs := strings.Split(value, sep)
		for _, pair := range pairs {
			kvPair := strings.SplitN(pair, ":", 2)
			if len(kvPair) != 2 {
				return nil, fmt.Errorf("invalid map item: %q", pair)
			}
			k := reflect.New(valueType.Key()).Elem()
			err := c.parseValue(k, kvPair[0], sep, layout)
			if err != nil {
				return nil, err
			}
			v := reflect.New(valueType.Elem()).Elem()
			err = c.parseValue(v, kvPair[1], sep, layout)
			if err != nil {
				return nil, err
			}
			mapValue.SetMapIndex(k, v)
		}
	}
	return &mapValue, nil
}

// GetDescription returns a description of environment variables.
// You can provide a custom header text.
func (c *Configurator) GetDescription(cfg interface{}, headerText *string) (string, error) {
	meta, err := c.readStructMetadata(cfg)
	if err != nil {
		return "", err
	}

	var header, description string

	if headerText != nil {
		header = *headerText
	} else {
		header = "Environment variables:"
	}

	for _, m := range meta {
		if len(m.env) == 0 {
			continue
		}

		for idx, env := range m.env {

			elemDescription := fmt.Sprintf("\n  %s %s", env, m.fieldValue.Kind())
			if idx > 0 {
				elemDescription += fmt.Sprintf(" (alternative to %s)", m.env[0])
			}
			elemDescription += fmt.Sprintf("\n    \t%s", m.description)
			if m.defValue != nil {
				elemDescription += fmt.Sprintf(" (default %q)", *m.defValue)
			}
			description += elemDescription
		}
	}

	if description != "" {
		return header + description, nil
	}
	return "", nil
}

// Usage returns a configuration usage help.
// Other usage instructions can be wrapped in and executed before this usage function.
// The default output is STDERR.
func (c *Configurator) Usage(cfg interface{}, headerText *string, usageFuncs ...func()) func() {
	return c.FUsage(os.Stderr, cfg, headerText, usageFuncs...)
}

// FUsage prints configuration help into the custom output.
// Other usage instructions can be wrapped in and executed before this usage function
func (c *Configurator) FUsage(w io.Writer, cfg interface{}, headerText *string, usageFuncs ...func()) func() {
	return func() {
		for _, fn := range usageFuncs {
			fn()
		}

		_ = flag.Usage

		text, err := c.GetDescription(cfg, headerText)
		if err != nil {
			return
		}
		if len(usageFuncs) > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, text)
	}
}
