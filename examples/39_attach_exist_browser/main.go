package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const attachTargetURL = "https://0xshoulderlab.site/automation"

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("示例39: 接管已启动浏览器")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("1. 请先手工启动 Firefox 或 Firefox 指纹浏览器")
	fmt.Println("   先执行以下固定命令，再运行本示例:")
	fmt.Printf("   %s\n", exampleutil.FixedAttachCommand)
	fmt.Println("   本示例不会自动启动浏览器，只会扫描并接管已打开实例。")
	fmt.Println("   若浏览器把固定端口改成随机端口，本示例会自动降级为按进程特征扫描。")

	fmt.Println("\n2. 按进程特征自动接管...")
	page, err := ruyipage.AutoAttachExistingBrowserByProcess("", 200*time.Millisecond, 16, 1, true)
	if err != nil {
		return err
	}

	address := ""
	if browser := page.Browser(); browser != nil {
		address = browser.Address()
	}
	title, _ := page.Title()
	currentURL, _ := page.URL()
	fmt.Printf("   已自动接入: %s\n", address)
	fmt.Printf("   当前标题: %s\n", title)
	fmt.Printf("   当前地址: %s\n", currentURL)

	if err := page.Get(attachTargetURL); err != nil {
		return err
	}
	page.Wait().Sleep(2 * time.Second)
	jumpedURL, _ := page.URL()
	fmt.Printf("   接管后已跳转到: %s\n", jumpedURL)
	fmt.Printf("   当前标签页数量: %d\n", page.TabsCount())

	fmt.Println("\n3. 打印可见标签页:")
	for index, tabID := range page.TabIDs() {
		tab, err := page.GetTab(index+1, "", "")
		if err != nil || tab == nil {
			fmt.Printf("   [%d] <获取失败> | context=%s\n", index+1, tabID)
			continue
		}
		tabTitle, _ := tab.Title()
		tabURL, _ := tab.URL()
		fmt.Printf("   [%d] %s | %s\n", index+1, tabTitle, tabURL)
	}

	fmt.Println("\n示例结束。这里不自动关闭浏览器，便于继续手工观察。")
	return nil
}
