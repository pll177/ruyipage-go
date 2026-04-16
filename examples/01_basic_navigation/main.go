package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

const exampleName = "01_basic_navigation"

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试1: 基础导航和页面操作")
	fmt.Println(strings.Repeat("=", 60))

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

	fmt.Printf("\n1. 导航到测试页面: %s\n", testURL)
	if err := page.Get(testURL); err != nil {
		return err
	}
	fmt.Printf("   ✓ 页面加载成功\n")

	title, err := page.Title()
	if err != nil {
		return err
	}
	currentURL, err := page.URL()
	if err != nil {
		return err
	}
	html, err := page.HTML()
	if err != nil {
		return err
	}
	fmt.Printf("\n2. 获取页面信息:\n")
	fmt.Printf("   标题: %s\n", title)
	fmt.Printf("   URL: %s\n", currentURL)
	fmt.Printf("   HTML长度: %d 字符\n", len(html))

	fmt.Printf("\n3. 导航到其他页面:\n")
	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}
	title, err = page.Title()
	if err != nil {
		return err
	}
	currentURL, err = page.URL()
	if err != nil {
		return err
	}
	fmt.Printf("   当前标题: %s\n", title)
	fmt.Printf("   当前URL: %s\n", currentURL)

	fmt.Printf("\n4. 后退到上一页:\n")
	if err := page.Back(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	title, err = page.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   当前标题: %s\n", title)

	fmt.Printf("\n5. 前进到下一页:\n")
	if err := page.Forward(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	title, err = page.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   当前标题: %s\n", title)

	fmt.Printf("\n6. 刷新页面:\n")
	if err := page.Refresh(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 页面已刷新\n")

	outputDir, err := exampleutil.OutputDir(exampleName)
	if err != nil {
		return err
	}
	fmt.Printf("\n7. 保存页面:\n")
	htmlPath, err := page.Save(outputDir, "example_page", false)
	if err != nil {
		return err
	}
	fmt.Printf("   HTML已保存: %s\n", htmlPath)

	pdfPath, err := page.Save(outputDir, "example_page", true)
	if err != nil {
		return err
	}
	fmt.Printf("   PDF已保存: %s\n", pdfPath)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有基础导航测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}
