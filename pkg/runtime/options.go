// Copyright 2026 Dipjyoti Metia.
// Portions copyright 2025 Redpanda Data, Inc. (Option/ExtraProperty pattern
// adapted from redpanda-data/protoc-gen-go-mcp, Apache-2.0).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

package runtime

import (
	"encoding/json"
)

// Option configures tool registration.
type Option func(*Config)

// Config is the resolved set of registration options. Generated code creates
// one with NewConfig, applies user Options to it, then passes it to
// ApplyConfig for each tool.
type Config struct {
	ExtraProperties []ExtraProperty
	NamePrefix      string
}

// ExtraProperty defines an additional schema property to add to every tool's
// input schema. The decoded value is placed on the request context via
// ContextKey so handlers can read it.
type ExtraProperty struct {
	Name        string
	Description string
	Required    bool
	ContextKey  any
}

// NewConfig returns a Config with default values.
func NewConfig() *Config { return &Config{} }

// WithNamePrefix prepends "<prefix>_" to every tool name at registration time.
// Useful when the same service is registered multiple times under different
// names (e.g. two instances of the same API behind different base URLs).
func WithNamePrefix(prefix string) Option {
	return func(c *Config) {
		c.NamePrefix = prefix
	}
}

// WithExtraProperties adds extra string-typed properties to every tool schema.
// At call time the values are extracted from request arguments and placed on
// the handler's context.
func WithExtraProperties(properties ...ExtraProperty) Option {
	return func(c *Config) {
		c.ExtraProperties = append(c.ExtraProperties, properties...)
	}
}

// ApplyConfig applies prefix and extra-property transformations to a tool and
// returns the modified copy.
func ApplyConfig(tool Tool, cfg *Config) Tool {
	if cfg == nil {
		return tool
	}
	if cfg.NamePrefix != "" {
		tool.Name = cfg.NamePrefix + "_" + tool.Name
	}
	if len(cfg.ExtraProperties) > 0 {
		tool = AddExtraPropertiesToTool(tool, cfg.ExtraProperties)
	}
	return tool
}

// AddExtraPropertiesToTool modifies a tool's schema to include the given
// extra properties. Each extra property is added as a string-typed field on
// the root object.
func AddExtraPropertiesToTool(tool Tool, properties []ExtraProperty) Tool {
	if len(properties) == 0 {
		return tool
	}

	var schema map[string]any
	if err := json.Unmarshal(tool.RawInputSchema, &schema); err != nil {
		return tool
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		props = make(map[string]any)
		schema["properties"] = props
	}

	var required []any
	if r, ok := schema["required"].([]any); ok {
		required = r
	}

	for _, p := range properties {
		props[p.Name] = map[string]any{
			"type":        "string",
			"description": p.Description,
		}
		if p.Required {
			required = append(required, p.Name)
		}
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	modified, err := json.Marshal(schema)
	if err != nil {
		return tool
	}
	out := tool
	out.RawInputSchema = modified
	return out
}

// ExtractExtraProperty pulls the value of an extra property out of args
// (removing it so generated body/path decoders don't see it). Returns the
// string value and a bool indicating whether the property was present.
func ExtractExtraProperty(args map[string]any, name string) (string, bool) {
	v, ok := args[name]
	if !ok {
		return "", false
	}
	delete(args, name)
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}
