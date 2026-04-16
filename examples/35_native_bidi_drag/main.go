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
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("测试 35: 原生 BiDi 拖拽")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Actions().ReleaseAll()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	testURL, err := exampleutil.TestPageURL("native_bidi_drag_test.html")
	if err != nil {
		return err
	}
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(800 * time.Millisecond)
	exampleutil.AddCheck(&results, "测试页加载", "成功", testURL)

	source, err := page.Ele("#drag-source", 1, 5*time.Second)
	if err != nil {
		return err
	}
	target, err := page.Ele("#drop-target", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if source == nil || target == nil {
		return fmt.Errorf("未找到 source 或 target 元素")
	}

	if err := page.Actions().Drag(source, target, 640*time.Millisecond, 16).Perform(); err != nil {
		return err
	}
	_ = page.Actions().ReleaseAll()
	page.Wait().Sleep(800 * time.Millisecond)

	stateValue, err := page.RunJS(`return window.nativeBidiDragState`)
	if err != nil {
		return err
	}
	state, _ := stateValue.(map[string]any)
	resultElement, err := page.Ele("#result", 1, 5*time.Second)
	if err != nil {
		return err
	}
	resultText := ""
	if resultElement != nil {
		resultText, _ = resultElement.Text()
	}

	if resultText == "拖拽成功" {
		exampleutil.AddCheck(&results, "拖拽结果文本", "成功", resultText)
	} else {
		exampleutil.AddCheck(&results, "拖拽结果文本", "失败", resultText)
	}
	if stateBool(state, "dropped") {
		exampleutil.AddCheck(&results, "HTML5 drop 命中", "成功", "dropped=True")
	} else {
		exampleutil.AddCheck(&results, "HTML5 drop 命中", "失败", fmt.Sprintf("state=%v", state))
	}
	if stateBool(state, "enteredTarget") {
		exampleutil.AddCheck(&results, "目标区域进入", "成功", "enteredTarget=True")
	} else {
		exampleutil.AddCheck(&results, "目标区域进入", "失败", fmt.Sprintf("state=%v", state))
	}
	if stateBool(state, "trustedMouseDown") && stateBool(state, "trustedMouseUp") {
		exampleutil.AddCheck(&results, "isTrusted mouse down/up", "成功", "mouseDown/mouseUp 均为 trusted")
	} else {
		exampleutil.AddCheck(&results, "isTrusted mouse down/up", "失败", fmt.Sprintf("down=%v up=%v", state["trustedMouseDown"], state["trustedMouseUp"]))
	}
	moveCount := stateInt(state, "trustedMoveCount")
	if moveCount > 0 {
		exampleutil.AddCheck(&results, "isTrusted move count", "成功", fmt.Sprintf("trustedMoveCount=%d", moveCount))
	} else {
		exampleutil.AddCheck(&results, "isTrusted move count", "失败", fmt.Sprintf("trustedMoveCount=%d", moveCount))
	}
	if state["lastClientX"] != nil && state["lastClientY"] != nil {
		exampleutil.AddCheck(&results, "最后指针坐标", "成功", fmt.Sprintf("(%v, %v)", state["lastClientX"], state["lastClientY"]))
	} else {
		exampleutil.AddCheck(&results, "最后指针坐标", "跳过", "页面未记录最后指针坐标")
	}

	exampleutil.PrintChecks(results)
	return nil
}

func stateBool(values map[string]any, key string) bool {
	if values == nil {
		return false
	}
	value, _ := values[key].(bool)
	return value
}

func stateInt(values map[string]any, key string) int {
	if values == nil {
		return 0
	}
	switch typed := values[key].(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}
