package main

import (
	"fmt"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const (
	quickPrivateTargetURL = "https://www.example.com"
	quickPrivateUserDir   = `D:\ruyipage_userdir`
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	userDir := exampleutil.ResolveEnvPath("RUYIPAGE_EXAMPLE_USER_DIR", quickPrivateUserDir)
	if err := runWithOptions(userDir); err != nil {
		return err
	}
	return runWithLaunchStyle(userDir)
}

func runWithOptions(userDir string) error {
	opts := exampleutil.FixedVisibleOptions().
		WithUserDir(userDir).
		PrivateMode(true)

	page, err := ruyipage.NewFirefoxPage(opts)
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	if err := page.Get(quickPrivateTargetURL); err != nil {
		return err
	}
	title, _ := page.Title()
	pageURL, _ := page.URL()
	fmt.Printf("[options] title: %s\n", title)
	fmt.Printf("[options] url: %s\n", pageURL)
	return nil
}

func runWithLaunchStyle(userDir string) error {
	page, err := ruyipage.Launch(ruyipage.FirefoxQuickStartOptions{
		BrowserPath: exampleutil.FirefoxPath(),
		UserDir:     userDir,
		Private:     true,
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	if err := page.Get(quickPrivateTargetURL); err != nil {
		return err
	}
	title, _ := page.Title()
	pageURL, _ := page.URL()
	fmt.Printf("[launch] title: %s\n", title)
	fmt.Printf("[launch] url: %s\n", pageURL)
	return nil
}
