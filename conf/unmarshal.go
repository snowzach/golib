package conf

import (
	"fmt"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/providers/structs"
	"github.com/mitchellh/mapstructure"
)

// UnmarshalConf is used to configure the unmarshaler
type UnmarshalConf struct {
	// The path in the configuration to start unmarshaling from.
	// Leave empty to unmarshal the root config structure.
	Path string
	// FlatPath interprets keys with delimeters literally instead of recursively unmarshaling structs.
	FlatPaths bool
	// If unmarshaling to a struct it will set defaults in the struct first using
	// github.com/creasty/defaults tags before unmarshaling the config into the struct.
	StructDefaults bool
	// Normally if a configuration field is not already set in the config it will not be used
	// from the environment. Setting this to true will also scan for environment variables
	// associated to the struct fields/tags even if it is not already set in the config.
	StructEnvironment bool
	// If you wish to look for an environment variable with a prefix, you can set it here.
	StructEnvironmentPrefix string
	// Decoder config is the github.com/mitchellh/mapstructure.DecoderConfig used to umarshal
	// configuration into data structures.
	DecoderConfig *mapstructure.DecoderConfig
}

// Unmarshal configuration to dest (it should be a pointer).
// See help for UnmarshalConf for more information.
func (c *Conf) Unmarshal(dest interface{}, unmarshalConfig UnmarshalConf) error {

	// If no UnmarshalConf is specified, use the default
	if unmarshalConfig.DecoderConfig == nil {
		unmarshalConfig.DecoderConfig = DefaultDecoderConfig()
		unmarshalConfig.DecoderConfig.TagName = c.Tag
	}
	unmarshalConfig.DecoderConfig.Result = dest

	// Make a reference to the config that we might change
	cc := c

	// If it's a struct we are unmarshaling into...
	if t := reflect.TypeOf(dest); t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {

		// If there are defaults in the struct, load those first using github.com/creasty/defaults
		if unmarshalConfig.StructDefaults {
			if err := defaults.Set(dest); err != nil {
				return fmt.Errorf("could not set struct defaults: %w", err)
			}
		}

		if unmarshalConfig.StructEnvironment {
			// Create a new config instance and parse the struct into the new config.
			// This will ensure that every value in the struct has a config setting.
			// This is required if we want to override the values in the struct with
			// environment variables but there is no default configuration applied in
			// in the config already.
			structConfig := New()
			if err := structConfig.Load(structs.Provider(dest, unmarshalConfig.DecoderConfig.TagName), nil); err != nil {
				return fmt.Errorf("could not parse environment struct for config: %w", err)
			}

			// Create a new config and merge in the struct config where it belongs at the path.
			cc = New()
			if err := cc.MergeAt(structConfig.Koanf, unmarshalConfig.Path); err != nil {
				return fmt.Errorf("could not merge environment struct for config: %w", err)
			}

			// Finally merge in the original passed in config to this copy.
			if err := cc.Merge(c.Koanf); err != nil {
				return fmt.Errorf("could not merge config: %w", err)
			}

			// Reparse any environment variables on our copy taking into account
			// any struct values that may exist but are not set in the config.
			if err := cc.ParseEnvPrefix(unmarshalConfig.StructEnvironmentPrefix); err != nil {
				return fmt.Errorf("could not reload env: %w", err)
			}
		}
	}

	// Get the source map
	source := cc.Get(unmarshalConfig.Path)
	// Flatten if requested
	if unmarshalConfig.FlatPaths {
		if f, ok := source.(map[string]interface{}); ok {
			fmp, _ := maps.Flatten(f, nil, c.Delimiter)
			source = fmp
		}
	}

	// Create the decoder and deccode into the destination.
	decoder, err := mapstructure.NewDecoder(unmarshalConfig.DecoderConfig)
	if err != nil {
		return fmt.Errorf("could not create decoder: %w", err)
	}
	return decoder.Decode(source)

}

// Allow using without a config just for environment variables into a struct.
func UnmarshalEnvStruct(path string, dest interface{}, tag string) error {
	decoderConfig := DefaultDecoderConfig()
	decoderConfig.TagName = tag
	return New().Unmarshal(dest, UnmarshalConf{Path: path, StructDefaults: true, StructEnvironment: true, DecoderConfig: decoderConfig})
}

// UnmarshalWithOpts unmarshals config into a data struct using config options.
func (c *Conf) UnmarshalWithOpts(dest interface{}, opts ...UnmarshalOption) error {
	var unmarshalConfig UnmarshalConf
	unmarshalConfig.DecoderConfig = DefaultDecoderConfig()
	for _, f := range opts {
		f(&unmarshalConfig)
	}
	return c.Unmarshal(dest, unmarshalConfig)
}

// UnmarshalOption are used to configure the Unmarshal behavior
type UnmarshalOption func(c *UnmarshalConf)

// WithPath sets the unmarshal path.
func WithPath(path string) UnmarshalOption {
	return func(c *UnmarshalConf) {
		c.Path = path
	}
}

// WithTag sets the unmarshal tag.
func WithTag(tag string) UnmarshalOption {
	return func(c *UnmarshalConf) {
		c.DecoderConfig.TagName = tag
	}
}

// WithFlatPath set unmarshaling to use flat path. See UnmarshalConf.
func WithFlatPaths(b bool) UnmarshalOption {
	return func(c *UnmarshalConf) {
		c.FlatPaths = b
	}
}

// WithStructDefaults sets the struct defaults. See UnmarshalConf.
func WithStructDefaults(b bool) UnmarshalOption {
	return func(c *UnmarshalConf) {
		c.StructDefaults = b
	}
}

// WithStructEnvironment sets the struct via environment. See UnmarshalConf.
func WithStructEnvironment(b bool) UnmarshalOption {
	return func(c *UnmarshalConf) {
		c.StructEnvironment = b
	}
}

// DecoderConfig is the decoder config used to decode into the struct.
func WithDecoderOpts(opts ...DecodeOption) UnmarshalOption {
	return func(c *UnmarshalConf) {
		for _, f := range opts {
			f(c.DecoderConfig)
		}
	}
}
