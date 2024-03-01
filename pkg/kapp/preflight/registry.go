// Copyright 2024 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package preflight

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	ctldgraph "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diffgraph"
)

const preflightFlag = "preflight"

// Registry is a collection of preflight checks and associated configuration
type Registry struct {
	known  map[string]Check
	config map[string]any
}

// NewRegistry will return a new *Registry with the
// provided set of preflight checks added to the registry
func NewRegistry(checks map[string]Check) *Registry {
	registry := &Registry{}
	for name, check := range checks {
		registry.AddCheck(name, check)
	}
	return registry
}

// String returns a string representation of the
// preflight checks. It follows the format:
// CheckName={true||false},...
// This method is needed so Registry implements
// the pflag.Value interface
func (c *Registry) String() string {
	defaults := []string{}
	for k, v := range c.known {
		defaults = append(defaults, fmt.Sprintf("%s=%v", k, v.Enabled()))
	}
	return strings.Join(defaults, ",")
}

// Type returns a string representing the type
// of the Registry. It is needed to implement the
// pflag.Value interface
func (c *Registry) Type() string {
	return fmt.Sprintf("%T", c)
}

// Set takes in a string in the format of
// CheckName={true||false},...
// and sets the specified preflight check
// as enabled if true, disabled if false
// Returns an error if there is a problem
// parsing the preflight checks
func (c *Registry) Set(s string) error {
	if c.known == nil {
		return nil
	}

	fmt.Printf("JSON input: %q\n", s)
	if !json.Valid([]byte(s)) {
		return fmt.Errorf("Invalid JSON format: %s", s)
	}

	err := json.Unmarshal([]byte(s), &c.config)
	if err != nil {
		return err
	}

	for name, values := range c.config {
		_, ok := c.known[name]
		if !ok {
			return fmt.Errorf("unknown preflight check %q specified", name)
		}
		config := values.(map[string]any)
		if config == nil {
			return fmt.Errorf("unable to parse config %v", values)
		}
		// Look for enabled
		enableStr, ok := config["enabled"].(string)
		if ok {
			enabled, err := strconv.ParseBool(enableStr)
			if err != nil {
				return fmt.Errorf("unable to parse boolean representation of %q: %w", enableStr, err)
			}
			c.known[name].SetEnabled(enabled)
		}
		// Give the check it's config
		err = c.known[name].SetConfig(config)
		if err != nil {
			return fmt.Error("unable to parse config for %q: %w", name, err)
		}
	}
	return nil
}

// AddFlags adds the --preflight flag to a
// pflag.FlagSet and configures the preflight
// checks in the registry based on the user provided
// values. If no values are provided by a user the
// default values are used.
func (c *Registry) AddFlags(flags *pflag.FlagSet) {
	knownChecks := []string{}
	for name := range c.known {
		knownChecks = append(knownChecks, name)
	}
	flags.Var(c, preflightFlag, fmt.Sprintf("preflight checks to run. Available preflight checks are [%s]", strings.Join(knownChecks, ",")))
}

// AddCheck adds a new preflight check to the registry.
// The name provided will map to the provided Check.
func (c *Registry) AddCheck(name string, check Check) {
	if c.known == nil {
		c.known = make(map[string]Check)
	}
	c.known[name] = check
}

// Run will execute any enabled preflight checks. The provided
// Context and ChangeGraph will be passed to the preflight checks
// that are being executed.
func (c *Registry) Run(ctx context.Context, cg *ctldgraph.ChangeGraph) error {
	for name, check := range c.known {
		if check.Enabled() {
			err := check.Run(ctx, cg)
			if err != nil {
				return fmt.Errorf("running preflight check %q: %w", name, err)
			}
		}
	}
	return nil
}
