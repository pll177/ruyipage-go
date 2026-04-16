package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试17: 用户提示框处理")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("说明: Go 版保留自动策略 + 步骤式 prompt 登录，同样展示 mouse / keyboard 两种触发方式。")

	options := exampleutil.VisibleOptions().WithUserPromptHandler(map[string]string{
		"alert":   "accept",
		"confirm": "accept",
		"prompt":  "ignore",
		"default": "accept",
	})
	page, err := ruyipage.NewFirefoxPage(options)
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(time.Second)
		_ = page.Quit(0, false)
	}()

	testURL, err := exampleutil.TestPageURL("native_user_prompts_test.html")
	if err != nil {
		return err
	}
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(800 * time.Millisecond)

	fmt.Println("\n1. alert（自动策略）:")
	if err := mustClick(page, "#alert-btn"); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	alertResult, err := mustText(page, "#alert-result")
	if err != nil {
		return err
	}
	fmt.Printf("   result: %s\n", alertResult)

	fmt.Println("\n2. confirm（自动策略）:")
	if err := mustClick(page, "#confirm-btn"); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	confirmResult, err := mustText(page, "#confirm-result")
	if err != nil {
		return err
	}
	fmt.Printf("   result: %s\n", confirmResult)
	fmt.Printf("   opened: %v\n", page.LastPromptOpened())
	fmt.Printf("   closed: %v\n", page.LastPromptClosed())

	fmt.Println("\n3. prompt 登录（步骤式 API / mouse）:")
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(800 * time.Millisecond)
	if err := promptLoginMouse(page, "#login-prompt-btn", "alice", "s3cr3t"); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	mouseResult, err := mustText(page, "#prompt-result")
	if err != nil {
		return err
	}
	fmt.Printf("   result: %s\n", mouseResult)

	fmt.Println("\n4. prompt 登录（步骤式 API / keyboard）:")
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(800 * time.Millisecond)
	if err := promptLoginKeyboard(page, "#login-prompt-btn", "bob", "654321"); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	keyboardResult, err := mustText(page, "#prompt-result")
	if err != nil {
		return err
	}
	fmt.Printf("   result: %s\n", keyboardResult)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 用户提示框自动策略 + 登录式步骤 API 测试完成！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func promptLoginMouse(page *ruyipage.FirefoxPage, selector string, username string, password string) error {
	if _, err := page.RunJS(`function(selector){
		setTimeout(function(){
			document.querySelector(selector).click();
		}, 0);
		return true;
	}`, selector); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)
	return handlePromptLogin(page, username, password)
}

func promptLoginKeyboard(page *ruyipage.FirefoxPage, selector string, username string, password string) error {
	if _, err := page.RunJS(`function(selector){
		const button = document.querySelector(selector);
		button.focus();
		setTimeout(function(){
			button.dispatchEvent(new KeyboardEvent("keydown", {key: "Enter", code: "Enter", bubbles: true}));
			button.dispatchEvent(new KeyboardEvent("keyup", {key: "Enter", code: "Enter", bubbles: true}));
			button.click();
		}, 0);
		return true;
	}`, selector); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)
	return handlePromptLogin(page, username, password)
}

func handlePromptLogin(page *ruyipage.FirefoxPage, username string, password string) error {
	firstPrompt, err := page.WaitPrompt(2 * time.Second)
	if err != nil {
		return err
	}
	fmt.Printf("   第一轮 prompt: %v\n", firstPrompt["message"])
	if err := page.HandlePrompt(true, &username, 2*time.Second); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)

	secondPrompt, err := page.WaitPrompt(2 * time.Second)
	if err != nil {
		return err
	}
	fmt.Printf("   第二轮 prompt: %v\n", secondPrompt["message"])
	return page.HandlePrompt(true, &password, 2*time.Second)
}

func mustClick(page *ruyipage.FirefoxPage, selector string) error {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if element == nil {
		return fmt.Errorf("未找到元素: %s", selector)
	}
	return element.ClickSelf(false, 0)
}

func mustText(page *ruyipage.FirefoxPage, selector string) (string, error) {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return "", err
	}
	if element == nil {
		return "", fmt.Errorf("未找到元素: %s", selector)
	}
	return element.Text()
}
