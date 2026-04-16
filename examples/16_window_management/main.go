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
	fmt.Println("测试16: 浏览器窗口管理")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("说明: maximize/minimize/fullscreen/center 为可见浏览器行为，建议肉眼观察窗口变化。")

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
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

	fmt.Println("\n1. 获取当前窗口大小:")
	windowSize := page.Rect().WindowSize()
	fmt.Printf("   窗口大小: %d x %d\n", windowSize["width"], windowSize["height"])

	fmt.Println("\n2. 获取视口大小:")
	viewportSize := page.Rect().ViewportSize()
	fmt.Printf("   视口大小: %d x %d\n", viewportSize["width"], viewportSize["height"])

	fmt.Println("\n3. 获取窗口位置:")
	windowLocation := page.Rect().WindowLocation()
	fmt.Printf("   窗口位置: (%d, %d)\n", windowLocation["x"], windowLocation["y"])

	fmt.Println("\n4. 获取页面完整大小:")
	pageSize := page.Rect().PageSize()
	fmt.Printf("   页面大小: %d x %d\n", pageSize["width"], pageSize["height"])

	fmt.Println("\n5. 获取滚动位置:")
	scrollPosition := page.Rect().ScrollPosition()
	fmt.Printf("   滚动位置: (%d, %d)\n", scrollPosition["x"], scrollPosition["y"])

	fmt.Println("\n6. 滚动页面并检查位置:")
	if err := page.Scroll().ToBottom(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	newScrollPosition := page.Rect().ScrollPosition()
	fmt.Printf("   滚动后位置: (%d, %d)\n", newScrollPosition["x"], newScrollPosition["y"])

	fmt.Println("\n6.1 窗口状态流测试:")
	if err := page.Window().Maximize(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   maximize 后窗口大小: %v\n", page.Rect().WindowSize())

	if err := page.Window().Minimize(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Println("   ✓ minimize 已调用")

	if err := page.Window().Fullscreen(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   fullscreen 后视口大小: %v\n", page.Rect().ViewportSize())

	if err := page.Window().Normal(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Println("   ✓ normal 已恢复")

	fmt.Println("\n6.2 设置窗口尺寸/位置/居中:")
	if err := page.Window().SetSize(1200, 820); err != nil {
		return err
	}
	if err := page.Window().SetPosition(40, 40); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   调整后窗口大小: %v\n", page.Rect().WindowSize())
	fmt.Printf("   调整后窗口位置: %v\n", page.Rect().WindowLocation())

	if err := page.Window().Center(0, 0); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   居中后窗口位置: %v\n", page.Rect().WindowLocation())

	fmt.Println("\n7. 创建新标签页:")
	newTab, err := page.NewTab("", false)
	if err != nil {
		return err
	}
	if newTab == nil {
		return fmt.Errorf("创建新标签页失败")
	}
	page.Wait().Sleep(time.Second)
	fmt.Println("   ✓ 新标签页已创建")

	fmt.Println("\n8. 在新标签页中打开页面:")
	if err := newTab.Get("https://www.example.com"); err != nil {
		return err
	}
	newTab.Wait().Sleep(2 * time.Second)
	newTabURL, err := newTab.URL()
	if err != nil {
		return err
	}
	newTabTitle, err := newTab.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   新标签页URL: %s\n", newTabURL)
	fmt.Printf("   新标签页标题: %s\n", newTabTitle)

	fmt.Println("\n9. 切换回原标签页:")
	if err := page.Activate(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	pageTitle, err := page.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   当前标签页标题: %s\n", pageTitle)

	fmt.Println("\n9.1 后台标签页激活测试:")
	bgTab, err := page.NewTab("https://www.example.com", true)
	if err != nil {
		return err
	}
	if bgTab == nil {
		return fmt.Errorf("创建后台标签页失败")
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   后台标签页ID: %s\n", bgTab.ContextID())
	if _, err := bgTab.Activate(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	bgTitle, err := bgTab.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   激活后标题: %s\n", bgTitle)

	if err := page.Activate(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)

	fmt.Println("\n10. 关闭新标签页:")
	if err := newTab.Close(false); err != nil {
		return err
	}
	if err := bgTab.Close(false); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Println("   ✓ 新标签页已关闭")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有窗口管理测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}
