// Package core provides the neutral orchestration layer of the framework:
// option configuration (the datastore), lifecycle and events. It must not depend
// on UI or infrastructure details; dependencies point inward toward this package.
package core

import "fmt"

// OptType is the declared type of a module option. It drives validation and how
// the console renders and parses the option's value.
type OptType int

const (
	// OptString is any non-empty text value.
	OptString OptType = iota
	// OptInt is a base-10 integer.
	OptInt
	// OptBool is a boolean (true/false/1/0/yes/no).
	OptBool
	// OptAddress is a single IP address or hostname.
	OptAddress
	// OptPort is a TCP/UDP port in the range 1-65535.
	OptPort
	// OptEnum is one of a fixed set of allowed values.
	OptEnum
	// OptPath is a filesystem path that must not contain traversal segments.
	OptPath
)

// String returns the lowercase name of the option type.
func (t OptType) String() string {
	switch t {
	case OptString:
		return "string"
	case OptInt:
		return "int"
	case OptBool:
		return "bool"
	case OptAddress:
		return "address"
	case OptPort:
		return "port"
	case OptEnum:
		return "enum"
	case OptPath:
		return "path"
	default:
		return "unknown"
	}
}

// Option describes a single configurable module option: its name, type, whether
// it must be set before running, a default value, a human-readable description,
// the allowed values for an enum, and an optional custom validator.
type Option struct {
	Name        string
	Type        OptType
	Required    bool
	Default     any
	Description string
	// Enum lists the valid values when Type is OptEnum.
	Enum []string
	// Validate, when non-nil, replaces the default validator for the type.
	Validate func(v string) error
}

// maxValueLen bounds the length of an option value so an untrusted caller cannot
// force disproportionate work or allocation during validation.
const maxValueLen = 8192

// Check validates a raw string value against the option. The error names the
// option but never includes the raw value, so sensitive values are not leaked
// into logs. A caller-supplied validator is not trusted to omit the value from
// its message, so its error is reduced to a classified sentinel (errors.Is still
// works) and its text is discarded.
func (o Option) Check(v string) error {
	if len(v) > maxValueLen {
		return fmt.Errorf("option %q: %w: value exceeds %d bytes", o.Name, ErrInvalidValue, maxValueLen)
	}
	if o.Validate != nil {
		if err := o.Validate(v); err != nil {
			return fmt.Errorf("option %q: %w", o.Name, classifyValidationError(err))
		}
		return nil
	}
	// Built-in validators are part of this package and never echo the raw value,
	// so their descriptive messages are safe to surface.
	if err := validatorFor(o.Type, o.Enum)(v); err != nil {
		return fmt.Errorf("option %q: %w", o.Name, err)
	}
	return nil
}
