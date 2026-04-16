package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

const exampleName = "15_comprehensive"

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试15: 综合测试 - 实战场景")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(3 * time.Second)
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

	fmt.Println("\n场景1: 填写并提交表单")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("1. 填写表单字段...")
	if err := mustInput(page, "#text-input", "张三", true); err != nil {
		return err
	}
	if err := mustInput(page, "#email-input", "zhangsan@example.com", true); err != nil {
		return err
	}
	if err := mustInput(page, "#password-input", "password123", true); err != nil {
		return err
	}
	if err := mustInput(page, "#number-input", "25", true); err != nil {
		return err
	}
	if err := mustInput(page, "#textarea", "这是一段测试文本\n包含多行内容", true); err != nil {
		return err
	}
	if err := mustClick(page, "#checkbox1"); err != nil {
		return err
	}
	if err := mustClick(page, "#checkbox2"); err != nil {
		return err
	}
	if err := mustClick(page, "#radio1"); err != nil {
		return err
	}
	selectElement, err := page.Ele("#select-single", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if selectElement == nil {
		return fmt.Errorf("未找到 #select-single")
	}
	selected, err := selectElement.Select().ByValue("opt2", "native_only")
	if err != nil {
		return err
	}
	if !selected {
		selected, err = selectElement.Select().ByValue("opt2", "compat")
		if err != nil {
			return err
		}
	}
	if !selected {
		return fmt.Errorf("下拉框选择 opt2 失败")
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Println("   ✓ 表单填写完成")

	fmt.Println("2. 提交表单...")
	if err := mustClick(page, "#submit-btn"); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	formResult, err := mustText(page, "#form-result")
	if err != nil {
		return err
	}
	fmt.Printf("   提交结果: %s...\n", trimPreview(formResult, 100))

	fmt.Println("\n场景2: 数据提取")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("1. 提取表格数据...")
	tableRows, err := page.Eles("#data-table tbody tr", 5*time.Second)
	if err != nil {
		return err
	}
	fmt.Printf("   表格共有 %d 行数据\n", len(tableRows))
	for _, row := range tableRows {
		cells, cellsErr := row.Eles("tag:td", 5*time.Second)
		if cellsErr != nil {
			return cellsErr
		}
		if len(cells) < 4 {
			continue
		}
		idValue, err := cells[0].Text()
		if err != nil {
			return err
		}
		nameValue, err := cells[1].Text()
		if err != nil {
			return err
		}
		ageValue, err := cells[2].Text()
		if err != nil {
			return err
		}
		cityValue, err := cells[3].Text()
		if err != nil {
			return err
		}
		fmt.Printf("   - map[id:%s name:%s age:%s city:%s]\n", idValue, nameValue, ageValue, cityValue)
	}

	fmt.Println("\n场景3: 动态交互")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("1. 触发延迟显示...")
	if err := mustClick(page, "#dynamic-section button:nth-of-type(1)"); err != nil {
		return err
	}
	dynamicElement, err := page.Wait().EleDisplayed("#dynamic-content", 5*time.Second)
	if err != nil {
		return err
	}
	if dynamicElement == nil {
		return fmt.Errorf("延迟内容未显示")
	}
	dynamicText, err := dynamicElement.Text()
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ 延迟内容已显示: %s\n", dynamicText)

	fmt.Println("2. 添加多个动态内容...")
	for index := 0; index < 3; index++ {
		if err := mustClick(page, "#dynamic-section button:nth-of-type(4)"); err != nil {
			return err
		}
		page.Wait().Sleep(300 * time.Millisecond)
	}
	newItems, err := page.Eles("#content-container .result", 5*time.Second)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ 已添加 %d 个新内容\n", len(newItems))

	fmt.Println("\n场景4: 页面截图")
	fmt.Println(strings.Repeat("-", 40))
	outputDir, err := exampleutil.OutputDir(exampleName)
	if err != nil {
		return err
	}
	fmt.Println("1. 截取关键区域...")
	formSection, err := page.Ele("#form-section", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if formSection != nil {
		if err := page.Scroll().ToSee(formSection, false); err != nil {
			return err
		}
		page.Wait().Sleep(500 * time.Millisecond)
		if _, err := formSection.Screenshot(filepath.Join(outputDir, "form_filled.png")); err != nil {
			fmt.Printf("   ⚠ 表单截图跳过: %s\n", trimPreview(err.Error(), 50))
		} else {
			fmt.Println("   ✓ 表单截图已保存")
		}
	}
	tableElement, err := page.Ele("#data-table", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if tableElement != nil {
		if err := page.Scroll().ToSee(tableElement, false); err != nil {
			return err
		}
		page.Wait().Sleep(500 * time.Millisecond)
		if _, err := tableElement.Screenshot(filepath.Join(outputDir, "table_data.png")); err != nil {
			fmt.Printf("   ⚠ 表格截图跳过: %s\n", trimPreview(err.Error(), 50))
		} else {
			fmt.Println("   ✓ 表格截图已保存")
		}
	}
	if _, err := page.Screenshot(filepath.Join(outputDir, "full_page.png"), true); err != nil {
		return err
	}
	fmt.Println("   ✓ 整页截图已保存")

	fmt.Println("\n场景5: 多次交互")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("1. 连续点击按钮...")
	clickButton, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if clickButton == nil {
		return fmt.Errorf("未找到 #click-btn")
	}
	for index := 0; index < 5; index++ {
		if err := clickButton.ClickSelf(false, 0); err != nil {
			return err
		}
		page.Wait().Sleep(200 * time.Millisecond)
	}
	clickResult, err := mustText(page, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   点击结果: %s\n", clickResult)

	fmt.Println("2. 鼠标悬停测试...")
	hoverTarget, err := page.Ele("#hover-target", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if hoverTarget == nil {
		return fmt.Errorf("未找到 #hover-target")
	}
	if err := hoverTarget.Hover(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	hoverResult, err := mustText(page, "#hover-result")
	if err != nil {
		return err
	}
	fmt.Printf("   悬停结果: %s\n", hoverResult)

	fmt.Println("3. isTrusted 行为验证...")
	if err := clickButton.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := mustInput(page, "#text-input", "T", false); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := hoverTarget.Hover(); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)
	fmt.Printf("   click isTrusted: %v\n", trustedValue(page, "lastClickTrusted"))
	fmt.Printf("   keydown isTrusted: %v\n", trustedValue(page, "lastKeydownTrusted"))
	fmt.Printf("   mouseenter isTrusted: %v\n", trustedValue(page, "lastMouseEnterTrusted"))

	fmt.Println("\n场景6: 数据验证")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("1. 验证元素状态...")
	disabledButton, err := page.Ele("#disabled-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if disabledButton == nil {
		return fmt.Errorf("未找到 #disabled-btn")
	}
	fmt.Printf("   禁用按钮是否可用: %v\n", disabledButton.States().IsEnabled())
	fmt.Printf("   禁用按钮是否显示: %v\n", disabledButton.States().IsDisplayed())

	fmt.Println("2. 验证输入值...")
	textValue, err := valueOf(page, "#text-input")
	if err != nil {
		return err
	}
	emailValue, err := valueOf(page, "#email-input")
	if err != nil {
		return err
	}
	fmt.Printf("   文本输入框: %s\n", textValue)
	fmt.Printf("   邮箱输入框: %s\n", emailValue)

	fmt.Println("3. 验证选择状态...")
	checkbox1, err := page.Ele("#checkbox1", 1, 5*time.Second)
	if err != nil {
		return err
	}
	radio1, err := page.Ele("#radio1", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if checkbox1 == nil || radio1 == nil {
		return fmt.Errorf("缺少复选框或单选框")
	}
	fmt.Printf("   复选框1选中: %v\n", checkbox1.States().IsChecked())
	fmt.Printf("   单选框1选中: %v\n", radio1.States().IsChecked())

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 综合测试全部通过！")
	fmt.Printf("截图保存在: %s\n", outputDir)
	fmt.Println(strings.Repeat("=", 60))
	return nil
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

func clickByText(page *ruyipage.FirefoxPage, text string) error {
	element, err := page.Ele("text:"+text, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if element == nil {
		return fmt.Errorf("未找到文本按钮: %s", text)
	}
	return element.ClickSelf(false, 0)
}

func mustInput(page *ruyipage.FirefoxPage, selector string, value string, clear bool) error {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if element == nil {
		return fmt.Errorf("未找到输入框: %s", selector)
	}
	return element.Input(value, clear, false)
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

func valueOf(page *ruyipage.FirefoxPage, selector string) (string, error) {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return "", err
	}
	if element == nil {
		return "", fmt.Errorf("未找到元素: %s", selector)
	}
	return element.Value()
}

func trustedValue(page *ruyipage.FirefoxPage, key string) any {
	value, err := page.RunJS("return window[arguments[0]]", key)
	if err != nil {
		return nil
	}
	return value
}

func trimPreview(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
