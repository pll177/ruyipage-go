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
	fmt.Println("测试10: 滚动操作")
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

	fmt.Printf("\n1. 滚动到页面底部:\n")
	if err := page.Actions().Scroll(0, 4000, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 已滚动到底部\n")

	fmt.Printf("\n2. 滚动到页面顶部:\n")
	if err := page.Actions().Scroll(0, -4000, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 已滚动到顶部\n")

	tableSection, err := page.Ele("#table-section", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if tableSection == nil {
		return fmt.Errorf("未找到 #table-section")
	}
	fmt.Printf("\n3. 滚动到特定元素:\n")
	for !tableSection.States().IsInViewport() {
		if err := page.Actions().Scroll(0, 500, nil, nil).Perform(); err != nil {
			return err
		}
		page.Wait().Sleep(200 * time.Millisecond)
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 已滚动到表格区域\n")

	fmt.Printf("\n4. 向下滚动500像素:\n")
	if err := page.Actions().Scroll(0, 500, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 已向下滚动\n")

	fmt.Printf("\n5. 向上滚动300像素:\n")
	if err := page.Actions().Scroll(0, -300, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 已向上滚动\n")

	formSection, err := page.Ele("#form-section", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if formSection == nil {
		return fmt.Errorf("未找到 #form-section")
	}
	fmt.Printf("\n6. 滚动到元素使其可见:\n")
	for !formSection.States().IsInViewport() {
		if err := page.Actions().Scroll(0, -400, nil, nil).Perform(); err != nil {
			return err
		}
		page.Wait().Sleep(200 * time.Millisecond)
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 元素已滚动到可见区域\n")

	scrollContainer, err := page.Ele("#scroll-container", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if scrollContainer == nil {
		return fmt.Errorf("未找到 #scroll-container")
	}
	fmt.Printf("\n7. 元素内部滚动:\n")
	for !scrollContainer.States().IsInViewport() {
		if err := page.Actions().Scroll(0, 400, nil, nil).Perform(); err != nil {
			return err
		}
		page.Wait().Sleep(200 * time.Millisecond)
	}
	page.Wait().Sleep(500 * time.Millisecond)
	if _, err := scrollContainer.RunJS(`function() { this.scrollTop = this.scrollHeight; }`); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 容器已滚动到底部\n")
	if _, err := scrollContainer.RunJS(`function() { this.scrollTop = 0; }`); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 容器已滚动到顶部\n")

	fmt.Printf("\n8. 滚动到容器内的元素:\n")
	scrollTarget, err := page.Ele("#scroll-target", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if scrollTarget == nil {
		return fmt.Errorf("未找到 #scroll-target")
	}
	if _, err := scrollTarget.RunJS(`function() { this.scrollIntoView({block: "nearest"}); }`); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 已滚动到目标元素\n")

	fmt.Printf("\n9. 获取页面滚动位置:\n")
	scrollPosition, err := page.RunJS("return {x: window.scrollX, y: window.scrollY}")
	if err != nil {
		return err
	}
	positionMap, _ := scrollPosition.(map[string]any)
	fmt.Printf("   当前滚动位置: X=%v, Y=%v\n", positionMap["x"], positionMap["y"])

	fmt.Printf("\n10. 平滑滚动到顶部:\n")
	if err := page.Actions().Scroll(0, -4000, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(2 * time.Second)
	fmt.Printf("   ✓ 平滑滚动完成\n")

	fmt.Printf("\n11. 滚动到页面中间:\n")
	if err := page.Actions().Scroll(0, 500, nil, nil).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 已滚动到指定位置\n")

	networkSection, err := page.Ele("#network-section", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if networkSection == nil {
		return fmt.Errorf("未找到 #network-section")
	}
	fmt.Printf("\n12. 将元素滚动到视图:\n")
	for !networkSection.States().IsInViewport() {
		if err := page.Actions().Scroll(0, 450, nil, nil).Perform(); err != nil {
			return err
		}
		page.Wait().Sleep(200 * time.Millisecond)
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 元素已滚动到视图\n")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有滚动操作测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}
