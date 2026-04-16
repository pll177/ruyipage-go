package main

import (
	"fmt"
	"sort"
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
	fmt.Println("测试7: JavaScript执行")
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

	fmt.Printf("\n1. 执行简单的JavaScript:\n")
	result, err := page.RunJS("return 1 + 2")
	if err != nil {
		return err
	}
	fmt.Printf("   1 + 2 = %v\n", result)

	fmt.Printf("\n2. 获取页面信息:\n")
	title, err := page.RunJS("return document.title")
	if err != nil {
		return err
	}
	currentURL, err := page.RunJS("return window.location.href")
	if err != nil {
		return err
	}
	fmt.Printf("   页面标题: %v\n", title)
	fmt.Printf("   页面URL: %v\n", currentURL)

	fmt.Printf("\n3. 修改页面内容:\n")
	if _, err := page.RunJS(`document.getElementById("main-title").textContent = "标题已被JS修改"`); err != nil {
		return err
	}
	newTitle, err := mustText(page, "#main-title")
	if err != nil {
		return err
	}
	fmt.Printf("   修改后的标题: %s\n", newTitle)

	fmt.Printf("\n4. 传递参数到JavaScript:\n")
	sum, err := page.RunJS("return arguments[0] + arguments[1]", 10, 20)
	if err != nil {
		return err
	}
	fmt.Printf("   10 + 20 = %v\n", sum)

	fmt.Printf("\n5. 在元素上执行JavaScript:\n")
	button, err := page.Ele("#click-btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if button == nil {
		return fmt.Errorf("未找到 #click-btn")
	}
	buttonText, err := button.RunJS("function() { return this.textContent; }")
	if err != nil {
		return err
	}
	fmt.Printf("   按钮文本: %v\n", buttonText)
	if _, err := button.RunJS(`function() { this.style.background = "red"; this.style.color = "white"; }`); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   ✓ 按钮样式已修改\n")

	fmt.Printf("\n6. 通过JS获取元素属性:\n")
	input, err := page.Ele("#text-input", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if input == nil {
		return fmt.Errorf("未找到 #text-input")
	}
	placeholder, err := input.RunJS("function() { return this.placeholder; }")
	if err != nil {
		return err
	}
	fmt.Printf("   输入框placeholder: %v\n", placeholder)

	fmt.Printf("\n7. 执行复杂的JavaScript:\n")
	list, err := page.RunJS(`
		(() => {
			const elements = document.querySelectorAll('.test-class');
			return Array.from(elements).map(el => el.textContent);
		})()
	`)
	if err != nil {
		return err
	}
	fmt.Printf("   所有.test-class元素的文本: %v\n", list)

	fmt.Printf("\n8. 通过JS修改输入框:\n")
	if _, err := page.RunJS(`document.getElementById("text-input").value = "JS设置的值"`); err != nil {
		return err
	}
	value, err := input.Value()
	if err != nil {
		return err
	}
	fmt.Printf("   输入框的值: %s\n", value)

	fmt.Printf("\n9. 通过JS触发事件:\n")
	if _, err := page.RunJS(`document.getElementById("click-btn").click()`); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	clickResult, err := mustText(page, "#click-result")
	if err != nil {
		return err
	}
	fmt.Printf("   点击结果: %s\n", clickResult)

	fmt.Printf("\n10. 获取元素的计算样式:\n")
	color, err := page.RunJS(`
		(() => {
			const elem = document.getElementById("click-btn");
			return window.getComputedStyle(elem).backgroundColor;
		})()
	`)
	if err != nil {
		return err
	}
	fmt.Printf("   按钮背景色: %v\n", color)

	fmt.Printf("\n11. 通过JS滚动到元素:\n")
	if _, err := page.RunJS(`document.getElementById("scroll-section").scrollIntoView()`); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   ✓ 已滚动到滚动测试区域\n")

	fmt.Printf("\n12. 通过JS创建新元素:\n")
	if _, err := page.RunJS(`
		const div = document.createElement("div");
		div.id = "js-created";
		div.textContent = "JavaScript创建的元素";
		div.style.padding = "10px";
		div.style.background = "#ffeb3b";
		document.body.appendChild(div);
	`); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	createdText, err := mustText(page, "#js-created")
	if err != nil {
		return err
	}
	fmt.Printf("   新元素文本: %s\n", createdText)

	fmt.Printf("\n13. 在sandbox中执行JavaScript:\n")
	sandboxValue, err := page.RunJSExprInSandbox(`
		(() => {
			globalThis.__ruyiSandboxCount = (globalThis.__ruyiSandboxCount || 0) + 1;
			return globalThis.__ruyiSandboxCount;
		})()
	`, "example07")
	if err != nil {
		return err
	}
	normalValue, err := page.RunJS("return globalThis.__ruyiSandboxCount || 0")
	if err != nil {
		return err
	}
	fmt.Printf("   sandbox 计数: %v\n", sandboxValue)
	fmt.Printf("   页面主世界计数: %v\n", normalValue)

	fmt.Printf("\n14. 获取脚本Realms:\n")
	realms, err := page.GetRealms("")
	if err != nil {
		return err
	}
	typeSet := map[string]struct{}{}
	for _, realm := range realms {
		typeName := realm.Type
		if typeName == "" {
			typeName = "unknown"
		}
		typeSet[typeName] = struct{}{}
	}
	realmTypes := make([]string, 0, len(typeSet))
	for typeName := range typeSet {
		realmTypes = append(realmTypes, typeName)
	}
	sort.Strings(realmTypes)
	fmt.Printf("   Realm数量: %d\n", len(realms))
	fmt.Printf("   Realm类型: %v\n", realmTypes)

	fmt.Printf("\n15. 通过高层脚本接口执行:\n")
	evalResult, err := page.EvalHandle(`JSON.stringify({title: document.title, ready: document.readyState})`, true)
	if err != nil {
		return err
	}
	fmt.Printf("   evaluate 返回类型: %s\n", evalResult.Type)
	fmt.Printf("   evaluate 结果: %v\n", evalResult.Result.Value)

	fmt.Printf("\n16. 预加载脚本 add/removePreloadScript:\n")
	preloadID, err := page.AddPreloadScript(`() => {
		window.__example07Preload = "preload-ready";
	}`)
	if err != nil {
		return err
	}
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	preloadValue, err := page.RunJS("return window.__example07Preload")
	if err != nil {
		return err
	}
	fmt.Printf("   preload脚本ID: %s\n", preloadID.ID)
	fmt.Printf("   preload注入结果: %v\n", preloadValue)
	if err := page.RemovePreloadScript(preloadID.ID); err != nil {
		return err
	}

	fmt.Printf("\n17. 移除预加载脚本后验证:\n")
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	removedValue, err := page.RunJS("return window.__example07Preload || null")
	if err != nil {
		return err
	}
	fmt.Printf("   移除后注入结果: %v\n", removedValue)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有JavaScript执行测试通过！")
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
