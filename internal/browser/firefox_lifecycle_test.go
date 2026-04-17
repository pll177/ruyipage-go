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
