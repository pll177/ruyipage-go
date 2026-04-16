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
	fmt.Println("测试4: 等待条件")
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

	waiter := page.Wait()

	fmt.Printf("\n1. 等待元素出现:\n")
	elem, err := waiter.Ele("#main-title", 5*time.Second)
	if err != nil {
		return err
	}
	if elem != nil {
		text, textErr := elem.Text()
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   ✓ 元素已找到: %s\n", text)
	} else {
		fmt.Printf("   ✗ 元素未找到\n")
		return fmt.Errorf("等待 #main-title 失败")
	}

	fmt.Printf("\n2. 等待元素可见:\n")
	button, err := waiter.EleDisplayed("#click-btn", 5*time.Second)
	if err != nil {
		return err
	}
	if button != nil {
		fmt.Printf("   ✓ 按钮可见\n")
	} else {
		fmt.Printf("   ✗ 按钮不可见\n")
		return fmt.Errorf("等待 #click-btn 可见失败")
	}

	fmt.Printf("\n3. 等待延迟显示的内容:\n")
	showButton, err := page.Ele(`button[onclick="showDelayedContent()"]`, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if showButton == nil {
		return fmt.Errorf("未找到显示延迟内容按钮")
	}
	if err := showButton.ClickSelf(false, 0); err != nil {
		return err
	}
	fmt.Printf("   已点击按钮，等待内容显示...\n")
	dynamicElem, err := waiter.EleDisplayed("#dynamic-content", 5*time.Second)
	if err != nil {
		return err
	}
	if dynamicElem != nil {
		text, textErr := dynamicElem.Text()
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   ✓ 延迟内容已显示: %s\n", text)
	} else {
		fmt.Printf("   ✗ 延迟内容未显示\n")
		return fmt.Errorf("等待 #dynamic-content 显示失败")
	}

	fmt.Printf("\n4. 等待元素隐藏:\n")
	hideButton, err := page.Ele(`button[onclick="hideContent()"]`, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if hideButton == nil {
		return fmt.Errorf("未找到隐藏内容按钮")
	}
	if err := hideButton.ClickSelf(false, 0); err != nil {
		return err
	}
	fmt.Printf("   已点击隐藏按钮...\n")
	hidden, err := waiter.EleHidden("#dynamic-content", 3*time.Second)
	if err != nil {
		return err
	}
	if hidden {
		fmt.Printf("   ✓ 元素已隐藏\n")
	} else {
		fmt.Printf("   ✗ 元素仍然可见\n")
		return fmt.Errorf("等待 #dynamic-content 隐藏失败")
	}

	fmt.Printf("\n5. 等待元素从DOM删除:\n")
	removeButton, err := page.Ele(`button[onclick="removeContent()"]`, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if removeButton == nil {
		return fmt.Errorf("未找到删除内容按钮")
	}
	if err := removeButton.ClickSelf(false, 0); err != nil {
		return err
	}
	fmt.Printf("   已点击删除按钮...\n")
	deleted, err := waiter.EleDeleted("#dynamic-content", 3*time.Second)
	if err != nil {
		return err
	}
	if deleted {
		fmt.Printf("   ✓ 元素已从DOM删除\n")
	} else {
		fmt.Printf("   ✗ 元素仍在DOM中\n")
		return fmt.Errorf("等待 #dynamic-content 删除失败")
	}

	fmt.Printf("\n6. 等待标题包含特定文本:\n")
	title, err := page.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   当前标题: %s\n", title)
	titleContains, err := waiter.TitleContains("RuyiPage", 2*time.Second)
	if err != nil {
		return err
	}
	if titleContains {
		fmt.Printf("   ✓ 标题包含'RuyiPage'\n")
	} else {
		fmt.Printf("   ✗ 标题不包含'RuyiPage'\n")
		return fmt.Errorf("等待标题包含 RuyiPage 失败")
	}

	fmt.Printf("\n7. 简单等待2秒:\n")
	waiter.Sleep(2 * time.Second)
	fmt.Printf("   ✓ 等待完成\n")

	fmt.Printf("\n8. 等待页面加载完成:\n")
	docLoaded, err := waiter.DocLoaded(5 * time.Second)
	if err != nil {
		return err
	}
	if !docLoaded {
		return fmt.Errorf("等待页面加载完成失败")
	}
	fmt.Printf("   ✓ 页面已加载完成\n")

	fmt.Printf("\n9. 添加新内容并等待:\n")
	addButton, err := page.Ele(`button[onclick="addContent()"]`, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if addButton == nil {
		return fmt.Errorf("未找到添加内容按钮")
	}
	if err := addButton.ClickSelf(false, 0); err != nil {
		return err
	}
	waiter.Sleep(500 * time.Millisecond)
	newContent, err := waiter.Ele("#content-container .result", 3*time.Second)
	if err != nil {
		return err
	}
	if newContent != nil {
		text, textErr := newContent.Text()
		if textErr != nil {
			return textErr
		}
		fmt.Printf("   ✓ 新内容已添加: %s\n", text)
	} else {
		fmt.Printf("   ✗ 新内容未找到\n")
		return fmt.Errorf("等待新内容失败")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有等待条件测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}
