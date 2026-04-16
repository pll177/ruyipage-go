package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const (
	zhipinURL     = "https://www.zhipin.com/hangzhou"
	zhipinKeyword = "爬虫工程师"
	zhipinAPI     = "https://www.zhipin.com/wapi/zpgeek/search/joblist.json"
)

var searchCoord = map[string]int{"x": 856, "y": 100}

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	page, err := ruyipage.NewFirefoxPage(
		exampleutil.FixedVisibleOptions().
			EnableXPathPicker(true).
			WithWindowSize(1600, 1100),
	)
	if err != nil {
		return err
	}
	defer func() {
		page.Listen().Stop()
		_ = page.Quit(0, false)
	}()

	if err := page.Get(zhipinURL); err != nil {
		return err
	}
	_, _ = page.Wait().DocLoaded(15 * time.Second)

	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("示例43: Boss 直聘杭州页按坐标搜索并打印 joblist.json")
	fmt.Printf("页面地址: %s\n", zhipinURL)
	fmt.Printf("搜索关键词: %s\n", zhipinKeyword)
	fmt.Printf("目标接口: %s\n", zhipinAPI)
	fmt.Println(strings.Repeat("=", 72))

	collector, err := page.Network().AddDataCollector([]string{"responseCompleted"}, []string{"response"}, 0)
	if err != nil {
		return err
	}
	defer func() {
		_ = collector.Remove()
	}()

	if err := page.Listen().Start(zhipinAPI, false, "POST"); err != nil {
		return err
	}
	if err := xyClickTypeEnter(page, zhipinKeyword); err != nil {
		return err
	}

	for {
		packet := page.Listen().Wait(30 * time.Second)
		if packet == nil {
			return fmt.Errorf("未捕获到 joblist.json 响应")
		}
		requestID := fmt.Sprint(packet.Request["request"])
		if requestID == "" {
			return fmt.Errorf("已捕获 joblist.json 响应，但缺少 request_id")
		}
		data, err := collector.Get(requestID, "response")
		if err != nil {
			return err
		}
		body := exampleutil.DecodeNetworkText(data)
		if body == "" {
			return fmt.Errorf("已捕获 joblist.json，但未取到响应体")
		}

		fmt.Printf("已捕获响应: %d %s\n", packet.Status, packet.URL)
		fmt.Println(body)
		if strings.Contains(body, "您的环境存在异常") {
			continue
		}
		break
	}
	return nil
}

func xyClickTypeEnter(page *ruyipage.FirefoxPage, keyword string) error {
	fmt.Printf("按坐标点击搜索框: (%d, %d)\n", searchCoord["x"], searchCoord["y"])
	if err := page.Actions().MoveTo(searchCoord, 0, 0, 0, "viewport").Click(nil, 1).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(300 * time.Millisecond)

	if err := page.Actions().Combo(ruyipage.Keys.CTRL, "a").Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(50 * time.Millisecond)
	if err := page.Actions().Press(ruyipage.Keys.DELETE).Perform(); err != nil {
		return err
	}
	page.Wait().Sleep(100 * time.Millisecond)

	if err := page.Actions().Type(keyword, 80*time.Millisecond).Press(ruyipage.Keys.ENTER).Perform(); err != nil {
		return err
	}
	fmt.Println("已输入并回车，等待页面 3 秒...")
	page.Wait().Sleep(3 * time.Second)
	return nil
}
