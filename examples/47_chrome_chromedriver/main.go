package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

// 示例47：自动下载 chromedriver 并接管一个已经用 --remote-debugging-port 启动的 Chrome。
//
// 先手工启动 Chrome：
//
//	"C:\Program Files\Google\Chrome\Application\chrome.exe" ^
//	    --remote-debugging-port=9222 ^
//	    --user-data-dir=D:\chrome_userdir
//
// 注意：--user-data-dir 必须是独立目录，不能是日常 Chrome 的 profile，否则 CDP 不会开放。
func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("示例47: Chrome + chromedriver (自动下载 + 接管已打开浏览器)")
	fmt.Println(strings.Repeat("=", 60))

	opts := ruyipage.NewChromeOptions().
		WithChromedriverVersion("144").
		WithDebuggerAddress("127.0.0.1:9222")

	page, err := ruyipage.NewChromePage(opts)
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(time.Second)
		_ = page.Quit(5*time.Second, false)
	}()

	if err := page.Get("https://example.com"); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)

	title, err := page.Title()
	if err != nil {
		return err
	}
	fmt.Printf("标题: %s\n", title)

	h1, err := page.Ele("tag:h1", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if h1 == nil {
		return fmt.Errorf("未找到 h1 元素")
	}
	text, err := h1.Text()
	if err != nil {
		return err
	}
	fmt.Printf("H1文本: %s\n", text)

	fmt.Println("\n✓ Chrome + chromedriver BiDi 跑通（接管 127.0.0.1:9222）")
	return nil
}
