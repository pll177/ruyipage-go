package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

const (
	cookieName = "demo_user"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试37: 三个隔离 user context 页面")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	userContexts := make([]string, 0, 3)
	tabIDs := make([]string, 0, 3)
	defer func() {
		for _, tabID := range tabIDs {
			tab, _ := page.GetTab(tabID, "", "")
			if tab != nil {
				_ = tab.Close(false)
			}
		}
		for _, userContext := range userContexts {
			_ = page.Contexts().RemoveUserContext(userContext)
		}
	}()

	fmt.Println("\n1. 创建三个 user context:")
	for index := 1; index <= 3; index++ {
		userContext, err := page.Contexts().CreateUserContext()
		if err != nil {
			return err
		}
		userContexts = append(userContexts, userContext)
		fmt.Printf("   user context %d: %s\n", index, userContext)
	}

	fmt.Println("\n2. 在三个 user context 中分别创建页面:")
	for index, userContext := range userContexts {
		tabID, err := page.Contexts().CreateTab(false, userContext, "")
		if err != nil {
			return err
		}
		tabIDs = append(tabIDs, tabID)
		fmt.Printf("   页面 %d: context=%s, userContext=%s\n", index+1, tabID, userContext)
	}

	tabs := make([]*ruyipage.FirefoxTab, 0, 3)
	for _, tabID := range tabIDs {
		tab, err := page.GetTab(tabID, "", "")
		if err != nil {
			return err
		}
		if tab == nil {
			return fmt.Errorf("未找到新建 tab: %s", tabID)
		}
		tabs = append(tabs, tab)
	}

	values := []string{"alpha_ctx", "beta_ctx", "gamma_ctx"}
	cookieURLs := []string{
		fmt.Sprintf("https://httpbin.org/cookies/set?%s=%s", cookieName, values[0]),
		fmt.Sprintf("https://httpbin.org/cookies/set?%s=%s", cookieName, values[1]),
		fmt.Sprintf("https://httpbin.org/cookies/set?%s=%s", cookieName, values[2]),
	}

	fmt.Println("\n3. 通过 httpbin 为每个页面设置不同 Cookie:")
	for index, tab := range tabs {
		if err := tab.Get(cookieURLs[index]); err != nil {
			return err
		}
		tab.Wait().Sleep(time.Second)
		fmt.Printf("   页面 %d 已访问: %s\n", index+1, cookieURLs[index])
	}

	fmt.Println("\n4. 分别验证三个页面 Cookie 内容:")
	for index, tab := range tabs {
		if err := printTabState(fmt.Sprintf("页面 %d", index+1), tab); err != nil {
			return err
		}
	}

	result1, err := readCookieValue(tabs[0])
	if err != nil {
		return err
	}
	result2, err := readCookieValue(tabs[1])
	if err != nil {
		return err
	}
	result3, err := readCookieValue(tabs[2])
	if err != nil {
		return err
	}

	fmt.Println("\n5. 隔离结果校验:")
	fmt.Printf("   页面1 %s = %s\n", cookieName, result1)
	fmt.Printf("   页面2 %s = %s\n", cookieName, result2)
	fmt.Printf("   页面3 %s = %s\n", cookieName, result3)
	if result1 != values[0] || result2 != values[1] || result3 != values[2] {
		return fmt.Errorf("三个页面的 Cookie 结果与预期不一致，隔离校验失败")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 三个 user context 页面 Cookie 已确认互相隔离")
	fmt.Println(strings.Repeat("=", 60))
	exampleutil.PrintManualKeepOpen(20*time.Second, "与 Python 示例一致，保留可见浏览器供人工观察隔离效果")
	page.Wait().Sleep(20 * time.Second)
	return nil
}

func readHTTPBinCookies(tab *ruyipage.FirefoxTab) (map[string]any, error) {
	if err := tab.Get("https://httpbin.org/cookies"); err != nil {
		return nil, err
	}
	tab.Wait().Sleep(time.Second)
	bodyText, err := tab.RunJS(`return document.body ? document.body.innerText : ""`)
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(fmt.Sprint(bodyText))
	if !strings.HasPrefix(text, "{") {
		return nil, fmt.Errorf("httpbin 返回内容不是 JSON: url=%s body=%q", safeURL(tab), truncate(text, 200))
	}
	data := map[string]any{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func printTabState(label string, tab *ruyipage.FirefoxTab) error {
	httpbinData, err := readHTTPBinCookies(tab)
	if err != nil {
		return err
	}
	fmt.Printf("   %s -> httpbin返回: %v\n", label, httpbinData)
	fmt.Printf("   %s -> API可见Cookie: %v\n", label, cookiesToDict(tab))
	return nil
}

func readCookieValue(tab *ruyipage.FirefoxTab) (string, error) {
	data, err := readHTTPBinCookies(tab)
	if err != nil {
		return "", err
	}
	cookies, _ := data["cookies"].(map[string]any)
	return fmt.Sprint(cookies[cookieName]), nil
}

func cookiesToDict(tab *ruyipage.FirefoxTab) map[string]string {
	cookies, err := tab.Cookies(true)
	if err != nil {
		return map[string]string{}
	}
	result := make(map[string]string, len(cookies))
	for _, cookie := range cookies {
		result[cookie.Name] = cookie.Value
	}
	return result
}

func safeURL(tab *ruyipage.FirefoxTab) string {
	if tab == nil {
		return ""
	}
	value, err := tab.URL()
	if err != nil {
		return ""
	}
	return value
}

func truncate(value string, size int) string {
	if len(value) <= size {
		return value
	}
	return value[:size]
}
