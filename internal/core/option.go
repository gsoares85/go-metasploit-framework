// Package core provides the neutral orchestration layer of the framework:
// option configuration (the datastore), lifecycle and events. It must not depend
// on UI or infrastructure details; dependencies point inward toward this package.
package core

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
