package core

import (
	"errors"
	"strings"
	"testing"
)

func TestOptTypeString(t *testing.T) {
	cases := map[OptType]string{
		OptString:   "string",
		OptInt:      "int",
		OptBool:     "bool",
		OptAddress:  "address",
		OptPort:     "port",
		OptEnum:     "enum",
		OptPath:     "path",
		OptType(99): "unknown",
	}
	for typ, want := range cases {
		if got := typ.String(); got != want {
			t.Errorf("OptType(%d).String() = %q, want %q", typ, got, want)
		}
	}
}

func TestValidatePort(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"low bound", "1", false},
		{"high bound", "65535", false},
		{"mid", "8080", false},
		{"zero", "0", true},
		{"too high", "65536", true},
		{"negative", "-1", true},
		{"not a number", "abc", true},
		{"empty", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidatePort(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidatePort(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
			}
			if err != nil {
				want := ErrInvalidValue
				if c.in == "" {
					want = ErrEmptyValue
				}
				if !errors.Is(err, want) {
					t.Errorf("ValidatePort(%q): error %v is not %v", c.in, err, want)
				}
			}
		})
	}
}

func TestValidateInt(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"42", false}, {"-5", false}, {"0", false},
		{"3.5", true}, {"x", true}, {"", true},
	}
	for _, c := range cases {
		if err := ValidateInt(c.in); (err != nil) != c.wantErr {
			t.Errorf("ValidateInt(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
		}
	}
}

func TestValidateBool(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"true", false}, {"false", false}, {"1", false}, {"0", false},
		{"yes", false}, {"no", false}, {"TRUE", false}, {"No", false},
		{"maybe", true}, {"2", true}, {"", true},
	}
	for _, c := range cases {
		if err := ValidateBool(c.in); (err != nil) != c.wantErr {
			t.Errorf("ValidateBool(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
		}
	}
}

func TestValidateAddress(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"127.0.0.1", false},
		{"::1", false},
		{"2001:db8::1", false},
		{"example.com", false},
		{"example.com.", false},
		{"localhost", false},
		{"", true},
		{"256.1.1.1", true},
		{"-bad.com", true},
		{"a..b", true},
		{"has space", true},
	}
	for _, c := range cases {
		if err := ValidateAddress(c.in); (err != nil) != c.wantErr {
			t.Errorf("ValidateAddress(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
		}
	}
}

func TestValidatePath(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"simple", "config/app.conf", false},
		{"collapses safely", "a/./b", false},
		{"inner backref stays in bounds", "a/../b", false},
		{"traversal", "../etc/passwd", true},
		{"deep traversal", "a/../../b", true},
		{"windows traversal", `..\\windows`, true},
		{"null byte", "a\x00b", true},
		{"empty", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := ValidatePath(c.in); (err != nil) != c.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	validate := ValidateEnum([]string{"low", "high"})
	for _, in := range []string{"low", "high", "HIGH", "Low"} {
		if err := validate(in); err != nil {
			t.Errorf("ValidateEnum accepted set: %q returned %v", in, err)
		}
	}
	for _, in := range []string{"mid", "", "lowest"} {
		if err := validate(in); err == nil {
			t.Errorf("ValidateEnum(%q) = nil, want error", in)
		}
	}
}

func TestValidateString(t *testing.T) {
	if err := ValidateString("x"); err != nil {
		t.Errorf("ValidateString(non-empty) = %v, want nil", err)
	}
	if err := ValidateString(""); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("ValidateString(empty) = %v, want ErrEmptyValue", err)
	}
}

func TestOptionCheck(t *testing.T) {
	// Default validator for the type.
	port := Option{Name: "RPORT", Type: OptPort}
	if err := port.Check("80"); err != nil {
		t.Errorf("Check valid port = %v, want nil", err)
	}
	err := port.Check("0")
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Check invalid port = %v, want ErrInvalidValue", err)
	}
	if !strings.Contains(err.Error(), "RPORT") {
		t.Errorf("Check error %q should name the option", err)
	}

	// Enum option uses its allowed list.
	enum := Option{Name: "ACTION", Type: OptEnum, Enum: []string{"scan", "list"}}
	if err := enum.Check("SCAN"); err != nil {
		t.Errorf("Check enum member = %v, want nil", err)
	}
	if err := enum.Check("delete"); err == nil {
		t.Error("Check enum non-member = nil, want error")
	}

	// Custom validator replaces the default one.
	custom := Option{
		Name: "X",
		Type: OptString,
		Validate: func(v string) error {
			if v == "bad" {
				return ErrInvalidValue
			}
			return nil
		},
	}
	if err := custom.Check("ok"); err != nil {
		t.Errorf("custom Check(ok) = %v, want nil", err)
	}
	if err := custom.Check("bad"); err == nil {
		t.Error("custom Check(bad) = nil, want error")
	}
}
