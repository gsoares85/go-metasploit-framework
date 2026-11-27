package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func newTestStore(t *testing.T) *Datastore {
	t.Helper()
	d := NewDatastore()
	if err := d.Register(
		Option{Name: "RPORT", Type: OptPort, Default: "80", Description: "target port"},
		Option{Name: "RHOSTS", Type: OptAddress, Required: true, Description: "target host"},
		Option{Name: "THREADS", Type: OptInt, Default: 10, Description: "worker count"},
	); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return d
}

func TestDatastoreRejectsInvalidDefault(t *testing.T) {
	d := NewDatastore()
	// Out-of-range port default.
	if err := d.Register(Option{Name: "RPORT", Type: OptPort, Default: 70000}); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Register invalid default = %v, want ErrInvalidValue", err)
	}
	if _, ok := d.Get("RPORT"); ok {
		t.Error("option was registered despite an invalid default")
	}
	// Empty default for a type that requires a value.
	if err := d.Register(Option{Name: "NAME", Type: OptString, Default: ""}); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("Register empty default = %v, want ErrEmptyValue", err)
	}
}

func TestDatastoreRejectsOversizeDefault(t *testing.T) {
	d := NewDatastore()
	// A default over the length cap must be rejected, the same as a value passed
	// to Set, so Get can never return a value that Set would reject.
	over := strings.Repeat("a", maxValueLen+1)
	if err := d.Register(Option{Name: "DATA", Type: OptString, Default: over}); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Register oversize default = %v, want ErrInvalidValue", err)
	}
	if _, ok := d.Get("DATA"); ok {
		t.Error("option registered despite an oversize default")
	}
	// A default exactly at the cap is accepted.
	atCap := strings.Repeat("a", maxValueLen)
	if err := d.Register(Option{Name: "DATA", Type: OptString, Default: atCap}); err != nil {
		t.Fatalf("Register default at cap = %v, want nil", err)
	}
	if v, ok := d.Get("DATA"); !ok || v != atCap {
		t.Errorf("DATA = (%q,%v), want the at-cap default", v[:min(len(v), 8)], ok)
	}
}

// mutableStringer returns a different string on each call, to prove the datastore
// renders a default exactly once (at registration) and stores that string.
type mutableStringer struct {
	calls *int
	vals  []string
}

func (m mutableStringer) String() string {
	i := *m.calls
	*m.calls++
	if i < len(m.vals) {
		return m.vals[i]
	}
	return m.vals[len(m.vals)-1]
}

func TestDatastoreRendersDefaultOnceAndStores(t *testing.T) {
	d := NewDatastore()
	calls := 0
	// First render is a valid port; a later render would be out of range. Only the
	// first must ever be observed.
	def := mutableStringer{calls: &calls, vals: []string{"80", "70000"}}
	if err := d.Register(Option{Name: "RPORT", Type: OptPort, Default: def}); err != nil {
		t.Fatalf("Register = %v, want nil", err)
	}
	if calls != 1 {
		t.Fatalf("default rendered %d times at registration, want 1", calls)
	}
	// Repeated reads return the stored string and never re-invoke String.
	for i := 0; i < 3; i++ {
		if v, ok := d.Get("RPORT"); !ok || v != "80" {
			t.Fatalf("Get #%d = (%q,%v), want (80,true)", i, v, ok)
		}
	}
	if calls != 1 {
		t.Errorf("String invoked %d times total, want 1 (default must be stored)", calls)
	}
}

func TestDatastoreRejectsEmptyEnum(t *testing.T) {
	d := NewDatastore()
	if err := d.Register(Option{Name: "ACTION", Type: OptEnum}); err == nil {
		t.Error("Register enum without allowed values = nil, want error")
	}
	if err := d.Register(Option{Name: "ACTION", Type: OptEnum, Enum: []string{"scan"}}); err != nil {
		t.Errorf("Register valid enum = %v, want nil", err)
	}
}

func TestDatastoreResolutionOrder(t *testing.T) {
	d := newTestStore(t)

	// Unset, no default -> not resolved.
	if v, ok := d.Get("RHOSTS"); ok {
		t.Errorf("RHOSTS unset resolved to %q, want not set", v)
	}
	// Default applies when nothing is set.
	if v, ok := d.Get("RPORT"); !ok || v != "80" {
		t.Errorf("RPORT = (%q,%v), want (80,true) from default", v, ok)
	}
	// A non-string default renders as a string.
	if v, ok := d.Get("THREADS"); !ok || v != "10" {
		t.Errorf("THREADS = (%q,%v), want (10,true)", v, ok)
	}
	// Global overrides default.
	if err := d.SetGlobal("RPORT", "8080"); err != nil {
		t.Fatal(err)
	}
	if v, _ := d.Get("RPORT"); v != "8080" {
		t.Errorf("RPORT after setg = %q, want 8080", v)
	}
	// Module overlay overrides global.
	if err := d.Set("RPORT", "9090"); err != nil {
		t.Fatal(err)
	}
	if v, _ := d.Get("RPORT"); v != "9090" {
		t.Errorf("RPORT after set = %q, want 9090", v)
	}
	// Unsetting the module overlay falls back to global.
	d.Unset("RPORT")
	if v, _ := d.Get("RPORT"); v != "8080" {
		t.Errorf("RPORT after unset = %q, want 8080 (global)", v)
	}
	// Unsetting the global falls back to the default.
	d.UnsetGlobal("RPORT")
	if v, _ := d.Get("RPORT"); v != "80" {
		t.Errorf("RPORT after unsetg = %q, want 80 (default)", v)
	}
}

func TestDatastoreUnknownOption(t *testing.T) {
	d := newTestStore(t)
	if err := d.Set("NOPE", "1"); !errors.Is(err, ErrUnknownOption) {
		t.Errorf("Set unknown = %v, want ErrUnknownOption", err)
	}
	if err := d.SetGlobal("NOPE", "1"); !errors.Is(err, ErrUnknownOption) {
		t.Errorf("SetGlobal unknown = %v, want ErrUnknownOption", err)
	}
}

func TestDatastoreInvalidValueNotStored(t *testing.T) {
	d := newTestStore(t)
	if err := d.Set("RPORT", "70000"); err == nil {
		t.Fatal("Set invalid port = nil, want error")
	}
	// The default must remain; the rejected value was not stored.
	if v, _ := d.Get("RPORT"); v != "80" {
		t.Errorf("RPORT after rejected set = %q, want 80 (default kept)", v)
	}
}

func TestDatastoreCaseInsensitive(t *testing.T) {
	d := newTestStore(t)
	if err := d.Set("rhosts", "1.2.3.4"); err != nil {
		t.Fatal(err)
	}
	if v, ok := d.Get("RHOSTS"); !ok || v != "1.2.3.4" {
		t.Errorf("case-insensitive Get = (%q,%v), want (1.2.3.4,true)", v, ok)
	}
}

func TestDatastoreValidateRequired(t *testing.T) {
	d := newTestStore(t)
	if err := d.Validate(); !errors.Is(err, ErrEmptyValue) {
		t.Fatalf("Validate with missing required = %v, want ErrEmptyValue", err)
	}
	if err := d.Set("RHOSTS", "10.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if err := d.Validate(); err != nil {
		t.Errorf("Validate after setting required = %v, want nil", err)
	}
}

func TestDatastoreOptionsSorted(t *testing.T) {
	d := newTestStore(t)
	opts := d.Options()
	if len(opts) != 3 {
		t.Fatalf("Options len = %d, want 3", len(opts))
	}
	wantOrder := []string{"RHOSTS", "RPORT", "THREADS"}
	for i, w := range wantOrder {
		if opts[i].Name != w {
			t.Errorf("Options[%d].Name = %q, want %q", i, opts[i].Name, w)
		}
	}
	for _, o := range opts {
		switch o.Name {
		case "RPORT":
			if o.Value != "80" || o.Source != "default" || !o.Set {
				t.Errorf("RPORT resolved = %+v, want value 80 / source default / set true", o)
			}
		case "RHOSTS":
			if o.Set || o.Source != "unset" {
				t.Errorf("RHOSTS resolved = %+v, want unset", o)
			}
		}
	}
}

func TestDatastoreRedactsCustomValidatorMessage(t *testing.T) {
	const secret = "s3cr3t-token"
	d := NewDatastore()
	if err := d.Register(Option{
		Name: "APIKEY",
		Type: OptString,
		// A hostile/naive validator that echoes the raw value into its message.
		Validate: func(v string) error {
			return fmt.Errorf("%w: rejected %q", ErrInvalidValue, v)
		},
	}); err != nil {
		t.Fatal(err)
	}
	err := d.Set("APIKEY", secret)
	if err == nil {
		t.Fatal("Set = nil, want error")
	}
	if !errors.Is(err, ErrInvalidValue) {
		t.Errorf("err = %v, want errors.Is ErrInvalidValue", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("error message leaked the raw value: %q", err.Error())
	}
	// The option name is still reported for diagnostics.
	if !strings.Contains(err.Error(), "APIKEY") {
		t.Errorf("error message = %q, want it to name the option", err.Error())
	}
}

func TestDatastoreConcurrentAccess(t *testing.T) {
	d := newTestStore(t)
	const workers = 50
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			_ = d.Set("RPORT", strconv.Itoa(1+i%65535))
			_ = d.SetGlobal("RHOSTS", "127.0.0.1")
			_, _ = d.Get("RPORT")
			_ = d.Options()
			_ = d.Validate()
			d.Unset("RPORT")
		}(i)
	}
	wg.Wait()
}

func TestDatastoreSetRunsCustomValidatorWithoutDeadlock(t *testing.T) {
	d := newTestStore(t)
	called := false
	if err := d.Register(Option{
		Name: "CB",
		Type: OptString,
		Validate: func(v string) error {
			called = true
			// A re-entrant read must not deadlock: validation runs outside the
			// datastore lock.
			_, _ = d.Get("RPORT")
			if v == "bad" {
				return ErrInvalidValue
			}
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("CB", "ok"); err != nil {
		t.Fatalf("Set with custom validator = %v, want nil", err)
	}
	if !called {
		t.Error("custom validator was not invoked by Set")
	}
	if err := d.Set("CB", "bad"); err == nil {
		t.Error("Set with failing custom validator = nil, want error")
	}
	if v, _ := d.Get("CB"); v != "ok" {
		t.Errorf("CB = %q, want ok (rejected value not stored)", v)
	}
}
