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
	fmt.Println("测试2: 元素查找")
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

	title, err := page.Ele("#main-title", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if title == nil {
		return fmt.Errorf("未找到 #main-title")
	}

	titleText, err := title.Text()
	if err != nil {
		return err
	}
	tagName, err := title.Tag()
	if err != nil {
		return err
	}
	fmt.Printf("\n1. 通过ID查找元素:\n")
	fmt.Printf("   标题文本: %s\n", titleText)
	fmt.Printf("   标签名: %s\n", tagName)

	testDiv, err := page.Ele(".test-class", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if testDiv == nil {
		return fmt.Errorf("未找到 .test-class")
	}
	testDivText, err := testDiv.Text()
	if err != nil {
		return err
	}
	fmt.Printf("\n2. 通过class查找元素:\n")
	fmt.Printf("   第一个.test-class元素: %s\n", testDivText)

	allTest, err := page.Eles(".test-class", 5*time.Second)
	if err != nil {
		return err
	}
	fmt.Printf("\n3. 查找所有.test-class元素:\n")
	fmt.Printf("   找到 %d 个元素\n", len(allTest))
	for index, elem := range allTest {
		text, textErr := elem.Text()
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   元素%d: %s\n", index+1, text)
	}

	link, err := page.Ele(`xpath://a[@id="test-link"]`, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if link == nil {
		return fmt.Errorf("未找到测试链接")
	}
	linkText, err := link.Text()
	if err != nil {
		return err
	}
	linkURL, err := link.Link()
	if err != nil {
		return err
	}
	fmt.Printf("\n4. 通过XPath查找:\n")
	fmt.Printf("   链接文本: %s\n", linkText)
	fmt.Printf("   链接地址: %s\n", linkURL)

	button, err := page.Ele("text:点击我", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if button == nil {
		return fmt.Errorf("未找到点击按钮")
	}
	buttonID, err := button.Attr("id")
	if err != nil {
		return err
	}
	fmt.Printf("\n5. 通过文本查找:\n")
	fmt.Printf("   按钮ID: %s\n", buttonID)

	dataDiv, err := page.Ele(`[data-test="value"]`, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if dataDiv == nil {
		return fmt.Errorf("未找到 data-test 元素")
	}
	dataText, err := dataDiv.Text()
	if err != nil {
		return err
	}
	attrs, err := dataDiv.Attrs()
	if err != nil {
		return err
	}
	fmt.Printf("\n6. 通过属性查找:\n")
	fmt.Printf("   data-test属性的元素: %s\n", dataText)
	fmt.Printf("   所有属性: %v\n", attrs)

	section, err := page.Ele("#form-section", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if section == nil {
		return fmt.Errorf("未找到 #form-section")
	}
	inputElem, err := section.Ele(`input[type="text"]`, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if inputElem == nil {
		return fmt.Errorf("未找到表单输入框")
	}
	placeholder, err := inputElem.Attr("placeholder")
	if err != nil {
		return err
	}
	fmt.Printf("\n7. 组合选择器:\n")
	fmt.Printf("   表单区域内的文本输入框: %s\n", placeholder)

	img, err := page.Ele("#test-img", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if img == nil {
		return fmt.Errorf("未找到测试图片")
	}
	altText, err := img.Attr("alt")
	if err != nil {
		return err
	}
	srcURL, err := img.Src()
	if err != nil {
		return err
	}
	fmt.Printf("\n8. 获取元素属性:\n")
	fmt.Printf("   图片alt: %s\n", altText)
	fmt.Printf("   图片src: %s\n", srcURL)

	disabledButton, err := page.Ele("#disabled-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if disabledButton == nil {
		return fmt.Errorf("未找到禁用按钮")
	}
	enabled, err := disabledButton.IsEnabled()
	if err != nil {
		return err
	}
	displayed, err := disabledButton.IsDisplayed()
	if err != nil {
		return err
	}
	fmt.Printf("\n9. 元素状态检查:\n")
	fmt.Printf("   禁用按钮是否可用: %v\n", enabled)
	fmt.Printf("   禁用按钮是否显示: %v\n", displayed)

	clickButton, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if clickButton == nil {
		return fmt.Errorf("未找到点击按钮")
	}
	size, err := clickButton.Size()
	if err != nil {
		return err
	}
	location, err := clickButton.Location()
	if err != nil {
		return err
	}
	fmt.Printf("\n10. 元素尺寸和位置:\n")
	fmt.Printf("   按钮尺寸: %v\n", size)
	fmt.Printf("   按钮位置: %v\n", location)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有元素查找测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}
