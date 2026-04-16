package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试14: Shadow DOM + 嵌套iframe")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(time.Second)
		_ = page.Quit(0, false)
	}()

	testURL, err := exampleutil.TestPageURL("complex_shadow_iframe.html")
	if err != nil {
		return err
	}
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)

	fmt.Println("\n1. 主页面 open shadow:")
	hostOpen, err := page.Ele("#host-open-shadow", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if hostOpen == nil {
		return fmt.Errorf("未找到 #host-open-shadow")
	}
	if err := withShadow(hostOpen, "open", func(root *ruyipage.FirefoxElement) error {
		text, textErr := mustTextFromElement(root, "#host-open-text")
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   open文本: %s\n", text)
		return nil
	}); err != nil {
		return err
	}

	fmt.Println("\n2. 主页面 closed shadow:")
	hostClosed, err := page.Ele("#host-closed-shadow", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if hostClosed == nil {
		return fmt.Errorf("未找到 #host-closed-shadow")
	}
	if err := withShadow(hostClosed, "closed", func(root *ruyipage.FirefoxElement) error {
		text, textErr := mustTextFromElement(root, "#host-closed-text")
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   closed文本: %s\n", text)
		return nil
	}); err != nil {
		return err
	}

	fmt.Println("\n3. 进入 iframe 并测试 shadow:")
	fmt.Println("   Go 版示例用本地 withFrame() / withShadow() 辅助函数保留 Python 示例结构。")
	if err := withFrame(page, "#outer-iframe", func(frame *ruyipage.FirefoxFrame) error {
		innerTitle, textErr := mustTextFromBase(frame.FirefoxBase, "#inner-title")
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   iframe标题: %s\n", innerTitle)

		innerOpen, openErr := frame.Ele("#inner-open-host", 1, 5*time.Second)
		if openErr != nil {
			return openErr
		}
		if innerOpen == nil {
			return fmt.Errorf("未找到 #inner-open-host")
		}
		if err := withShadow(innerOpen, "open", func(root *ruyipage.FirefoxElement) error {
			text, textErr := mustTextFromElement(root, "#inner-open-text")
			if textErr != nil {
				return textErr
			}
			fmt.Printf("   iframe open文本: %s\n", text)
			return nil
		}); err != nil {
			return err
		}

		innerClosed, closedErr := frame.Ele("#inner-closed-host", 1, 5*time.Second)
		if closedErr != nil {
			return closedErr
		}
		if innerClosed == nil {
			return fmt.Errorf("未找到 #inner-closed-host")
		}
		return withShadow(innerClosed, "closed", func(root *ruyipage.FirefoxElement) error {
			text, textErr := mustTextFromElement(root, "#inner-closed-text")
			if textErr != nil {
				return textErr
			}
			fmt.Printf("   iframe closed文本: %s\n", text)
			return nil
		})
	}); err != nil {
		return err
	}

	fmt.Println("\n4. 退出 iframe 后主页面可访问:")
	hostTitle, err := mustTextFromBase(page.FirefoxBase, "#host-title")
	if err != nil {
		return err
	}
	fmt.Printf("   主页面标题: %s\n", hostTitle)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ Shadow DOM + 嵌套iframe 测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func withShadow(host *ruyipage.FirefoxElement, mode string, handler func(*ruyipage.FirefoxElement) error) error {
	if host == nil {
		return fmt.Errorf("shadow host 不能为空")
	}
	root, err := host.WithShadow(mode)
	if err != nil {
		return err
	}
	if root == nil {
		return fmt.Errorf("未获取到 %s shadow root", mode)
	}
	return handler(root)
}

func withFrame(page *ruyipage.FirefoxPage, locator any, handler func(*ruyipage.FirefoxFrame) error) error {
	frame, err := page.GetFrame(locator)
	if err != nil {
		return err
	}
	if frame == nil {
		return fmt.Errorf("未找到目标 iframe/frame")
	}
	return handler(frame)
}

func mustTextFromBase(base *ruyipage.FirefoxBase, locator any) (string, error) {
	if base == nil {
		return "", fmt.Errorf("页面上下文不能为空")
	}
	element, err := base.Ele(locator, 1, 5*time.Second)
	if err != nil {
		return "", err
	}
	if element == nil {
		return "", fmt.Errorf("未找到元素: %v", locator)
	}
	return element.Text()
}

func mustTextFromElement(root *ruyipage.FirefoxElement, locator any) (string, error) {
	if root == nil {
		return "", fmt.Errorf("shadow root 不能为空")
	}
	element, err := root.Ele(locator, 1, 5*time.Second)
	if err != nil {
		return "", err
	}
	if element == nil {
		return "", fmt.Errorf("未找到元素: %v", locator)
	}
	return element.Text()
}
