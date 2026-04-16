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
	fmt.Println("测试13: iframe操作")
	fmt.Println(strings.Repeat("=", 60))

	options := exampleutil.VisibleOptions().WithUserPromptHandler(map[string]string{
		"alert": "accept",
	})
	page, err := ruyipage.NewFirefoxPage(options)
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(2 * time.Second)
		_ = page.Quit(0, false)
	}()

	testURL, err := exampleutil.TestPageURL("test_page.html")
	if err != nil {
		return err
	}
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)

	iframeSection, err := page.Ele("#iframe-section", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if iframeSection == nil {
		return fmt.Errorf("未找到 #iframe-section")
	}
	if err := page.Scroll().ToSee(iframeSection, false); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)

	fmt.Println("\n1. 获取 iframe 元素:")
	iframeElem, err := page.Ele("#test-iframe", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if iframeElem == nil {
		return fmt.Errorf("未找到 #test-iframe")
	}
	tagName, err := iframeElem.Tag()
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ iframe元素已找到: %s\n", tagName)

	fmt.Println("\n1.1 获取所有 iframe:")
	frames, err := page.GetFrames()
	if err != nil {
		return err
	}
	fmt.Printf("   当前页面 frame 数量: %d\n", len(frames))

	fmt.Println("\n2. 切换到 iframe:")
	iframe, err := page.GetFrame("#test-iframe")
	if err != nil {
		return err
	}
	if iframe == nil {
		return fmt.Errorf("未找到 iframe frame 对象")
	}
	fmt.Println("   ✓ 已切换到 iframe")

	fmt.Println("\n2.1 通过 index 获取 iframe:")
	iframeByIndex, err := page.GetFrame(0)
	if err != nil {
		return err
	}
	fmt.Printf("   index=0 是否可用: %v\n", iframeByIndex != nil)

	fmt.Println("\n2.2 通过 context_id 获取 iframe:")
	iframeByContext, err := page.GetFrame(iframe.ContextID())
	if err != nil {
		return err
	}
	fmt.Printf("   context_id 是否匹配: %v\n", iframeByContext != nil && iframeByContext.ContextID() == iframe.ContextID())

	fmt.Println("\n3. 在 iframe 中查找元素:")
	iframeTitle, err := iframe.Ele("tag:h1", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if iframeTitle == nil {
		return fmt.Errorf("未找到 iframe h1")
	}
	titleText, err := iframeTitle.Text()
	if err != nil {
		return err
	}
	fmt.Printf("   iframe标题: %s\n", titleText)

	fmt.Println("\n3.1 iframe 跨域判断:")
	fmt.Printf("   is_cross_origin: %v\n", iframe.IsCrossOrigin())

	fmt.Println("\n4. 在 iframe 中操作按钮:")
	iframeButton, err := iframe.Ele("#iframe-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if iframeButton == nil {
		return fmt.Errorf("未找到 iframe 按钮")
	}
	buttonText, err := iframeButton.Text()
	if err != nil {
		return err
	}
	fmt.Printf("   找到iframe按钮: %s\n", buttonText)
	if err := iframeButton.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Println("   ✓ iframe按钮已点击")

	fmt.Println("\n5. 获取 iframe 内容:")
	iframeHTML, err := iframe.HTML()
	if err != nil {
		return err
	}
	fmt.Printf("   iframe HTML长度: %d 字符\n", len(iframeHTML))

	fmt.Println("\n6. 在 iframe 中执行 JS:")
	bodyHTML, err := iframe.RunJS("return document.body.innerHTML")
	if err != nil {
		return err
	}
	fmt.Printf("   iframe body内容长度: %d 字符\n", len(fmt.Sprint(bodyHTML)))

	fmt.Println("\n6.1 iframe 内修改 DOM 并验证:")
	if _, err := iframe.RunJS(`
		const title = document.querySelector("h1");
		title.textContent = "iframe内容-已修改";
		return title.textContent;
	`); err != nil {
		return err
	}
	changedTitle, err := mustTextFromBase(iframe.FirefoxBase, "tag:h1")
	if err != nil {
		return err
	}
	fmt.Printf("   修改后标题: %s\n", changedTitle)

	fmt.Println("\n7. 切换回主页面:")
	mainTitle, err := mustTextFromBase(page.FirefoxBase, "#main-title")
	if err != nil {
		return err
	}
	fmt.Printf("   主页面标题: %s\n", mainTitle)

	fmt.Println("\n7.1 验证主页面元素仍可操作:")
	clickButton, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if clickButton == nil {
		return fmt.Errorf("未找到 #click-btn")
	}
	if err := clickButton.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(400 * time.Millisecond)
	clickResult, err := mustTextFromBase(page.FirefoxBase, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   主页面点击结果: %s\n", clickResult)

	fmt.Println("\n8. 再次切换到 iframe:")
	iframeAgain, err := page.GetFrame("#test-iframe")
	if err != nil {
		return err
	}
	if iframeAgain == nil {
		return fmt.Errorf("再次访问 iframe 失败")
	}
	titleAgain, err := mustTextFromBase(iframeAgain.FirefoxBase, "tag:h1")
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ 再次访问iframe成功: %s\n", titleAgain)

	fmt.Println("\n8.1 使用 with_frame() 访问 iframe:")
	fmt.Println("   Go 版示例用本地 withFrame() 辅助函数保留同等级阅读体验。")
	if err := withFrame(page, "#test-iframe", func(frame *ruyipage.FirefoxFrame) error {
		frameText, textErr := mustTextFromBase(frame.FirefoxBase, "tag:h1")
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   with_frame 读取标题: %s\n", frameText)
		return nil
	}); err != nil {
		return err
	}

	stillMain, err := mustTextFromBase(page.FirefoxBase, "#main-title")
	if err != nil {
		return err
	}
	fmt.Printf("   with_frame 退出后主页面标题: %s\n", stillMain)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有iframe操作测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
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
