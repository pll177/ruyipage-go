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
	fmt.Println("测试36: 原生 BiDi Select（仅原生）")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(time.Second)
		_ = page.Quit(0, false)
	}()

	testURL, err := exampleutil.TestPageURL("native_bidi_select_test.html")
	if err != nil {
		return err
	}
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)

	selectElement, err := page.Ele("#single-select", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if selectElement == nil {
		return fmt.Errorf("未找到 #single-select")
	}

	fmt.Println("\n1. 初始状态:")
	initialState, err := readSelectState(page, selectElement)
	if err != nil {
		return err
	}
	fmt.Printf("   value: %s\n", initialState["value"])
	fmt.Printf("   selected_option: %v\n", initialState["selected_option"])
	fmt.Printf("   focused: %v\n", initialState["focused"])

	fmt.Println("\n2. 手动焦点探测（原生 click + 一步键盘）:")
	if err := selectElement.ClickSelf(false, 0); err != nil {
		return err
	}
	page.Wait().Sleep(100 * time.Millisecond)
	if err := page.Actions().KeyDown(ruyipage.Keys.DOWN).KeyUp(ruyipage.Keys.DOWN).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(100 * time.Millisecond)
	probeState, err := readSelectState(page, selectElement)
	if err != nil {
		return err
	}
	fmt.Printf("   probe value: %s\n", probeState["value"])
	fmt.Printf("   probe focused: %v\n", probeState["focused"])
	fmt.Printf("   probe changeCount: %v\n", probeState["change_count"])

	fmt.Println("\n3. 使用 selector native_only 选择 opt2:")
	ok, err := selectElement.Select().ByValue("opt2", "native_only")
	if err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)
	finalState, err := readSelectState(page, selectElement)
	if err != nil {
		return err
	}
	fmt.Printf("   native_only结果: %v\n", ok)
	fmt.Printf("   value: %s\n", finalState["value"])
	fmt.Printf("   selected_option: %v\n", finalState["selected_option"])
	fmt.Printf("   focused: %v\n", finalState["focused"])
	fmt.Printf("   trustedChange: %v\n", finalState["trusted_change"])
	fmt.Printf("   changeCount: %v\n", finalState["change_count"])

	fmt.Println("\n4. 页面事件日志:")
	if events, ok := finalState["events"].([]any); ok {
		for _, line := range events {
			fmt.Printf("   %v\n", line)
		}
	}

	if !ok || finalState["value"] != "opt2" {
		return fmt.Errorf("native_only 未成功切换到 opt2")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 原生 BiDi Select 测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func readSelectState(page *ruyipage.FirefoxPage, selectElement *ruyipage.FirefoxElement) (map[string]any, error) {
	stateValue, err := page.RunJS(`return window.nativeBidiSelectState`)
	if err != nil {
		return nil, err
	}
	state, _ := stateValue.(map[string]any)
	if state == nil {
		state = map[string]any{}
	}
	value, err := selectElement.Value()
	if err != nil {
		return nil, err
	}
	state["value"] = value
	state["selected_option"] = selectElement.Select().SelectedOption()
	return state, nil
}
