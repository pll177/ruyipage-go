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
	fmt.Println("测试3: 元素交互")
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

	clickBtn, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if clickBtn == nil {
		return fmt.Errorf("未找到 #click-btn")
	}

	fmt.Printf("\n1. 点击按钮:\n")
	if err := clickBtn.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := mustElementText(page, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   点击结果: %s\n", result)

	if err := clickBtn.ClickSelf(false, 0); err != nil {
		return err
	}
	if err := clickBtn.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err = mustElementText(page, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   多次点击后: %s\n", result)

	textInput, err := page.Ele("#text-input", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if textInput == nil {
		return fmt.Errorf("未找到 #text-input")
	}

	fmt.Printf("\n2. 输入文本:\n")
	if err := textInput.Input("Hello RuyiPage!", false, false); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	textValue, err := textInput.Value()
	if err != nil {
		return err
	}
	fmt.Printf("   输入的值: %s\n", textValue)

	fmt.Printf("\n3. 清空并重新输入:\n")
	if err := textInput.Clear(); err != nil {
		return err
	}
	if err := textInput.Input("新的文本内容", false, false); err != nil {
		return err
	}
	textValue, err = textInput.Value()
	if err != nil {
		return err
	}
	fmt.Printf("   新的值: %s\n", textValue)

	fmt.Printf("\n4. 输入到不同类型的输入框:\n")
	if err := inputBySelector(page, "#email-input", "test@example.com"); err != nil {
		return err
	}
	if err := inputBySelector(page, "#password-input", "password123"); err != nil {
		return err
	}
	if err := inputBySelector(page, "#number-input", "42"); err != nil {
		return err
	}
	if err := inputBySelector(page, "#textarea", "这是多行文本\n第二行\n第三行"); err != nil {
		return err
	}
	fmt.Printf("   ✓ 所有输入框填写完成\n")

	fmt.Printf("\n5. 复选框操作:\n")
	checkbox1, err := page.Ele("#checkbox1", 1, 5*time.Second)
	if err != nil {
		return err
	}
	checkbox2, err := page.Ele("#checkbox2", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if checkbox1 == nil || checkbox2 == nil {
		return fmt.Errorf("未找到复选框")
	}
	checked, err := checkbox1.IsChecked()
	if err != nil {
		return err
	}
	fmt.Printf("   checkbox1初始状态: %v\n", checked)
	if err := checkbox1.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	checked, err = checkbox1.IsChecked()
	if err != nil {
		return err
	}
	fmt.Printf("   checkbox1点击后: %v\n", checked)
	if err := checkbox2.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	checked, err = checkbox2.IsChecked()
	if err != nil {
		return err
	}
	fmt.Printf("   checkbox2点击后: %v\n", checked)

	fmt.Printf("\n6. 单选框操作:\n")
	radio1, err := page.Ele("#radio1", 1, 5*time.Second)
	if err != nil {
		return err
	}
	radio2, err := page.Ele("#radio2", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if radio1 == nil || radio2 == nil {
		return fmt.Errorf("未找到单选框")
	}
	if err := radio1.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	radio1Checked, err := radio1.IsChecked()
	if err != nil {
		return err
	}
	fmt.Printf("   radio1选中: %v\n", radio1Checked)
	if err := radio2.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	radio2Checked, err := radio2.IsChecked()
	if err != nil {
		return err
	}
	radio1Checked, err = radio1.IsChecked()
	if err != nil {
		return err
	}
	fmt.Printf("   radio2选中: %v\n", radio2Checked)
	fmt.Printf("   radio1现在: %v\n", radio1Checked)

	fmt.Printf("\n7. 下拉选择:\n")
	selectElem, err := page.Ele("#select-single", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if selectElem == nil {
		return fmt.Errorf("未找到下拉框")
	}
	selectOK, err := selectElem.Select().ByValue("opt2", "native_only")
	if err != nil {
		return err
	}
	if !selectOK {
		fmt.Println("   native_only 失败，切换到 compat 模式保底...")
		selectOK, err = selectElem.Select().ByValue("opt2", "compat")
		if err != nil {
			return err
		}
	}
	page.Wait().Sleep(500 * time.Millisecond)
	selectedValue, err := selectElem.Value()
	if err != nil {
		return err
	}
	selected := selectElem.Select().SelectedOption()
	fmt.Printf("   选中的值: %s\n", selectedValue)
	fmt.Printf("   选中的文本: %v\n", selected["text"])
	fmt.Printf("   原生选择结果: %v\n", selectOK)

	fmt.Printf("\n8. 双击按钮:\n")
	doubleButton, err := page.Ele("#double-click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if doubleButton == nil {
		return fmt.Errorf("未找到双击按钮")
	}
	if err := doubleButton.DoubleClick(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err = mustElementText(page, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   双击结果: %s\n", result)

	fmt.Printf("\n9. 右键点击:\n")
	rightButton, err := page.Ele("#right-click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if rightButton == nil {
		return fmt.Errorf("未找到右键按钮")
	}
	if err := rightButton.RightClick(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err = mustElementText(page, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   右键点击结果: %s\n", result)

	fmt.Printf("\n10. 鼠标悬停:\n")
	hoverTarget, err := page.Ele("#hover-target", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if hoverTarget == nil {
		return fmt.Errorf("未找到悬停区域")
	}
	if err := hoverTarget.Hover(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	hoverResult, err := mustElementText(page, "#hover-result")
	if err != nil {
		return err
	}
	fmt.Printf("   悬停结果: %s\n", hoverResult)

	fmt.Printf("\n11. 提交表单:\n")
	submitButton, err := page.Ele("#submit-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if submitButton == nil {
		return fmt.Errorf("未找到提交按钮")
	}
	if err := submitButton.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	formResult, err := mustElementText(page, "#form-result")
	if err != nil {
		return err
	}
	fmt.Printf("   表单结果: %s\n", trimForPreview(formResult, 50))

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有元素交互测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func inputBySelector(page *ruyipage.FirefoxPage, selector string, value string) error {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if element == nil {
		return fmt.Errorf("未找到元素: %s", selector)
	}
	return element.Input(value, false, false)
}

func mustElementText(page *ruyipage.FirefoxPage, selector string) (string, error) {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return "", err
	}
	if element == nil {
		return "", fmt.Errorf("未找到元素: %s", selector)
	}
	return element.Text()
}

func trimForPreview(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
