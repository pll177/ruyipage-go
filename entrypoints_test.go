package ruyipage

import (
	"testing"

	"github.com/pll177/ruyipage-go/internal/support"
)

func TestLaunchUsesDefaultQuickStartOptions(t *testing.T) {
	var captured *FirefoxOptions
	original := newFirefoxPageForEntry
	t.Cleanup(func() {
		newFirefoxPageForEntry = original
	})
	newFirefoxPageForEntry = func(addrOrOpts any) (*FirefoxPage, error) {
		opts, ok := addrOrOpts.(*FirefoxOptions)
		if !ok {
			t.Fatalf("Launch() passed unexpected input type %T", addrOrOpts)
		}
		captured = opts.Clone()
		return &FirefoxPage{}, nil
	}

	if _, err := Launch(); err != nil {
		t.Fatalf("Launch() returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("Launch() did not pass options to NewFirefoxPage")
	}
	if captured.Port() != support.DefaultPort {
		t.Fatalf("Launch() port = %d, want %d", captured.Port(), support.DefaultPort)
	}
	if captured.BrowserPath() != `C:\Program Files\Mozilla Firefox\firefox.exe` {
		t.Fatalf("Launch() browser path = %q", captured.BrowserPath())
	}
	if !containsString(captured.Arguments(), "--width=1280") {
		t.Fatalf("Launch() arguments missing default width: %#v", captured.Arguments())
	}
	if !containsString(captured.Arguments(), "--height=800") {
		t.Fatalf("Launch() arguments missing default height: %#v", captured.Arguments())
	}
}

func TestLaunchAppliesProvidedQuickStartOptions(t *testing.T) {
	var captured *FirefoxOptions
	original := newFirefoxPageForEntry
	t.Cleanup(func() {
		newFirefoxPageForEntry = original
	})
	newFirefoxPageForEntry = func(addrOrOpts any) (*FirefoxPage, error) {
		opts, ok := addrOrOpts.(*FirefoxOptions)
		if !ok {
			t.Fatalf("Launch() passed unexpected input type %T", addrOrOpts)
		}
		captured = opts.Clone()
		return &FirefoxPage{}, nil
	}

	_, err := Launch(FirefoxQuickStartOptions{
		Port:            9333,
		BrowserPath:     `D:\Firefox\firefox.exe`,
		UserDir:         `D:\ruyipage_userdir`,
		Private:         true,
		Headless:        true,
		XPathPicker:     true,
		ActionVisual:    true,
		WindowWidth:     1440,
		WindowHeight:    900,
		TimeoutBase:     1,
		TimeoutPageLoad: 2,
		TimeoutScript:   3,
	})
	if err != nil {
		t.Fatalf("Launch(options) returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("Launch(options) did not pass options to NewFirefoxPage")
	}
	if captured.Port() != 9333 {
		t.Fatalf("Launch(options) port = %d, want 9333", captured.Port())
	}
	if captured.BrowserPath() != `D:\Firefox\firefox.exe` {
		t.Fatalf("Launch(options) browser path = %q", captured.BrowserPath())
	}
	if captured.UserDir() != `D:\ruyipage_userdir` {
		t.Fatalf("Launch(options) user dir = %q", captured.UserDir())
	}
	if !captured.IsPrivateMode() || !captured.IsHeadless() {
		t.Fatalf("Launch(options) did not apply private/headless flags: private=%v headless=%v", captured.IsPrivateMode(), captured.IsHeadless())
	}
	if !captured.IsXPathPickerEnabled() || !captured.IsActionVisualEnabled() {
		t.Fatalf("Launch(options) did not apply picker/action visual flags")
	}
	if !containsString(captured.Arguments(), "--width=1440") || !containsString(captured.Arguments(), "--height=900") {
		t.Fatalf("Launch(options) window args = %#v", captured.Arguments())
	}
	timeouts := captured.Timeouts()
	if timeouts.Base != 1 || timeouts.PageLoad != 2 || timeouts.Script != 3 {
		t.Fatalf("Launch(options) timeouts = %#v", timeouts)
	}
}

func TestLaunchRejectsMultipleQuickStartOptions(t *testing.T) {
	called := false
	original := newFirefoxPageForEntry
	t.Cleanup(func() {
		newFirefoxPageForEntry = original
	})
	newFirefoxPageForEntry = func(addrOrOpts any) (*FirefoxPage, error) {
		called = true
		return &FirefoxPage{}, nil
	}

	_, err := Launch(FirefoxQuickStartOptions{}, FirefoxQuickStartOptions{})
	if err == nil {
		t.Fatal("Launch() expected error when passed multiple options")
	}
	if called {
		t.Fatal("Launch() should not construct page when option arity is invalid")
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
