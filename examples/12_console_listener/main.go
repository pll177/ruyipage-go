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
	fmt.Println("测试12: 控制台监听")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Console().Stop()
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

	fmt.Println("\n1. 启动控制台监听:")
	if err := page.Console().Start(""); err != nil {
		return err
	}
	fmt.Println("   ✓ 控制台监听已启动")

	fmt.Println("\n2. 触发 console.log:")
	if err := clickByText(page, "console.log"); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	logs := page.Console().Get("", "")
	fmt.Printf("   捕获到 %d 条日志\n", len(logs))
	for _, logEntry := range logs {
		fmt.Printf("   - [%s] %s\n", logEntry.Level, logEntry.Text)
	}

	fmt.Println("\n3. 触发 console.warn:")
	if err := clickByText(page, "console.warn"); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	logs = page.Console().Get("", "")
	fmt.Printf("   捕获到 %d 条日志\n", len(logs))
	if len(logs) > 0 {
		last := logs[len(logs)-1]
		fmt.Printf("   - [%s] %s\n", last.Level, last.Text)
	}

	fmt.Println("\n4. 触发 console.error:")
	if err := clickByText(page, "console.error"); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	logs = page.Console().Get("", "")
	fmt.Printf("   捕获到 %d 条日志\n", len(logs))
	if len(logs) > 0 {
		last := logs[len(logs)-1]
		fmt.Printf("   - [%s] %s\n", last.Level, last.Text)
	}

	fmt.Println("\n5. 触发 console.info:")
	if err := clickByText(page, "console.info"); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	logs = page.Console().Get("", "")
	fmt.Printf("   捕获到 %d 条日志\n", len(logs))

	fmt.Println("\n5.1 级别过滤 (error):")
	errorLogs := page.Console().Get("error", "")
	fmt.Printf("   error日志数量: %d\n", len(errorLogs))

	fmt.Println("\n6. 清空日志:")
	page.Console().Clear()
	fmt.Printf("   清空后日志数量: %d\n", len(page.Console().Get("", "")))

	fmt.Println("\n7. 通过 JS 输出日志:")
	if _, err := page.RunJS(`console.log("通过JS输出的日志"); console.error("通过JS输出的错误"); return true;`); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	logs = page.Console().Get("", "")
	fmt.Printf("   捕获到 %d 条日志\n", len(logs))
	for _, logEntry := range logs {
		fmt.Printf("   - [%s] %s\n", logEntry.Level, logEntry.Text)
	}

	fmt.Println("\n7.1 wait() 等待指定日志:")
	if _, err := page.RunJS(`console.error("wait-target-message"); return true;`); err != nil {
		return err
	}
	waited := page.Console().Wait("error", "wait-target-message", 5*time.Second)
	if waited == nil {
		return fmt.Errorf("wait 未捕获到目标日志")
	}
	fmt.Printf("   ✓ wait捕获: [%s] %s\n", waited.Level, waited.Text)

	fmt.Println("\n8. 停止监听:")
	page.Console().Stop()
	fmt.Println("   ✓ 控制台监听已停止")

	fmt.Println("\n8.1 停止后验证不再捕获:")
	before := len(page.Console().Get("", ""))
	if _, err := page.RunJS(`console.log("should-not-be-captured-after-stop"); return true;`); err != nil {
		return err
	}
	page.Wait().Sleep(600 * time.Millisecond)
	after := len(page.Console().Get("", ""))
	fmt.Printf("   停止前后日志数量: %d -> %d\n", before, after)
	if before != after {
		return fmt.Errorf("Stop 后日志数量仍在增加: %d -> %d", before, after)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有控制台监听测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func clickByText(page *ruyipage.FirefoxPage, text string) error {
	button, err := page.Ele("text:"+text, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if button == nil {
		return fmt.Errorf("未找到按钮: %s", text)
	}
	return button.ClickSelf(false, 0)
}
