package core

import (
	"errors"
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
)

// Sentinel errors returned by option validation and by the datastore. Callers
// match them with errors.Is rather than comparing error strings.
var (
	// ErrUnknownOption is returned when an option name is not registered.
	ErrUnknownOption = errors.New("unknown option")
	// ErrInvalidValue is returned when a value fails its type's validation.
	ErrInvalidValue = errors.New("invalid value")
	// ErrEmptyValue is returned when a non-empty value is required but missing.
	ErrEmptyValue = errors.New("value is required")
)

// boolValues holds the string forms accepted by ValidateBool, compared
// case-insensitively.
var boolValues = map[string]struct{}{
	"true": {}, "false": {},
	"1": {}, "0": {},
	"yes": {}, "no": {},
}

// ValidateString accepts any non-empty value.
func ValidateString(v string) error {
	if v == "" {
		return ErrEmptyValue
	}
	return nil
}

// ValidateInt accepts a base-10 integer.
func ValidateInt(v string) error {
	if v == "" {
		return ErrEmptyValue
	}
	if _, err := strconv.Atoi(v); err != nil {
		return fmt.Errorf("%w: expected an integer", ErrInvalidValue)
	}
	return nil
}

// ValidateBool accepts true/false/1/0/yes/no, case-insensitively.
func ValidateBool(v string) error {
	if v == "" {
		return ErrEmptyValue
	}
	if _, ok := boolValues[strings.ToLower(v)]; !ok {
		return fmt.Errorf("%w: expected a boolean (true/false)", ErrInvalidValue)
	}
	return nil
}

// ValidatePort accepts an integer in the range 1-65535.
func ValidatePort(v string) error {
	if v == "" {
		return ErrEmptyValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("%w: port must be a number", ErrInvalidValue)
	}
	if n < 1 || n > 65535 {
		return fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalidValue)
	}
	return nil
}

// ValidateAddress accepts a single IP address or a syntactically valid hostname.
// CIDR ranges and host lists are intentionally out of scope here; scanner
// modules parse those separately.
func ValidateAddress(v string) error {
	if v == "" {
		return ErrEmptyValue
	}
	if net.ParseIP(v) != nil {
		return nil
	}
	// A dotted-decimal that is not a valid IP (e.g. "256.1.1.1") is a malformed
	// address, not a hostname, even though its labels are syntactically valid.
	if looksLikeIPv4(v) {
		return fmt.Errorf("%w: not a valid IPv4 address", ErrInvalidValue)
	}
	if isValidHostname(v) {
		return nil
	}
	return fmt.Errorf("%w: not a valid IP address or hostname", ErrInvalidValue)
}

// ValidatePath accepts a path that does not attempt directory traversal. It does
// not touch the filesystem; it only rejects obviously unsafe values (traversal
// segments and null bytes). Separators are normalized so the check is identical
// on Windows and POSIX.
func ValidatePath(v string) error {
	if v == "" {
		return ErrEmptyValue
	}
	if strings.ContainsRune(v, '\x00') {
		return fmt.Errorf("%w: path contains a null byte", ErrInvalidValue)
	}
	cleaned := path.Clean(strings.ReplaceAll(v, "\\", "/"))
	for _, seg := range strings.Split(cleaned, "/") {
		if seg == ".." {
			return fmt.Errorf("%w: path must not contain '..'", ErrInvalidValue)
		}
	}
	return nil
}

// ValidateEnum returns a validator that accepts only values present in allowed,
// compared case-insensitively.
func ValidateEnum(allowed []string) func(string) error {
	return func(v string) error {
		if v == "" {
			return ErrEmptyValue
		}
		for _, a := range allowed {
			if strings.EqualFold(v, a) {
				return nil
			}
		}
		return fmt.Errorf("%w: must be one of %s", ErrInvalidValue, strings.Join(allowed, ", "))
	}
}

// validatorFor returns the default validator for an option type. The enum list
// is only used for OptEnum.
func validatorFor(t OptType, enum []string) func(string) error {
	switch t {
	case OptInt:
		return ValidateInt
	case OptBool:
		return ValidateBool
	case OptPort:
		return ValidatePort
	case OptAddress:
		return ValidateAddress
	case OptPath:
		return ValidatePath
	case OptEnum:
		return ValidateEnum(enum)
	default: // OptString and any unknown type
		return ValidateString
	}
}

// isValidHostname reports whether v is a syntactically valid hostname per the
// RFC 1123 label rules. A single trailing dot (FQDN form) is allowed.
func isValidHostname(v string) bool {
	if len(v) > 253 {
		return false
	}
	host := strings.TrimSuffix(v, ".")
	if host == "" {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			switch {
			case c >= 'a' && c <= 'z',
				c >= 'A' && c <= 'Z',
				c >= '0' && c <= '9',
				c == '-':
			default:
				return false
			}
		}
	}
	return true
}

// looksLikeIPv4 reports whether v has the shape of a dotted-decimal IPv4 address
// (four all-numeric labels), regardless of whether each octet is in range.
func looksLikeIPv4(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for i := 0; i < len(p); i++ {
			if p[i] < '0' || p[i] > '9' {
				return false
			}
		}
	}
	return true
}
