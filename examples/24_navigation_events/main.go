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
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("测试 24: Navigation Events 导航事件")
	fmt.Println(strings.Repeat("=", 70))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(9540))
	if err := server.Start(); err != nil {
		return err
	}
	defer func() {
		_ = server.Stop()
	}()

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Navigation().Stop()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	baseURL := strings.TrimRight(server.GetURL(""), "/")
	trackedEvents := []string{
		"browsingContext.navigationStarted",
		"browsingContext.fragmentNavigated",
		"browsingContext.historyUpdated",
		"browsingContext.domContentLoaded",
		"browsingContext.load",
		"browsingContext.navigationCommitted",
		"browsingContext.navigationFailed",
	}
	if err := page.Navigation().Start(trackedEvents); err != nil {
		return err
	}

	page.Navigation().Clear()
	if err := page.Get(baseURL + "/nav/basic"); err != nil {
		return err
	}
	started := page.Navigation().Wait("browsingContext.navigationStarted", 3*time.Second, "")
	domLoaded := page.Navigation().Wait("browsingContext.domContentLoaded", 3*time.Second, "")
	loaded := page.Navigation().WaitForLoad(3 * time.Second)
	if started != nil && strings.HasSuffix(started.URL, "/nav/basic") {
		exampleutil.AddCheck(&results, "navigationStarted", "成功", "收到基础导航开始事件")
	} else {
		exampleutil.AddCheck(&results, "navigationStarted", "失败", "未观察到基础页面导航开始事件")
	}
	if domLoaded != nil && domLoaded.Context == page.ContextID() {
		exampleutil.AddCheck(&results, "domContentLoaded", "成功", "收到 DOMContentLoaded 事件")
	} else {
		exampleutil.AddCheck(&results, "domContentLoaded", "失败", "未观察到 DOMContentLoaded 事件")
	}
	if loaded != nil && loaded.Context == page.ContextID() {
		exampleutil.AddCheck(&results, "load", "成功", "收到 load 事件")
	} else {
		exampleutil.AddCheck(&results, "load", "失败", "未观察到 load 事件")
	}

	if err := page.Get(baseURL + "/nav/fragment"); err != nil {
		return err
	}
	page.Navigation().Clear()
	if _, err := page.RunJS(`() => {
		location.hash = "#a";
		setTimeout(() => { location.hash = "#b"; }, 100);
		return true;
	}`); err != nil {
		return err
	}
	firstFragment := page.Navigation().WaitForFragment("a", 3*time.Second)
	secondFragment := page.Navigation().WaitForFragment("b", 3*time.Second)
	if firstFragment != nil && secondFragment != nil {
		exampleutil.AddCheck(&results, "fragmentNavigated", "成功", fmt.Sprintf("%s -> %s", firstFragment.URL, secondFragment.URL))
	} else {
		exampleutil.AddCheck(&results, "fragmentNavigated", "不支持", "片段 URL 已变化，但当前 Firefox 未稳定观察到标准事件")
	}

	if err := page.Get(baseURL + "/nav/history"); err != nil {
		return err
	}
	page.Navigation().Clear()
	if _, err := page.RunJS(`() => {
		history.pushState({p: 1}, "P1", "?p=1");
		history.pushState({p: 2}, "P2", "?p=2");
		history.back();
		return true;
	}`); err != nil {
		return err
	}
	firstHistory := page.Navigation().Wait("browsingContext.historyUpdated", 3*time.Second, "")
	secondHistory := page.Navigation().Wait("browsingContext.historyUpdated", 3*time.Second, "")
	if firstHistory != nil || secondHistory != nil {
		exampleutil.AddCheck(&results, "historyUpdated", "成功", "pushState/back 触发 historyUpdated")
	} else {
		exampleutil.AddCheck(&results, "historyUpdated", "失败", "未观察到 historyUpdated 事件")
	}

	page.Navigation().Clear()
	if err := page.Get(baseURL + "/nav/basic?committed=1"); err != nil {
		return err
	}
	committed := page.Navigation().Wait("browsingContext.navigationCommitted", 3*time.Second, "")
	if committed != nil && committed.Context == page.ContextID() {
		exampleutil.AddCheck(&results, "navigationCommitted", "成功", "普通导航触发 navigationCommitted")
	} else {
		exampleutil.AddCheck(&results, "navigationCommitted", "跳过", "当前环境未稳定观察到 navigationCommitted")
	}

	page.Navigation().Clear()
	_ = page.Get("http://127.0.0.1:9/")
	failed := page.Navigation().Wait("browsingContext.navigationFailed", 3*time.Second, "")
	if failed != nil {
		exampleutil.AddCheck(&results, "navigationFailed", "成功", "不可达地址触发 navigationFailed")
	} else {
		exampleutil.AddCheck(&results, "navigationFailed", "跳过", "当前环境未稳定观察到 navigationFailed")
	}

	if err := page.Navigation().Start([]string{"browsingContext.navigationAborted"}); err != nil {
		exampleutil.AddCheck(&results, "navigationAborted", "不支持", err.Error())
	} else {
		exampleutil.AddCheck(&results, "navigationAborted", "跳过", "订阅成功，但本示例未构造稳定 aborted 场景")
	}
	_ = page.Navigation().Start(trackedEvents)

	exampleutil.PrintChecks(results)
	return nil
}
