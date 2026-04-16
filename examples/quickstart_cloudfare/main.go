package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

const (
	copilotQuestion = "你好，今天天气怎么样？"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("copilot.microsoft.com Cloudflare 测试")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Firefox 路径: %s\n", exampleutil.FixedFirefoxPath)

	fmt.Println("\n-> 访问 https://copilot.microsoft.com/ ...")
	if err := page.Navigate("https://copilot.microsoft.com/", "none"); err != nil {
		return err
	}
	page.Wait().Sleep(5 * time.Second)

	fmt.Println("-> 等待输入框...")
	inputBox, err := findInputBox(page)
	if err != nil {
		return err
	}

	if inputBox != nil {
		fmt.Println("-> 找到输入框，开始输入问题...")
		if err := inputBox.ClickSelf(false, 0); err == nil {
			page.Wait().Sleep(800 * time.Millisecond)
			if err := inputBox.Input(copilotQuestion, true, false); err == nil {
				page.Wait().Sleep(800 * time.Millisecond)
				sendButton, _ := page.Ele(`css:button[aria-label*="Send"]`, 1, time.Second)
				if sendButton == nil {
					sendButton, _ = page.Ele(`css:button[type="submit"]`, 1, time.Second)
				}
				if sendButton != nil {
					fmt.Println("-> 点击发送按钮...")
					_ = sendButton.ClickSelf(false, 0)
				} else {
					fmt.Println("-> 按 Enter 发送...")
					_ = page.Actions().Press(ruyipage.Keys.ENTER).Perform()
				}
				fmt.Println("-> 已发送问题，等待 Cloudflare 触发...")
				page.Wait().Sleep(15 * time.Second)
			}
		}
	} else {
		fmt.Println("-> 未找到输入框，直接等待 Cloudflare...")
		page.Wait().Sleep(5 * time.Second)
	}

	fmt.Println("\n-> 开始自动处理 Cloudflare...")
	passed := page.HandleCloudflareChallenge(120*time.Second, 2*time.Second)

	fmt.Println("\n" + strings.Repeat("=", 60))
	if passed {
		fmt.Println("✅ 成功通过 Cloudflare！")
		if err := printFullCookies(page); err != nil {
			return err
		}
	} else {
		fmt.Println("❌ 超时未通过")
	}

	exampleutil.PrintManualKeepOpen(500*time.Second, "与 Python quickstart 一致，保留页面供 Cloudflare 结果人工观察")
	page.Wait().Sleep(500 * time.Second)
	return nil
}

func findInputBox(page *ruyipage.FirefoxPage) (*ruyipage.FirefoxElement, error) {
	selectors := []string{
		"css:textarea",
		`css:[contenteditable="true"]`,
		"css:.input-area",
	}
	for count := 0; count < 30; count++ {
		for _, selector := range selectors {
			element, err := page.Ele(selector, 1, time.Second)
			if err != nil {
				return nil, err
			}
			if element != nil {
				return element, nil
			}
		}
		page.Wait().Sleep(time.Second)
	}
	return nil, nil
}

func printFullCookies(page *ruyipage.FirefoxPage) error {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Cloudflare / 页面 Cookie")
	fmt.Println(strings.Repeat("=", 60))

	rawCookie, err := page.RunJS(`return document.cookie`)
	if err != nil {
		return err
	}
	fmt.Printf("document.cookie: %v\n", rawCookie)

	cookies, err := page.Cookies(true)
	if err != nil {
		return err
	}
	fmt.Printf("Cookie 数量: %d\n", len(cookies))
	for index, cookie := range cookies {
		fmt.Printf("[%d] name=%s\n", index+1, cookie.Name)
		fmt.Printf("    value=%s\n", cookie.Value)
		fmt.Printf("    domain=%s\n", cookie.Domain)
		fmt.Printf("    path=%s\n", cookie.Path)
		fmt.Printf("    httpOnly=%v\n", cookie.HTTPOnly)
		fmt.Printf("    secure=%v\n", cookie.Secure)
		fmt.Printf("    sameSite=%s\n", cookie.SameSite)
		fmt.Printf("    expiry=%v\n", cookie.Expiry)
	}
	return nil
}
