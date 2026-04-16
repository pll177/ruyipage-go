package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

const (
	shopeeURL          = "https://shopee.tw/search?keyword=iphone%E6%89%8B%E9%8C%B6"
	defaultShopeeDir   = `F:\ruyipage\user1`
	itemsXPath         = "xpath://section[1]/ul[1]/li"
	nextPageXPath      = "xpath:/html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[1]/div[1]/div[1]/div[1]/div[2]/section[1]/div[1]/nav[1]/a[5]"
	shopeeCollectPages = 3
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	userDir := exampleutil.ResolveEnvPath("RUYIPAGE_EXAMPLE_USER_DIR", defaultShopeeDir)
	page, err := ruyipage.NewFirefoxPage(
		exampleutil.FixedVisibleOptions().
			WithUserDir(userDir).
			EnableXPathPicker(true).
			WithWindowSize(1600, 1100),
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	if err := page.Get(shopeeURL); err != nil {
		return err
	}
	_, _ = page.Wait().DocLoaded(20 * time.Second)
	page.Wait().Sleep(5 * time.Second)

	title, _ := page.Title()
	pageURL, _ := page.URL()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("示例44: Shopee 搜索页 + user_dir + XPath picker + 3 页采集")
	fmt.Printf("页面地址: %s\n", pageURL)
	fmt.Printf("标题: %s\n", title)
	fmt.Printf("user_dir: %s\n", userDir)
	fmt.Println("XPath picker: 已启用")
	fmt.Printf("采集 XPath: %s\n", itemsXPath)
	fmt.Printf("翻页 XPath: %s\n", nextPageXPath)
	fmt.Println(strings.Repeat("=", 72))

	allRows := make([]map[string]any, 0, 120)
	for pageNo := 1; pageNo <= shopeeCollectPages; pageNo++ {
		rows, err := collectPageItems(page, pageNo)
		if err != nil {
			return err
		}
		allRows = append(allRows, rows...)

		fmt.Printf("\n[第 %d 页] 共采集 %d 项\n", pageNo, len(rows))
		for _, row := range rows {
			text := strings.TrimSpace(fmt.Sprint(row["text"]))
			if text == "" {
				text = "<空>"
			}
			fmt.Printf("  %02d. %s\n", row["index"], text)
		}

		if pageNo < shopeeCollectPages {
			if err := gotoNextPage(page, pageNo); err != nil {
				return err
			}
		}
	}

	fmt.Printf("\n总计采集 %d 项，来自 %d 页。\n", len(allRows), shopeeCollectPages)
	exampleutil.PrintManualKeepOpen(1000*time.Second, "与 Python 示例一致，保留浏览器长时间打开供 XPath picker 与翻页结果人工观察")
	page.Wait().Sleep(1000 * time.Second)
	return nil
}

func collectPageItems(page *ruyipage.FirefoxPage, pageNo int) ([]map[string]any, error) {
	fmt.Printf("第 %d 页先滚动到底部...\n", pageNo)
	if err := page.Scroll().ToBottom(); err != nil {
		return nil, err
	}
	page.Wait().Sleep(5 * time.Second)

	if _, err := page.Wait().Ele(itemsXPath, 20*time.Second); err != nil {
		return nil, err
	}
	page.Wait().Sleep(2 * time.Second)

	items, err := page.Eles(itemsXPath, 2*time.Second)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, len(items))
	for index, item := range items {
		text, _ := item.Text()
		rows = append(rows, map[string]any{
			"page":  pageNo,
			"index": index + 1,
			"text":  text,
		})
	}
	return rows, nil
}

func gotoNextPage(page *ruyipage.FirefoxPage, currentPage int) error {
	nextButton, err := page.Ele(nextPageXPath, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if nextButton == nil {
		return fmt.Errorf("第 %d 页未找到下一页按钮: %s", currentPage, nextPageXPath)
	}
	fmt.Printf("第 %d 页采集完成，点击下一页...\n", currentPage)
	if err := nextButton.ClickSelf(false, 0); err != nil {
		return err
	}
	if err := page.Wait().LoadComplete(20 * time.Second); err != nil {
		return err
	}
	page.Wait().Sleep(3 * time.Second)
	return nil
}
