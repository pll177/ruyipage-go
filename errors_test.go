package ruyipage

import (
	"errors"
	"testing"

	internalpages "github.com/pll177/ruyipage-go/internal/pages"
	internalunits "github.com/pll177/ruyipage-go/internal/units"
)

func TestNavigateReturnsIncorrectURLError(t *testing.T) {
	page := &internalpages.FirefoxBase{}

	err := page.Navigate("not-a-url", "")
	if err == nil {
		t.Fatal("Navigate() expected IncorrectURLError for malformed URL")
	}

	var incorrectURL *IncorrectURLError
	if !errors.As(err, &incorrectURL) {
		t.Fatalf("Navigate() error = %T, want IncorrectURLError", err)
	}

	var baseErr *RuyiPageError
	if !errors.As(err, &baseErr) {
		t.Fatalf("Navigate() error = %T, want RuyiPageError ancestry", err)
	}
}

func TestInterceptedRequestReturnsNetworkInterceptError(t *testing.T) {
	req := &internalunits.InterceptedRequest{}

	err := req.ContinueRequest("", "", nil, nil)
	if err == nil {
		t.Fatal("ContinueRequest() expected NetworkInterceptError when driver is nil")
	}

	var interceptErr *NetworkInterceptError
	if !errors.As(err, &interceptErr) {
		t.Fatalf("ContinueRequest() error = %T, want NetworkInterceptError", err)
	}

	var baseErr *RuyiPageError
	if !errors.As(err, &baseErr) {
		t.Fatalf("ContinueRequest() error = %T, want RuyiPageError ancestry", err)
	}
}
