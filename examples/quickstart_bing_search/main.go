package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	if err := page.Get("https://cn.bing.com/"); err != nil {
		return err
	}
	searchBox, err := page.Ele("#sb_form_q", 1, 10*time.Second)
	if err != nil {
		return err
	}
	if searchBox == nil {
		return fmt.Errorf("未找到 Bing 搜索框")
	}
	if err := searchBox.Input("小肩膀教育", true, false); err != nil {
		return err
	}
	if err := page.Actions().Press(ruyipage.Keys.ENTER).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(3 * time.Second)

	for pageNo := 1; pageNo <= 3; pageNo++ {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("第 %d 页\n", pageNo)
		fmt.Println(strings.Repeat("=", 80))

		items, err := page.Eles("css:#b_results > li.b_algo", 5*time.Second)
		if err != nil {
			return err
		}
		for index, item := range items {
			titleEle, _ := item.Ele("css:h2 a", 1, 2*time.Second)
			if titleEle == nil {
				continue
			}
			title, _ := titleEle.Text()
			title = strings.Join(strings.Fields(title), " ")
			itemURL, _ := titleEle.Attr("href")

			descEle, _ := item.Ele("css:.b_caption p", 1, time.Second)
			content := ""
			if descEle != nil {
				content, _ = descEle.Text()
			} else {
				content, _ = item.Text()
			}
			content = strings.Join(strings.Fields(content), " ")

			fmt.Printf("%d. %s\n", index+1, title)
			fmt.Printf("   URL: %s\n", itemURL)
			fmt.Printf("   内容: %s\n", content)
		}

		if pageNo < 3 {
			nextButton, err := page.Ele("css:a.sb_pagN", 1, 5*time.Second)
			if err != nil {
				return err
			}
			if nextButton == nil {
				break
			}
			if err := nextButton.ClickSelf(false, 0); err != nil {
				return err
			}
			page.Wait().Sleep(2 * time.Second)
		}
	}
	return nil
}
