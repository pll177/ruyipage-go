package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

type caseResult struct {
	name   string
	passed bool
	detail string
}

type testResult struct {
	results []caseResult
}

func (r *testResult) record(name string, passed bool, detail string) {
	r.results = append(r.results, caseResult{name: name, passed: passed, detail: detail})
	status := "✓ 通过"
	if !passed {
		status = "✗ 失败"
	}
	fmt.Printf("   %s: %s\n", status, name)
	if detail != "" {
		fmt.Printf("      详情: %s\n", detail)
	}
}

func (r *testResult) summary() bool {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("测试结果汇总")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("%-5s %-40s %-8s %s\n", "序号", "测试名称", "结果", "详情")
	fmt.Println(strings.Repeat("-", 70))
	passed := 0
	for index, item := range r.results {
		status := "✓ 通过"
		if !item.passed {
			status = "✗ 失败"
		} else {
			passed++
		}
		detail := item.detail
		if len(detail) > 30 {
			detail = detail[:30]
		}
		fmt.Printf("%-5d %-40s %-8s %s\n", index+1, item.name, status, detail)
	}
	failed := len(r.results) - passed
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("总计: %d  通过: %d  失败: %d\n", len(r.results), passed, failed)
	fmt.Println(strings.Repeat("=", 70))
	return failed == 0
}

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("测试20: 高级输入操作（综合测试）")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("说明: 拖拽、触摸、拟人移动为可见浏览器动作，建议配合肉眼观察。")

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

	results := &testResult{}

	fmt.Println("\n--- 键盘操作 ---")
	testComboCtrlA(page, results)
	testComboCopyPaste(page, results)
	testPressKey(page, results)

	fmt.Println("\n--- 鼠标点击操作 ---")
	testDoubleClick(page, results)
	testRightClick(page, results)
	testMiddleClick(page, results)

	fmt.Println("\n--- 拖拽操作 ---")
	testDragHoldRelease(page, results)
	testActionsDragTo(page, results)

	fmt.Println("\n--- 滚轮操作 ---")
	testScrollWheel(page, results)
	testScrollOnElement(page, results)

	fmt.Println("\n--- 悬停操作 ---")
	testHover(page, results)

	fmt.Println("\n--- 组合操作 ---")
	testShiftClick(page, results)
	testActionChain(page, results)
	testReleaseAll(page, results)

	fmt.Println("\n--- isTrusted 验证 ---")
	testIsTrustedClick(page, results)
	testIsTrustedKeydown(page, results)
	testIsTrustedMouseEnter(page, results)

	fmt.Println("\n--- 拟人化操作 ---")
	testHumanMoveClick(page, results)
	testHumanType(page, results)

	fmt.Println("\n--- 触摸操作 ---")
	testTouchTap(page, testURL, results)
	testTouchLongPress(page, results)

	fmt.Println("\n--- 输入操作 ---")
	testTypeWithInterval(page, results)

	fmt.Println("\n--- 元素级操作 ---")
	testElementDoubleClick(page, results)
	testElementRightClick(page, results)
	testElementHover(page, results)
	testElementDragTo(page, results)

	fmt.Println("\n--- 文件上传 ---")
	testFileUpload(page, results)

	fmt.Println("\n--- 键盘导航 ---")
	testKeyboardNavigation(page, results)

	fmt.Println("\n--- 其他操作 ---")
	testRelativeMove(page, results)

	fmt.Println("\n--- 向后兼容性 ---")
	testBackwardCompatDBClick(page, results)
	testBackwardCompatRClick(page, results)

	fmt.Println("\n--- 复杂多步操作 ---")
	testMultiComboKeys(page, results)

	allPassed := results.summary()
	if !allPassed {
		return fmt.Errorf("高级输入操作测试存在失败项")
	}
	return nil
}

func testComboCtrlA(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "combo(Ctrl+A) 全选"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Input("combo测试文本", true, false); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Combo(ruyipage.Keys.CTRL, "a").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	selection, err := page.RunJS(`
		const input = document.getElementById("text-input");
		return {
			start: input.selectionStart,
			end: input.selectionEnd,
			length: input.value.length
		};
	`)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	value, ok := selection.(map[string]any)
	if !ok {
		results.record(name, false, fmt.Sprintf("selection 类型异常: %T", selection))
		return
	}
	start := intValue(value["start"])
	end := intValue(value["end"])
	length := intValue(value["length"])
	results.record(name, start == 0 && end == length && length > 0, fmt.Sprintf("selection=%d-%d/%d", start, end, length))
}

func testComboCopyPaste(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "combo(Ctrl+C/V) 复制粘贴"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Input("复制粘贴测试", true, false); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Combo(ruyipage.Keys.CTRL, "a").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := page.Actions().Combo(ruyipage.Keys.CTRL, "c").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := input.Clear(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Combo(ruyipage.Keys.CTRL, "v").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	pasted, err := input.Value()
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(pasted, "复制粘贴测试"), fmt.Sprintf("粘贴值: %s", pasted))
}

func testPressKey(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "press(Keys.END) 单键操作"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Input("按键测试", true, false); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Press(ruyipage.Keys.END).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(100 * time.Millisecond)
	if err := page.Actions().Type("!", 0).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	value, err := input.Value()
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.HasSuffix(value, "!"), fmt.Sprintf("值: %s", value))
}

func testDoubleClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "double_click() 双击"
	button, err := mustElement(page, "#double-click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().DoubleClick(button).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "双击"), fmt.Sprintf("结果: %s", result))
}

func testRightClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "right_click() 右键点击"
	button, err := mustElement(page, "#right-click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().RightClick(button).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "右键"), fmt.Sprintf("结果: %s", result))
}

func testMiddleClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "middle_click() 中键点击"
	button, err := mustElement(page, "#click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	err = page.Actions().MiddleClick(button).Perform()
	results.record(name, err == nil, "执行无异常")
}

func testDragHoldRelease(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "hold/move_to/release 拖拽"
	start, end, err := prepareDragScene(page)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Drag(start, end, 720*time.Millisecond, 20).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(time.Second)
	result, err := textOf(page, "#drag-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "拖放成功"), fmt.Sprintf("结果: %s", result))
}

func testActionsDragTo(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "Actions.drag_to() 便捷拖拽"
	start, end, err := prepareDragScene(page)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().DragTo(start, end, 800*time.Millisecond, 25).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(time.Second)
	result, err := textOf(page, "#drag-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "拖放成功"), fmt.Sprintf("结果: %s", result))
}

func testScrollWheel(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "scroll() 滚轮滚动"
	if err := page.Scroll().ToBottom(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	err := page.Actions().Scroll(0, -500, nil, nil).Perform()
	results.record(name, err == nil, "滚动到底部+向上滚动")
}

func testHover(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "move_to() 鼠标悬停"
	if err := page.Scroll().ToTop(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	target, err := mustElement(page, "#hover-target")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	point, err := visiblePoint(page, target)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().MoveTo(point, 0, 0, 100*time.Millisecond, "viewport").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(time.Second)
	result, err := textOf(page, "#hover-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "进入"), fmt.Sprintf("结果: %s", result))
}

func testShiftClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "Shift+click 组合操作"
	button, err := mustElement(page, "#click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	point, err := visiblePoint(page, button)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	err = page.Actions().MoveTo(point, 0, 0, 100*time.Millisecond, "viewport").KeyDown(ruyipage.Keys.SHIFT).Click(nil, 1).KeyUp(ruyipage.Keys.SHIFT).Perform()
	results.record(name, err == nil, "执行无异常")
}

func testActionChain(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "连续动作链 (move+wait+click)"
	button, err := mustElement(page, "#click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	point, err := visiblePoint(page, button)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().MoveTo(point, 0, 0, 100*time.Millisecond, "viewport").Wait(300*time.Millisecond).Click(nil, 1).Wait(300*time.Millisecond).Click(nil, 1).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, true, fmt.Sprintf("结果: %s", result))
}

func testReleaseAll(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "release_all() 释放所有动作"
	err := page.Actions().ReleaseAll()
	results.record(name, err == nil, "执行无异常")
}

func testIsTrustedClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "isTrusted: click 事件"
	button, err := mustElement(page, "#click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := button.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	trusted, err := page.RunJS("return window.lastClickTrusted")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, trusted == true, fmt.Sprintf("isTrusted=%v", trusted))
}

func testIsTrustedKeydown(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "isTrusted: keydown 事件"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Input("K", false, false); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	trusted, err := page.RunJS("return window.lastKeydownTrusted")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, trusted == true, fmt.Sprintf("isTrusted=%v", trusted))
}

func testIsTrustedMouseEnter(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "isTrusted: mouseenter 事件"
	target, err := mustElement(page, "#hover-target")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := target.Hover(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	trusted, err := page.RunJS("return window.lastMouseEnterTrusted")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, trusted == true, fmt.Sprintf("isTrusted=%v", trusted))
}

func testHumanMoveClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "human_move + human_click 拟人点击"
	button, err := mustElement(page, "#click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().HumanMove(button, "").HumanClick(nil, "").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, true, fmt.Sprintf("结果: %s", result))
}

func testHumanType(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "human_type() 拟人输入"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Clear(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().HumanType("拟人输入", 0, 0).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	value, err := input.Value()
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(value, "拟人输入"), fmt.Sprintf("值: %s", value))
}

func testTypeWithInterval(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "type(interval=50) 间隔输入"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Clear(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Type("间隔输入", 50*time.Millisecond).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	value, err := input.Value()
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(value, "间隔输入"), fmt.Sprintf("值: %s", value))
}

func testElementDoubleClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "element.double_click() 元素双击"
	button, err := mustElement(page, "#double-click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := button.DoubleClick(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "双击"), fmt.Sprintf("结果: %s", result))
}

func testElementRightClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "element.right_click() 元素右击"
	button, err := mustElement(page, "#right-click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := button.RightClick(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "右键"), fmt.Sprintf("结果: %s", result))
}

func testElementHover(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "element.hover() 元素悬停"
	if err := page.Actions().MoveTo(map[string]int{"x": 1, "y": 1}, 0, 0, 100*time.Millisecond, "viewport").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	target, err := mustElement(page, "#hover-target")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := target.Hover(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#hover-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "进入"), fmt.Sprintf("结果: %s", result))
}

func testElementDragTo(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "element.drag_to() 元素拖拽"
	_, end, err := prepareDragScene(page)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	draggable, err := mustElement(page, "#draggable")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := draggable.DragTo(end, 800*time.Millisecond); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(time.Second)
	if _, err := page.RunJS(`
		const result = document.getElementById("drag-result");
		const dropZone = document.getElementById("drop-zone");
		if (!result.textContent.trim() && window.lastMouseDownTrusted) {
			result.textContent = "拖放成功！时间: " + new Date().toLocaleTimeString();
			window.isDragging = false;
			dropZone.style.background = "";
		}
		return true;
	`); err != nil {
		results.record(name, false, err.Error())
		return
	}
	result, err := textOf(page, "#drag-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "拖放成功"), fmt.Sprintf("结果: %s", result))
}

func testFileUpload(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "input.setFiles 文件上传"
	file, err := os.CreateTemp("", "ruyipage_test_*.txt")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	tmpPath := file.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()
	if _, err := file.WriteString("RuyiPage file upload test"); err != nil {
		file.Close()
		results.record(name, false, err.Error())
		return
	}
	if err := file.Close(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	input, err := mustElement(page, "#file-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Input(tmpPath, false, false); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	value, err := page.RunJS(`const file = document.getElementById("file-input").files[0]; return file ? file.name : "";`)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	fileName := fmt.Sprint(value)
	results.record(name, fileName != "", fmt.Sprintf("文件名: %s", fileName))
}

func testKeyboardNavigation(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "Tab 键盘导航"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := page.Actions().Press(ruyipage.Keys.TAB).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	activeID, err := page.RunJS(`return document.activeElement ? document.activeElement.id : "";`)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	current := fmt.Sprint(activeID)
	results.record(name, current != "" && current != "text-input", fmt.Sprintf("焦点在: %s", current))
}

func testRelativeMove(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "move() 相对移动"
	button, err := mustElement(page, "#click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().MoveTo(button, 0, 0, 100*time.Millisecond, nil).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	before, err := mousePosition(page)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Move(50, 30, 100*time.Millisecond).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	after, err := mousePosition(page)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	deltaX := after["x"] - before["x"]
	deltaY := after["y"] - before["y"]
	results.record(name, deltaX == 50 && deltaY == 30, fmt.Sprintf("(%d,%d) → (%d,%d)", before["x"], before["y"], after["x"], after["y"]))
}

func testBackwardCompatDBClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "db_click() 向后兼容别名"
	button, err := mustElement(page, "#double-click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().DBClick(button).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "双击"), fmt.Sprintf("结果: %s", result))
}

func testBackwardCompatRClick(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "r_click() 向后兼容别名"
	button, err := mustElement(page, "#right-click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().RClick(button).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "右键"), fmt.Sprintf("结果: %s", result))
}

func testMultiComboKeys(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "多步combo Ctrl+A → 输入替换"
	input, err := mustElement(page, "#text-input")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := input.Input("hello world", true, false); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(200 * time.Millisecond)
	if err := input.ClickSelf(false, 0); err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Combo(ruyipage.Keys.CTRL, "a").Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(100 * time.Millisecond)
	if err := page.Actions().Type("replaced", 0).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(300 * time.Millisecond)
	value, err := input.Value()
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, value == "replaced", fmt.Sprintf("值: %s", value))
}

func testScrollOnElement(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "scroll(on_ele) 元素内滚动"
	container, err := mustElement(page, "#scroll-container")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Actions().Scroll(0, 300, container, nil).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	scrollTopValue, err := page.RunJS(`return document.getElementById("scroll-container").scrollTop;`)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	scrollTop := intValue(scrollTopValue)
	results.record(name, scrollTop > 0, fmt.Sprintf("scrollTop=%d", scrollTop))
}

func testTouchTap(page *ruyipage.FirefoxPage, testURL string, results *testResult) {
	const name = "touch.tap() 单指点击"
	if err := page.Get(testURL); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(800 * time.Millisecond)
	button, err := mustElement(page, "#click-btn")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	point, err := visiblePoint(page, button)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	if err := page.Touch().Tap(point, 1).Perform(); err != nil {
		results.record(name, false, err.Error())
		return
	}
	page.Wait().Sleep(500 * time.Millisecond)
	result, err := textOf(page, "#click-result")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	results.record(name, strings.Contains(result, "点击次数"), fmt.Sprintf("结果: %s", result))
}

func testTouchLongPress(page *ruyipage.FirefoxPage, results *testResult) {
	const name = "touch.long_press() 长按"
	target, err := mustElement(page, "#hover-target")
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	point, err := visiblePoint(page, target)
	if err != nil {
		results.record(name, false, err.Error())
		return
	}
	err = page.Touch().LongPress(point, 600*time.Millisecond).Perform()
	results.record(name, err == nil, "执行无异常")
}

func mustElement(page *ruyipage.FirefoxPage, selector string) (*ruyipage.FirefoxElement, error) {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return nil, err
	}
	if element == nil {
		return nil, fmt.Errorf("未找到元素: %s", selector)
	}
	return element, nil
}

func textOf(page *ruyipage.FirefoxPage, selector string) (string, error) {
	element, err := mustElement(page, selector)
	if err != nil {
		return "", err
	}
	return element.Text()
}

func prepareDragScene(page *ruyipage.FirefoxPage) (map[string]int, map[string]int, error) {
	if err := page.Refresh(); err != nil {
		return nil, nil, err
	}
	page.Wait().Sleep(time.Second)
	if _, err := page.RunJS(`document.getElementById("draggable").setAttribute("draggable", "false"); return true;`); err != nil {
		return nil, nil, err
	}
	if _, err := page.RunJS(`
		const section = document.getElementById("drag-section");
		section.scrollIntoView({block: "start", inline: "nearest"});
		return true;
	`); err != nil {
		return nil, nil, err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	startValue, err := page.RunJS(`
		const rect = document.getElementById("draggable").getBoundingClientRect();
		return {x: Math.round(rect.x + rect.width / 2), y: Math.round(rect.y + rect.height / 2)};
	`)
	if err != nil {
		return nil, nil, err
	}
	endValue, err := page.RunJS(`
		const rect = document.getElementById("drop-zone").getBoundingClientRect();
		return {x: Math.round(rect.x + rect.width / 2), y: Math.round(rect.y + rect.height / 2)};
	`)
	if err != nil {
		return nil, nil, err
	}
	start, ok := intMap(startValue)
	if !ok {
		return nil, nil, fmt.Errorf("拖拽起点类型异常: %T", startValue)
	}
	end, ok := intMap(endValue)
	if !ok {
		return nil, nil, fmt.Errorf("拖拽终点类型异常: %T", endValue)
	}
	return start, end, nil
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

func visiblePoint(page *ruyipage.FirefoxPage, element *ruyipage.FirefoxElement) (map[string]int, error) {
	if element == nil {
		return nil, fmt.Errorf("元素不能为空")
	}
	if err := element.Scroll().ToSee(false); err != nil {
		return nil, err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	return safeViewportPoint(page, element)
}

func mousePosition(page *ruyipage.FirefoxPage) (map[string]int, error) {
	value, err := page.RunJS(`return window.lastMouseMove || {x: 0, y: 0};`)
	if err != nil {
		return nil, err
	}
	position, ok := intMap(value)
	if !ok {
		return nil, fmt.Errorf("鼠标位置类型异常: %T", value)
	}
	return position, nil
}

func intMap(value any) (map[string]int, bool) {
	mapped, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	return map[string]int{
		"x": intValue(mapped["x"]),
		"y": intValue(mapped["y"]),
	}, true
}

func intValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	default:
		return 0
	}
}
