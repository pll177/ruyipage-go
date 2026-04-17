package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const actionVisualAutoPortStart = 19334

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	testURL, err := exampleutil.TestPageURL("action_visual_mouse_only.html")
	if err != nil {
		return err
	}

	page, err := ruyipage.NewFirefoxPage(
		exampleutil.FixedVisibleOptions().
			ActionVisualEnabled(true).
			WithWindowSize(1400, 900).
			AutoPortEnabled(true).
			WithAutoPortStart(actionVisualAutoPortStart),
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(5*time.Second, false)
	}()

	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)

	clickButton, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	doubleClickButton, err := page.Ele("#double-click-btn", 1, 5*time.Second)
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
	jsClickButton, err := page.Ele("#js-click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	emailInput, err := page.Ele("#email-input", 1, 5*time.Second)
	if err != nil {
		return err
	}

	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("示例42_2: action_visual 鼠标行为可视化")
	fmt.Printf("页面地址: %s\n", testURL)
	fmt.Println("已开启: ActionVisualEnabled(true)")
	fmt.Println("将依次演示：轨迹、点击、拟人移动、拖拽、JS click、JS input 高亮")
	fmt.Println(strings.Repeat("=", 72))

	if err := page.Actions().MoveTo(map[string]int{"x": 180, "y": 180}, 0, 0, 0, nil).Perform(); err != nil {
		return err
	}
	step(page, "move_to(180,180)")

	if err := page.Actions().HumanMove(map[string]int{"x": 760, "y": 220}, "").Perform(); err != nil {
		return err
	}
	step(page, "human_move 到右上区域")

	if err := page.Actions().Click(clickButton, 1).Perform(); err != nil {
		return err
	}
	step(page, "点击标准按钮")

	if err := page.Actions().DoubleClick(doubleClickButton).Perform(); err != nil {
		return err
	}
	step(page, "双击按钮")

	if err := hoverTarget.Hover(); err != nil {
		return err
	}
	step(page, "元素 hover")

	if err := draggable.DragTo(dropZone, 700*time.Millisecond); err != nil {
		return err
	}
	step(page, "拖拽到目标区域")

	if err := jsClickButton.ClickSelf(true, 0); err != nil {
		return err
	}
	step(page, "JS click 反馈")

	if err := emailInput.Input("js-demo@ruyi.dev", true, true); err != nil {
		return err
	}
	step(page, "JS input 高亮反馈")

	exampleutil.PrintManualKeepOpen(10*time.Second, "观察 action_visual 的轨迹、点击圈和目标高亮")
	page.Wait().Sleep(10 * time.Second)
	return nil
}

func step(page *ruyipage.FirefoxPage, title string) {
	fmt.Printf("  >> %s\n", title)
	page.Wait().Sleep(1200 * time.Millisecond)
}
