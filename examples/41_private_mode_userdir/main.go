package main

import (
	"fmt"
	"strings"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const (
	privateModeTargetURL = "https://www.example.com"
	defaultUserDir       = `D:\ruyipage_userdir`
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	userDir := exampleutil.ResolveEnvPath("RUYIPAGE_EXAMPLE_USER_DIR", defaultUserDir)

	if err := exampleWithOptions(userDir); err != nil {
		return err
	}
	return exampleWithLaunchStyle(userDir)
}

func exampleWithOptions(userDir string) error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("示例41: private 模式 + user_dir")
	fmt.Println(strings.Repeat("=", 60))

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

	if err := page.Get(privateModeTargetURL); err != nil {
		return err
	}
	title, _ := page.Title()
	pageURL, _ := page.URL()
	fmt.Printf("标题: %s\n", title)
	fmt.Printf("地址: %s\n", pageURL)
	fmt.Printf("当前 user_dir: %s\n", userDir)
	fmt.Println("private 模式: 已启用")
	fmt.Printf("Firefox 路径: %s\n", exampleutil.FirefoxPath())
	return nil
}

func exampleWithLaunchStyle(userDir string) error {
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

	if err := page.Get(privateModeTargetURL); err != nil {
		return err
	}
	title, _ := page.Title()
	pageURL, _ := page.URL()
	fmt.Printf("[launch] 标题: %s\n", title)
	fmt.Printf("[launch] 地址: %s\n", pageURL)
	return nil
}
