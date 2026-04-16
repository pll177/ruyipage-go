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
	fmt.Println("示例01: 一步上手")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.Launch()
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(time.Second)
		_ = page.Quit(0, false)
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

	h1Text, err := h1.Text()
	if err != nil {
		return err
	}
	fmt.Printf("H1文本: %s\n", h1Text)

	fmt.Println("\n✓ 快速上手完成")
	return nil
}
