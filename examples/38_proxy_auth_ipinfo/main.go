package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const targetURL = "http://ipinfo.io/json"

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("示例38: 通过 fpfile 自动处理 HTTP 代理认证")
	fmt.Println(strings.Repeat("=", 60))

	proxyURL, err := exampleutil.FixedProxyURL()
	if err != nil {
		return err
	}
	outputDir, err := exampleutil.OutputDir("38_proxy_auth_ipinfo")
	if err != nil {
		return err
	}
	fpfilePath, err := exampleutil.WriteFixedProxyFPFile(outputDir)
	if err != nil {
		return err
	}

	opts := exampleutil.FixedVisibleOptions().
		WithProxy(proxyURL).
		WithFPFile(fpfilePath)

	page, err := ruyipage.NewFirefoxPage(opts)
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	fmt.Println("\n0. 已启用代理自动认证:")
	fmt.Printf("   Firefox: %s\n", exampleutil.FirefoxPath())
	host, port, username, _, _ := exampleutil.FixedProxyParts()
	fmt.Printf("   代理主机: %s:%d\n", host, port)
	if username != "" {
		fmt.Printf("   代理用户名: %s\n", username)
	}
	fmt.Printf("   代理URL: %s\n", proxyURL)
	fmt.Printf("   fpfile: %s\n", fpfilePath)
	fmt.Println("   认证信息将由内核从 fpfile 自动读取")

	fmt.Printf("\n1. 通过代理访问: %s\n", targetURL)
	if err := page.Get(targetURL); err != nil {
		return err
	}
	page.Wait().Sleep(2 * time.Second)

	title, _ := page.Title()
	fmt.Println("\n2. 页面标题:")
	fmt.Printf("   %s\n", title)

	bodyValue, err := page.RunJS(`return document.body ? document.body.innerText : ""`)
	if err != nil {
		return err
	}
	bodyText := strings.TrimSpace(fmt.Sprint(bodyValue))
	fmt.Println("\n3. 响应内容:")
	fmt.Println(bodyText)

	fmt.Println("\n4. 解析返回内容:")
	data := map[string]any{}
	if err := json.Unmarshal([]byte(bodyText), &data); err != nil {
		data = extractIPInfoFromText(bodyText)
	}
	if len(data) > 0 {
		fmt.Printf("   IP: %v\n", data["ip"])
		fmt.Printf("   城市: %v\n", data["city"])
		fmt.Printf("   地区: %v\n", data["region"])
		fmt.Printf("   国家: %v\n", data["country"])
		if data["status"] != nil || data["error"] != nil || data["message"] != nil {
			fmt.Printf("   状态: %v\n", data["status"])
			fmt.Printf("   错误: %v\n", data["error"])
			fmt.Printf("   消息: %v\n", data["message"])
		}
	} else {
		fmt.Println("   返回内容不是标准 JSON，可能是目标站限流或页面被中间页接管。")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("[OK] fpfile 代理认证示例执行完成")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func extractIPInfoFromText(text string) map[string]any {
	lines := strings.Split(text, "\n")
	fields := map[string]any{}
	keys := map[string]struct{}{
		"ip": {}, "city": {}, "region": {}, "country": {}, "loc": {},
		"org": {}, "postal": {}, "timezone": {}, "readme": {},
	}
	for index := 0; index < len(lines)-1; index++ {
		key := strings.TrimSpace(lines[index])
		if _, ok := keys[key]; !ok {
			continue
		}
		fields[key] = strings.Trim(strings.TrimSpace(lines[index+1]), `"`)
		index++
	}
	return fields
}
