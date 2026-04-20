package base

import "testing"

func TestExtractEventContextFallsBackToNestedRequestContext(t *testing.T) {
	params := map[string]any{
		"request": map[string]any{
			"context": "request-context",
		},
	}

	if got := extractEventContext(params); got != "request-context" {
		t.Fatalf("extractEventContext(request) = %q, want %q", got, "request-context")
	}
}

func TestExtractEventContextFallsBackToNestedSourceContext(t *testing.T) {
	params := map[string]any{
		"source": map[string]any{
			"context": "source-context",
		},
	}

	if got := extractEventContext(params); got != "source-context" {
		t.Fatalf("extractEventContext(source) = %q, want %q", got, "source-context")
	}
}
