package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

const xpathPickerAutoPortStart = 19322

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	testURL, err := exampleutil.TestPageURL("xpath_picker_complex_showcase.html")
	if err != nil {
		return err
	}

	userDir, err := os.MkdirTemp("", "ruyipage-example42-*")
	if err != nil {
		return err
	}
	cleanupUserDir := true
	defer func() {
		if cleanupUserDir {
			_ = os.RemoveAll(userDir)
		}
	}()

	options := exampleutil.FixedVisibleOptions().
		EnableXPathPicker(true).
		WithWindowSize(1600, 1100).
		WithUserDir(userDir).
		EnableAutoPort(true).
		WithAutoPortStart(xpathPickerAutoPortStart)

	page, err := ruyipage.NewFirefoxPage(
		options,
	)
	if err != nil {
		return err
	}
	cleanupPage := true
	defer func() {
		if cleanupPage {
			_ = page.Quit(5*time.Second, false)
		}
	}()

	if err := page.Get(testURL); err != nil {
		return err
	}

	browser := page.Browser()
	debugAddress := options.Address()
	sessionID := ""
	profilePath := userDir
	if browser != nil {
		if browser.Address() != "" {
			debugAddress = browser.Address()
		}
		sessionID = browser.SessionID()
		if browserOptions := browser.Options(); browserOptions != nil {
			if browserOptions.UserDir() != "" {
				profilePath = browserOptions.UserDir()
			} else if browserOptions.ProfilePath() != "" {
				profilePath = browserOptions.ProfilePath()
			}
		}
	}

	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("示例42: XPath Picker 综合复杂场景展示页")
	fmt.Printf("Firefox 路径: %s\n", exampleutil.FixedFirefoxPath)
	fmt.Printf("本次独立调试地址: %s\n", debugAddress)
	fmt.Printf("本次 session id: %s\n", sessionID)
	fmt.Printf("本次 user_dir/profile: %s\n", profilePath)
	fmt.Printf("页面地址: %s\n", testURL)
	fmt.Println("本次运行会启动独立浏览器实例，请手动点选页面中的复杂节点进行测试。")
	fmt.Println("可重点测试：主页面 shadow、outer iframe、inner iframe、SVG、contenteditable。")
	fmt.Println("按 Ctrl+C 结束脚本；本示例会主动关闭本次浏览器实例并清理临时 user_dir。")
	fmt.Println(strings.Repeat("=", 72))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	signal.Stop(stop)

	fmt.Println("\n收到 Ctrl+C，正在关闭本次独立浏览器实例...")
	if err := page.Quit(5*time.Second, false); err != nil {
		return err
	}
	cleanupPage = false
	if err := os.RemoveAll(userDir); err != nil {
		return err
	}
	cleanupUserDir = false
	fmt.Println("示例42结束。本次独立浏览器实例和临时 user_dir 已清理。")
	return nil
}
