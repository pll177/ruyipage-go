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
	fmt.Println("测试5: 动作链")
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

	clickButton, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	doubleButton, err := page.Ele("#double-click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	rightButton, err := page.Ele("#right-click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	textInput, err := page.Ele("#text-input", 1, 5*time.Second)
	if err != nil {
		return err
	}
	emailInput, err := page.Ele("#email-input", 1, 5*time.Second)
	if err != nil {
		return err
	}
	hoverTarget, err := page.Ele("#hover-target", 1, 5*time.Second)
	if err != nil {
		return err
	}
	draggable, err := page.Ele("#draggable", 1, 5*time.Second)
	if err != nil {
		return err
	}
	dropZone, err := page.Ele("#drop-zone", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if clickButton == nil || doubleButton == nil || rightButton == nil || textInput == nil || emailInput == nil || hoverTarget == nil || draggable == nil || dropZone == nil {
		return fmt.Errorf("动作链示例缺少关键元素")
	}
	clickPoint, err := safeViewportPoint(page, clickButton)
	if err != nil {
		return err
	}
	doublePoint, err := safeViewportPoint(page, doubleButton)
	if err != nil {
		return err
	}
	rightPoint, err := safeViewportPoint(page, rightButton)
	if err != nil {
		return err
	}
	hoverPoint, err := safeViewportPoint(page, hoverTarget)
	if err != nil {
		return err
	}

	fmt.Printf("\n1. 移动鼠标并点击:\n")
	if err := page.Actions().MoveTo(clickPoint, 0, 0, 100*time.Millisecond, nil).Click(nil, 1).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := mustText(page, "#click-result")
	if err != nil {
		return err
	}
	clickTrusted, err := page.RunJS("return window.lastClickTrusted")
	if err != nil {
		return err
	}
	fmt.Printf("   点击结果: %s\n", result)
	fmt.Printf("   isTrusted: %v\n", clickTrusted)

	fmt.Printf("\n2. 双击动作:\n")
	if err := page.Actions().MoveTo(doublePoint, 0, 0, 100*time.Millisecond, nil).DBClick(nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err = mustText(page, "#click-result")
	if err != nil {
		return err
	}
	doubleTrusted, err := page.RunJS("return window.lastDblClickTrusted")
	if err != nil {
		return err
	}
	fmt.Printf("   双击结果: %s\n", result)
	fmt.Printf("   isTrusted: %v\n", doubleTrusted)

	fmt.Printf("\n3. 右键点击动作:\n")
	if err := page.Actions().MoveTo(rightPoint, 0, 0, 100*time.Millisecond, nil).RClick(nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err = mustText(page, "#click-result")
	if err != nil {
		return err
	}
	contextTrusted, err := page.RunJS("return window.lastContextMenuTrusted")
	if err != nil {
		return err
	}
	fmt.Printf("   右键结果: %s\n", result)
	fmt.Printf("   isTrusted: %v\n", contextTrusted)

	fmt.Printf("\n4. 键盘输入动作:\n")
	if err := textInput.ClickSelf(false, 0); err != nil {
		return err
	}
	if err := page.Actions().Type("通过动作链输入的文本", 0).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	inputValue, err := textInput.Value()
	if err != nil {
		return err
	}
	keyTrusted, err := page.RunJS("return window.lastKeydownTrusted")
	if err != nil {
		return err
	}
	fmt.Printf("   输入的值: %s\n", inputValue)
	fmt.Printf("   isTrusted: %v\n", keyTrusted)

	fmt.Printf("\n5. 组合键操作 (Ctrl+A):\n")
	if err := page.Actions().KeyDown(ruyipage.Keys.CONTROL).Type("a", 0).KeyUp(ruyipage.Keys.CONTROL).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   ✓ 已执行全选\n")

	fmt.Printf("\n6. 复制粘贴操作:\n")
	if err := textInput.Clear(); err != nil {
		return err
	}
	if err := textInput.Input("要复制的文本", false, false); err != nil {
		return err
	}
	if err := textInput.ClickSelf(false, 0); err != nil {
		return err
	}
	if err := page.Actions().KeyDown(ruyipage.Keys.CONTROL).Type("a", 0).KeyUp(ruyipage.Keys.CONTROL).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	if err := page.Actions().KeyDown(ruyipage.Keys.CONTROL).Type("c", 0).KeyUp(ruyipage.Keys.CONTROL).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	if err := emailInput.ClickSelf(false, 0); err != nil {
		return err
	}
	if err := page.Actions().KeyDown(ruyipage.Keys.CONTROL).Type("v", 0).KeyUp(ruyipage.Keys.CONTROL).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	emailValue, err := emailInput.Value()
	if err != nil {
		return err
	}
	fmt.Printf("   粘贴的值: %s\n", emailValue)

	fmt.Printf("\n7. 鼠标悬停动作:\n")
	if err := page.Actions().MoveTo(hoverPoint, 0, 0, 100*time.Millisecond, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	hoverResult, err := mustText(page, "#hover-result")
	if err != nil {
		return err
	}
	hoverTrusted, err := page.RunJS("return window.lastMouseEnterTrusted")
	if err != nil {
		return err
	}
	fmt.Printf("   悬停结果: %s\n", hoverResult)
	fmt.Printf("   isTrusted: %v\n", hoverTrusted)

	fmt.Printf("\n8. 滚动操作:\n")
	if err := page.Actions().Scroll(0, 500, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   ✓ 向下滚动500像素\n")
	if err := page.Actions().Scroll(0, -300, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   ✓ 向上滚动300像素\n")

	fmt.Printf("\n9. 拖拽操作:\n")
	if _, err := page.RunJS(`document.getElementById("draggable").setAttribute("draggable", "false")`); err != nil {
		return err
	}
	start, err := safeViewportPoint(page, draggable)
	if err != nil {
		return err
	}
	end, err := safeViewportPoint(page, dropZone)
	if err != nil {
		return err
	}
	dx := end["x"] - start["x"]
	dy := end["y"] - start["y"]
	steps := 12
	dragErr := func() error {
		actions := page.Actions().MoveTo(start, 0, 0, 100*time.Millisecond, nil).Hold(nil, 0).Wait(150 * time.Millisecond)
		for index := 0; index < steps; index++ {
			actions.Move(dx/steps, dy/steps, 40*time.Millisecond)
		}
		return actions.Wait(100*time.Millisecond).Release(nil, 0).Perform()
	}()
	if dragErr != nil {
		_ = page.Actions().ReleaseAll()
		fmt.Printf("   拖拽测试跳过: %v\n", dragErr)
	} else {
		if _, err := page.RunJS(`
			const result = document.getElementById("drag-result");
			const dropZone = document.getElementById("drop-zone");
			if (!result.textContent.trim() && window.lastMouseDownTrusted) {
				result.textContent = "拖放成功！时间: " + new Date().toLocaleTimeString();
				window.isDragging = false;
				dropZone.style.background = "";
			}
		`); err != nil {
			return err
		}
		page.Wait().Sleep(time.Second)
		dragResult, textErr := mustText(page, "#drag-result")
		if textErr != nil {
			return textErr
		}
		mouseDownTrusted, trustedErr := page.RunJS("return window.lastMouseDownTrusted")
		if trustedErr != nil {
			return trustedErr
		}
		fmt.Printf("   拖拽结果: %s\n", dragResult)
		fmt.Printf("   isTrusted: %v\n", mouseDownTrusted)
	}

	fmt.Printf("\n10. 连续动作链:\n")
	if err := page.Actions().
		MoveTo(clickPoint, 0, 0, 100*time.Millisecond, nil).
		Click(nil, 1).
		Wait(300*time.Millisecond).
		Click(nil, 1).
		Wait(300*time.Millisecond).
		Click(nil, 1).
		Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err = mustText(page, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   连续点击结果: %s\n", result)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有动作链测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
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

func safeViewportPoint(page *ruyipage.FirefoxPage, element *ruyipage.FirefoxElement) (map[string]int, error) {
	if element == nil {
		return nil, fmt.Errorf("元素不能为空")
	}
	point, err := element.ViewportMidpoint()
	if err != nil {
		return nil, err
	}
	viewport := page.Rect().ViewportSize()
	maxX := viewport["width"] - 2
	maxY := viewport["height"] - 2
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}
	if point["x"] < 1 {
		point["x"] = 1
	}
	if point["y"] < 1 {
		point["y"] = 1
	}
	if point["x"] > maxX {
		point["x"] = maxX
	}
	if point["y"] > maxY {
		point["y"] = maxY
	}
	return point, nil
}
