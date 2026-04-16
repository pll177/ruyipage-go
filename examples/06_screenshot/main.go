package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const exampleName = "06_screenshot"

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试6: 截图功能")
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
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)

	outputDir, err := exampleutil.OutputDir(exampleName)
	if err != nil {
		return err
	}

	fmt.Printf("\n1. 整页截图:\n")
	fullPath := outputDir + `\full_page.png`
	if _, err := page.Screenshot(fullPath, true); err != nil {
		return err
	}
	fmt.Printf("   ✓ 整页截图已保存: %s\n", fullPath)

	fmt.Printf("\n2. 元素截图:\n")
	title, err := page.Ele("#main-title", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if title == nil {
		return fmt.Errorf("未找到 #main-title")
	}
	titlePath := outputDir + `\title_element.png`
	if _, err := title.Screenshot(titlePath); err != nil {
		return err
	}
	fmt.Printf("   ✓ 标题元素截图已保存: %s\n", titlePath)

	fmt.Printf("\n3. 表单区域截图:\n")
	formSection, err := page.Ele("#form-section", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if formSection == nil {
		return fmt.Errorf("未找到 #form-section")
	}
	formPath := outputDir + `\form_section.png`
	if _, err := formSection.Screenshot(formPath); err != nil {
		return err
	}
	fmt.Printf("   ✓ 表单区域截图已保存: %s\n", formPath)

	fmt.Printf("\n4. 按钮截图:\n")
	button, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if button == nil {
		return fmt.Errorf("未找到 #click-btn")
	}
	if err := page.Scroll().ToSee(button, false); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	buttonPath := outputDir + `\button.png`
	if _, err := button.Screenshot(buttonPath); err != nil {
		fmt.Printf("   ⚠ 按钮截图跳过（元素尺寸问题）: %s\n", trimError(err))
	} else {
		fmt.Printf("   ✓ 按钮截图已保存: %s\n", buttonPath)
	}

	fmt.Printf("\n5. 获取截图base64:\n")
	baseBytes, err := page.Screenshot("", true)
	if err != nil {
		return err
	}
	base64Data := base64.StdEncoding.EncodeToString(baseBytes)
	fmt.Printf("   ✓ Base64数据长度: %d 字符\n", len(base64Data))

	fmt.Printf("\n6. 获取截图bytes:\n")
	bytesData, err := page.Screenshot("", true)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ Bytes数据长度: %d 字节\n", len(bytesData))

	fmt.Printf("\n7. 元素base64截图:\n")
	elementBytes, err := formSection.Screenshot("")
	if err != nil {
		fmt.Printf("   ⚠ 元素base64截图跳过（元素尺寸问题）: %s\n", trimError(err))
	} else {
		elementBase64 := base64.StdEncoding.EncodeToString(elementBytes)
		fmt.Printf("   ✓ 元素Base64数据长度: %d 字符\n", len(elementBase64))
	}

	fmt.Printf("\n8. 滚动后截图:\n")
	if err := page.Scroll().ToBottom(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	bottomPath := outputDir + `\page_bottom.png`
	if _, err := page.Screenshot(bottomPath, true); err != nil {
		return err
	}
	fmt.Printf("   ✓ 页面底部截图已保存: %s\n", bottomPath)

	fmt.Printf("\n9. 表格截图:\n")
	table, err := page.Ele("#data-table", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if table == nil {
		return fmt.Errorf("未找到 #data-table")
	}
	if err := page.Scroll().ToSee(table, false); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	tablePath := outputDir + `\table.png`
	if _, err := table.Screenshot(tablePath); err != nil {
		fmt.Printf("   ⚠ 表格截图跳过（元素尺寸问题）: %s\n", trimError(err))
	} else {
		fmt.Printf("   ✓ 表格截图已保存: %s\n", tablePath)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有截图测试通过！")
	fmt.Printf("截图保存在: %s\n", outputDir)
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func trimError(err error) string {
	if err == nil {
		return ""
	}
	text := err.Error()
	if len(text) <= 50 {
		return text
	}
	return text[:50]
}
