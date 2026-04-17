package ruyipage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSanitizedFilesDoNotContainRedactedSecrets(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	root := filepath.Dir(file)

	files := []string{
		"README.md",
		"internal/config/firefox_options.go",
		"examples/internal/exampleutil/special_env.go",
		"examples/quickstart_fingerprint_browser/main.go",
		"docs/python_parity_audit.md",
	}
	patterns := []string{
		`C:\Users\` + "pll177",
		"us." + "ipwo.net",
		"pll177" + "_custom_zone_US",
	}

	for _, relPath := range files {
		content, err := os.ReadFile(filepath.Join(root, relPath))
		if err != nil {
			t.Fatalf("ReadFile(%q): %v", relPath, err)
		}
		text := string(content)
		for _, pattern := range patterns {
			if strings.Contains(text, pattern) {
				t.Fatalf("%s still contains forbidden pattern %q", relPath, pattern)
			}
		}
	}
}
