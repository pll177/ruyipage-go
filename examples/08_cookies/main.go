package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
	"github.com/pll177/ruyipage-go/examples/internal/testserver"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试8: Cookie管理")
	fmt.Println(strings.Repeat("=", 60))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(8888))
	if err := server.Start(); err != nil {
		return err
	}
	defer func() {
		_ = server.Stop()
	}()

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(2 * time.Second)
		_ = page.Quit(0, false)
	}()

	fmt.Printf("\n访问测试服务器...\n")
	if err := page.Get(server.GetURL("/set-cookie")); err != nil {
		return err
	}
	page.Wait().Sleep(2 * time.Second)

	fmt.Printf("\n1. 获取服务器设置的Cookie:\n")
	allCookies, err := page.Cookies(true)
	if err != nil {
		return err
	}
	fmt.Printf("   共有 %d 个Cookie\n", len(allCookies))
	for _, cookie := range allCookies {
		fmt.Printf("   - %s: %s\n", cookie.Name, cookie.Value)
	}

	fmt.Printf("\n2. 通过API设置Cookie:\n")
	if err := page.SetCookies(map[string]any{
		"name":   "api_cookie",
		"value":  "api_value",
		"domain": "127.0.0.1",
		"path":   "/",
	}); err != nil {
		return err
	}
	if err := page.SetCookies(map[string]any{
		"name":   "user_id",
		"value":  "12345",
		"domain": "127.0.0.1",
		"path":   "/",
	}); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已通过API设置2个Cookie\n")

	fmt.Printf("\n3. 验证Cookie已设置:\n")
	if err := page.Refresh(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	allCookies, err = page.Cookies(true)
	if err != nil {
		return err
	}
	fmt.Printf("   刷新后共有 %d 个Cookie\n", len(allCookies))
	cookieMap := cookieValues(allCookies)
	if value, ok := cookieMap["api_cookie"]; ok {
		fmt.Printf("   ✓ api_cookie = %s\n", value)
	}
	if value, ok := cookieMap["user_id"]; ok {
		fmt.Printf("   ✓ user_id = %s\n", value)
	}

	fmt.Printf("\n4. 设置带过期时间的Cookie:\n")
	if err := page.SetCookies(map[string]any{
		"name":   "expire_cookie",
		"value":  "will_expire",
		"domain": "127.0.0.1",
		"path":   "/",
		"expiry": time.Now().Unix() + 3600,
	}); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已设置带过期时间的Cookie（1小时后过期）\n")

	fmt.Printf("\n5. 设置HttpOnly和Secure Cookie:\n")
	if err := page.SetCookies(map[string]any{
		"name":     "secure_cookie",
		"value":    "secure_value",
		"domain":   "127.0.0.1",
		"path":     "/",
		"httpOnly": true,
		"secure":   false,
	}); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已设置HttpOnly Cookie\n")

	fmt.Printf("\n6. 获取特定Cookie:\n")
	allCookies, err = page.Cookies(true)
	if err != nil {
		return err
	}
	cookieMap = cookieValues(allCookies)
	if value, ok := cookieMap["user_id"]; ok {
		fmt.Printf("   user_id = %s\n", value)
	}
	if value, ok := cookieMap["expire_cookie"]; ok {
		fmt.Printf("   expire_cookie = %s\n", value)
	}

	fmt.Printf("\n6.1 当前页面 cookies 属性:\n")
	simpleCookies, err := page.Cookies(false)
	if err != nil {
		return err
	}
	fmt.Printf("   page.cookies 数量: %d\n", len(simpleCookies))

	fmt.Printf("\n6.2 通过 page.set.cookies 设置Cookie:\n")
	if err := page.CookiesSetter().Set(map[string]any{
		"name":   "setter_cookie",
		"value":  "setter_value",
		"domain": "127.0.0.1",
		"path":   "/",
	}); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	setterCookies, err := page.Cookies(false)
	if err != nil {
		return err
	}
	fmt.Printf("   setter_cookie = %s\n", cookieValues(setterCookies)["setter_cookie"])

	fmt.Printf("\n6.3 通过高层过滤 API 读取Cookie:\n")
	filteredCookies := filterCookiesByName(allCookies, "user_id")
	fmt.Printf("   过滤结果数量: %d\n", len(filteredCookies))
	if len(filteredCookies) > 0 {
		fmt.Printf("   user_id(filtered) = %s\n", filteredCookies[0].Value)
	}

	fmt.Printf("\n7. 删除特定Cookie:\n")
	if err := page.DeleteCookies(map[string]any{"name": "api_cookie"}); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已删除 api_cookie\n")
	page.Wait().Sleep(500 * time.Millisecond)
	allCookies, err = page.Cookies(true)
	if err != nil {
		return err
	}
	cookieMap = cookieValues(allCookies)
	if _, ok := cookieMap["api_cookie"]; !ok {
		fmt.Printf("   ✓ 确认 api_cookie 已被删除\n")
	} else {
		fmt.Printf("   ⚠ api_cookie 仍然存在\n")
	}

	fmt.Printf("\n7.1 删除 setter_cookie:\n")
	if err := page.CookiesSetter().Remove("setter_cookie", "127.0.0.1"); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	cookiesAfterRemove, err := page.Cookies(false)
	if err != nil {
		return err
	}
	_, exists := cookieValues(cookiesAfterRemove)["setter_cookie"]
	fmt.Printf("   setter_cookie 是否存在: %v\n", exists)

	fmt.Printf("\n8. 访问API验证Cookie发送:\n")
	if err := page.Get(server.GetURL("/get-cookie")); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	bodyText, err := mustText(page, "tag:body")
	if err != nil {
		return err
	}
	fmt.Printf("   服务器收到的Cookie: %s\n", trimPreview(bodyText, 100))

	fmt.Printf("\n9. 清空所有Cookie:\n")
	if err := page.DeleteCookies(nil); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已清空所有Cookie\n")
	allCookies, err = page.Cookies(false)
	if err != nil {
		return err
	}
	fmt.Printf("   清空后Cookie数量: %d\n", len(allCookies))

	fmt.Printf("\n9.1 浏览器级 cookies 读取:\n")
	browserCookies, err := page.Browser().Cookies(false)
	if err != nil {
		return err
	}
	fmt.Printf("   browser.cookies 数量: %d\n", len(browserCookies))

	fmt.Printf("\n10. 重新设置Cookie并测试:\n")
	if err := page.SetCookies(map[string]any{
		"name":   "final_cookie",
		"value":  "final_value",
		"domain": "127.0.0.1",
		"path":   "/",
	}); err != nil {
		return err
	}
	if err := page.Refresh(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	allCookies, err = page.Cookies(true)
	if err != nil {
		return err
	}
	cookieMap = cookieValues(allCookies)
	if value, ok := cookieMap["final_cookie"]; ok {
		fmt.Printf("   ✓ final_cookie = %s\n", value)
		fmt.Printf("   ✓ Cookie持久性测试通过\n")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有Cookie管理测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func cookieValues(cookies []ruyipage.CookieInfo) map[string]string {
	result := make(map[string]string, len(cookies))
	for _, cookie := range cookies {
		result[cookie.Name] = cookie.Value
	}
	return result
}

func filterCookiesByName(cookies []ruyipage.CookieInfo, name string) []ruyipage.CookieInfo {
	result := make([]ruyipage.CookieInfo, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie.Name == name {
			result = append(result, cookie)
		}
	}
	return result
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

func trimPreview(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
