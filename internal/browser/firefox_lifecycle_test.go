package browser

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pll177/ruyipage-go/internal/config"
)

func TestFirefoxQuitCleansManagedFPFile(t *testing.T) {
	fpfilePath := filepath.Join(t.TempDir(), "managed.txt")
	if err := os.WriteFile(fpfilePath, []byte("webdriver:0\n"), 0o644); err != nil {
		t.Fatalf("write managed fpfile: %v", err)
	}

	opts := config.NewFirefoxOptions()
	opts.SetManagedFPFile(fpfilePath)
	firefox := newFirefoxInstance(opts)

	if err := firefox.Quit(0, false); err != nil {
		t.Fatalf("Quit returned error: %v", err)
	}
	if _, err := os.Stat(fpfilePath); !os.IsNotExist(err) {
		t.Fatalf("expected managed fpfile to be deleted, stat err = %v", err)
	}
}

func TestHandleManagedProcessExitCleansManagedFPFile(t *testing.T) {
	fpfilePath := filepath.Join(t.TempDir(), "managed.txt")
	if err := os.WriteFile(fpfilePath, []byte("webdriver:0\n"), 0o644); err != nil {
		t.Fatalf("write managed fpfile: %v", err)
	}

	opts := config.NewFirefoxOptions()
	opts.SetManagedFPFile(fpfilePath)
	firefox := newFirefoxInstance(opts)
	process := &exec.Cmd{}

	firefox.mu.Lock()
	firefox.process = process
	firefox.mu.Unlock()

	firefox.handleManagedProcessExit(process)

	if _, err := os.Stat(fpfilePath); !os.IsNotExist(err) {
		t.Fatalf("expected managed fpfile to be deleted, stat err = %v", err)
	}
}

func TestFirefoxQuitDoesNotDeleteManualFPFile(t *testing.T) {
	fpfilePath := filepath.Join(t.TempDir(), "manual.txt")
	if err := os.WriteFile(fpfilePath, []byte("webdriver:0\n"), 0o644); err != nil {
		t.Fatalf("write manual fpfile: %v", err)
	}

	opts := config.NewFirefoxOptions().WithFPFile(fpfilePath)
	firefox := newFirefoxInstance(opts)

	if err := firefox.Quit(0, false); err != nil {
		t.Fatalf("Quit returned error: %v", err)
	}
	if _, err := os.Stat(fpfilePath); err != nil {
		t.Fatalf("expected manual fpfile to remain, stat err = %v", err)
	}
}

func TestFirefoxAutoPortRegistryReserveReleaseReuse(t *testing.T) {
	registry := newFirefoxAutoPortRegistry()

	address1, _, err := registry.reserve("127.0.0.1", 9400)
	if err != nil {
		t.Fatalf("reserve first address: %v", err)
	}
	address2, _, err := registry.reserve("127.0.0.1", 9400)
	if err != nil {
		t.Fatalf("reserve second address: %v", err)
	}
	if address1 == address2 {
		t.Fatalf("expected unique reserved addresses, got %s", address1)
	}

	registry.activate(address1)
	registry.release(address1)

	address3, _, err := registry.reserve("127.0.0.1", 9400)
	if err != nil {
		t.Fatalf("reserve reused address: %v", err)
	}
	if address3 != address1 {
		t.Fatalf("expected released address %s to become reusable, got %s", address1, address3)
	}
}

func TestRequiresAutoPortReservationSkipsExplicitAddress(t *testing.T) {
	autoOpts := config.NewFirefoxOptions().AutoPortEnabled(true)
	if !requiresAutoPortReservation(autoOpts) {
		t.Fatal("expected auto port reservation for implicit address")
	}

	explicitOpts := config.NewFirefoxOptions().
		AutoPortEnabled(true).
		WithAddress("127.0.0.1:9555")
	if requiresAutoPortReservation(explicitOpts) {
		t.Fatal("expected explicit address to bypass auto port reservation")
	}
}

func TestFirefoxQuitReleasesAutoPortLease(t *testing.T) {
	previousCoordinator := firefoxAutoPortCoordinator
	firefoxAutoPortCoordinator = newFirefoxAutoPortRegistry()
	defer func() {
		firefoxAutoPortCoordinator = previousCoordinator
	}()

	opts := config.NewFirefoxOptions().AutoPortEnabled(true)
	firefox := newFirefoxInstance(opts)
	firefox.address = "127.0.0.1:9600"
	firefox.autoPortLease = firefox.address
	firefoxAutoPortCoordinator.activate(firefox.address)

	if err := firefox.Quit(0, false); err != nil {
		t.Fatalf("Quit returned error: %v", err)
	}

	if _, exists := firefoxAutoPortCoordinator.active[firefox.address]; exists {
		t.Fatalf("expected auto port lease %s to be released", firefox.address)
	}
}

func TestPrepareNextLaunchAttemptRebindsReservedAddress(t *testing.T) {
	previousCoordinator := firefoxAutoPortCoordinator
	previousRegistry := firefoxRegistry
	firefoxAutoPortCoordinator = newFirefoxAutoPortRegistry()
	firefoxRegistry = make(map[string]*Firefox)
	defer func() {
		firefoxAutoPortCoordinator = previousCoordinator
		firefoxRegistry = previousRegistry
	}()

	opts := config.NewFirefoxOptions().AutoPortEnabled(true)
	initialAddress, initialPort, err := firefoxAutoPortCoordinator.reserve(opts.Host(), 9700)
	if err != nil {
		t.Fatalf("reserve initial address: %v", err)
	}
	opts.SetResolvedPort(initialPort)
	firefox := newFirefoxInstance(opts)
	firefox.address = initialAddress
	firefox.autoPortLease = initialAddress
	firefoxRegistry[initialAddress] = firefox

	if err := firefox.prepareNextLaunchAttempt(); err != nil {
		t.Fatalf("prepareNextLaunchAttempt returned error: %v", err)
	}

	if firefox.Address() == initialAddress {
		t.Fatalf("expected address to change after prepareNextLaunchAttempt")
	}
	if firefox.autoPortLease != firefox.Address() {
		t.Fatalf("expected auto port lease to track rebound address, got lease=%s address=%s", firefox.autoPortLease, firefox.Address())
	}
	if _, exists := firefoxAutoPortCoordinator.reserved[initialAddress]; exists {
		t.Fatalf("expected old reservation %s to be released", initialAddress)
	}
	if _, exists := firefoxAutoPortCoordinator.reserved[firefox.Address()]; !exists {
		t.Fatalf("expected new reservation %s to exist", firefox.Address())
	}
}
