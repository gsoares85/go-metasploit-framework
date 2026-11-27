package core

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Datastore holds module option values with two scopes: a global scope (set with
// setg, shared by all modules) and a per-active-module overlay (set with set,
// which takes precedence). Only options registered in the schema can be set.
//
// A single module overlay is modeled here; a datastore per module instance is
// introduced later together with the module manager. The zero value is not
// usable; construct a Datastore with NewDatastore.
type Datastore struct {
	mu     sync.RWMutex
	schema map[string]Option // known options, keyed by normalized name
	global map[string]string // setg values
	module map[string]string // set values (active-module overlay)
}

// ResolvedOption is the effective state of an option, for rendering by the
// console (show options). It carries no internal references.
type ResolvedOption struct {
	Name        string
	Type        OptType
	Required    bool
	Description string
	Value       string // resolved value (or default); empty when unset
	Set         bool   // whether a value resolved, including a default
	Source      string // "module", "global", "default" or "unset"
}

// Source values reported by ResolvedOption.Source and by resolve.
const (
	SourceModule  = "module"
	SourceGlobal  = "global"
	SourceDefault = "default"
	SourceUnset   = "unset"
)

// NewDatastore returns an empty, ready-to-use Datastore.
func NewDatastore() *Datastore {
	return &Datastore{
		schema: make(map[string]Option),
		global: make(map[string]string),
		module: make(map[string]string),
	}
}

// Register adds options to the schema and returns an error (registering nothing)
// if any option is misconfigured: an enum with no allowed values, or a default
// that its own type would reject. Re-registering an option name replaces its
// definition.
//
// Register is a setup-time operation: all options are expected to be registered
// before the datastore is used concurrently or any value is set, and the schema
// is treated as immutable afterward. That contract is what lets value validation
// (which may run a caller-supplied validator) happen outside the datastore lock
// without a type-change race between validation and storage.
func (d *Datastore) Register(opts ...Option) error {
	normalized := make([]Option, 0, len(opts))
	for _, opt := range opts {
		if opt.Type == OptEnum && len(opt.Enum) == 0 {
			return fmt.Errorf("option %q: enum option requires a non-empty Enum list", opt.Name)
		}
		if opt.Default != nil {
			// Render the default once and validate that exact string through Check,
			// so defaults get the same length cap and redaction as values passed to
			// Set. Store the rendered string back on the option so later reads never
			// re-invoke a fmt.Stringer (whose output could change, or could run while
			// the datastore lock is held) and so Get/Set agree on the default.
			rendered := defaultToString(opt.Default)
			if err := opt.Check(rendered); err != nil {
				return fmt.Errorf("invalid default: %w", err)
			}
			opt.Default = rendered
		}
		normalized = append(normalized, opt)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, opt := range normalized {
		d.schema[normalizeKey(opt.Name)] = opt
	}
	return nil
}

// Set validates value and stores it in the active-module overlay. It returns
// ErrUnknownOption if the name is not registered, or a validation error (in
// which case nothing is stored).
func (d *Datastore) Set(name, value string) error {
	return d.set(name, value, false)
}

// SetGlobal validates value and stores it in the global scope.
func (d *Datastore) SetGlobal(name, value string) error {
	return d.set(name, value, true)
}

func (d *Datastore) set(name, value string, global bool) error {
	key := normalizeKey(name)

	// Look up the option under a read lock, then validate outside any lock.
	// Option.Validate is caller-supplied and must not run inside the critical
	// section: it could block every other reader and writer, or deadlock if it
	// re-enters the datastore (a write lock cannot be re-acquired as a read lock).
	d.mu.RLock()
	opt, ok := d.schema[key]
	d.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownOption, name)
	}
	if err := opt.Check(value); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	// Re-check under the write lock: the option may have been replaced since the
	// read above.
	if _, ok := d.schema[key]; !ok {
		return fmt.Errorf("%w: %q", ErrUnknownOption, name)
	}
	if global {
		d.global[key] = value
	} else {
		d.module[key] = value
	}
	return nil
}

// Unset removes the active-module overlay value for name, so Get falls back to
// the global value or the default.
func (d *Datastore) Unset(name string) {
	key := normalizeKey(name)
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.module, key)
}

// UnsetGlobal removes the global value for name.
func (d *Datastore) UnsetGlobal(name string) {
	key := normalizeKey(name)
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.global, key)
}

// Get resolves the effective value for name, trying the module overlay, then the
// global scope, then the registered default. The bool reports whether a value
// resolved (including from a default).
func (d *Datastore) Get(name string) (string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, _, ok := d.resolve(normalizeKey(name))
	return v, ok
}

// Options returns the resolved state of every registered option, sorted by name.
// The returned slice is a fresh copy and never exposes internal maps.
func (d *Datastore) Options() []ResolvedOption {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]ResolvedOption, 0, len(d.schema))
	for key, opt := range d.schema {
		v, source, ok := d.resolve(key)
		out = append(out, ResolvedOption{
			Name:        opt.Name,
			Type:        opt.Type,
			Required:    opt.Required,
			Description: opt.Description,
			Value:       v,
			Set:         ok,
			Source:      source,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Validate reports an error naming every required option that has no resolved
// value. It is meant to be called before running a module.
func (d *Datastore) Validate() error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var missing []string
	for key, opt := range d.schema {
		if !opt.Required {
			continue
		}
		if _, _, ok := d.resolve(key); !ok {
			missing = append(missing, opt.Name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return fmt.Errorf("%w: required option(s) not set: %s", ErrEmptyValue, strings.Join(missing, ", "))
}

// resolve returns the effective value for a normalized key and its source. It
// does not lock; callers must hold at least a read lock.
func (d *Datastore) resolve(key string) (value, source string, ok bool) {
	if v, found := d.module[key]; found {
		return v, SourceModule, true
	}
	if v, found := d.global[key]; found {
		return v, SourceGlobal, true
	}
	if opt, found := d.schema[key]; found && opt.Default != nil {
		return defaultToString(opt.Default), SourceDefault, true
	}
	return "", SourceUnset, false
}

// normalizeKey makes option lookups case-insensitive and whitespace-tolerant.
func normalizeKey(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

// defaultToString renders a registered default value as a string.
func defaultToString(def any) string {
	if s, ok := def.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", def)
}
