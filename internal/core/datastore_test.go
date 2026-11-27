package core

import (
	"errors"
	"strconv"
	"sync"
	"testing"
)

func newTestStore() *Datastore {
	d := NewDatastore()
	d.Register(
		Option{Name: "RPORT", Type: OptPort, Default: "80", Description: "target port"},
		Option{Name: "RHOSTS", Type: OptAddress, Required: true, Description: "target host"},
		Option{Name: "THREADS", Type: OptInt, Default: 10, Description: "worker count"},
	)
	return d
}

func TestDatastoreResolutionOrder(t *testing.T) {
	d := newTestStore()

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
	d := newTestStore()
	if err := d.Set("NOPE", "1"); !errors.Is(err, ErrUnknownOption) {
		t.Errorf("Set unknown = %v, want ErrUnknownOption", err)
	}
	if err := d.SetGlobal("NOPE", "1"); !errors.Is(err, ErrUnknownOption) {
		t.Errorf("SetGlobal unknown = %v, want ErrUnknownOption", err)
	}
}

func TestDatastoreInvalidValueNotStored(t *testing.T) {
	d := newTestStore()
	if err := d.Set("RPORT", "70000"); err == nil {
		t.Fatal("Set invalid port = nil, want error")
	}
	// The default must remain; the rejected value was not stored.
	if v, _ := d.Get("RPORT"); v != "80" {
		t.Errorf("RPORT after rejected set = %q, want 80 (default kept)", v)
	}
}

func TestDatastoreCaseInsensitive(t *testing.T) {
	d := newTestStore()
	if err := d.Set("rhosts", "1.2.3.4"); err != nil {
		t.Fatal(err)
	}
	if v, ok := d.Get("RHOSTS"); !ok || v != "1.2.3.4" {
		t.Errorf("case-insensitive Get = (%q,%v), want (1.2.3.4,true)", v, ok)
	}
}

func TestDatastoreValidateRequired(t *testing.T) {
	d := newTestStore()
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
	d := newTestStore()
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

func TestDatastoreConcurrentAccess(t *testing.T) {
	d := newTestStore()
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
