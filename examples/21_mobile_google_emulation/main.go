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
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("测试 21-Mobile: 移动端模拟访问 Google")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 6)
	if err := page.Get("about:blank"); err != nil {
		return err
	}

	iphoneUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	touch := false
	support := page.Emulation().ApplyMobilePreset(
		iphoneUA,
		390,
		844,
		3.0,
		"portrait-primary",
		0,
		"en-US",
		"America/Los_Angeles",
		&touch,
	)
	exampleutil.AddCheck(&results, "应用移动端预设", "成功", fmt.Sprint(support))

	touchNotes := make([]string, 0, 3)
	for _, scope := range []string{"context", "global", "user_context"} {
		ok, scopeErr := page.Emulation().SetTouchEnabled(true, 5, scope)
		if scopeErr != nil {
			touchNotes = append(touchNotes, fmt.Sprintf("%s:error=%v", scope, scopeErr))
			continue
		}
		_ = page.Refresh()
		touchPoints, _ := page.RunJSExpr("navigator.maxTouchPoints || 0")
		touchNotes = append(touchNotes, fmt.Sprintf("%s:supported=%v,maxTouchPoints=%v", scope, ok, touchPoints))
	}
	exampleutil.AddCheck(&results, "触摸作用域结果", "成功", strings.Join(touchNotes, " | "))

	if err := page.Get("https://www.google.com/ncr"); err != nil {
		exampleutil.AddCheck(&results, "打开 Google 首页", "失败", err.Error())
		exampleutil.PrintChecks(results)
		return nil
	}
	page.Wait().Sleep(2 * time.Second)

	ua, _ := page.RunJSExpr("navigator.userAgent")
	width, _ := page.RunJSExpr("window.innerWidth")
	height, _ := page.RunJSExpr("window.innerHeight")
	dpr, _ := page.RunJSExpr("window.devicePixelRatio")
	language, _ := page.RunJSExpr("navigator.language")
	touchPoints, _ := page.RunJSExpr("navigator.maxTouchPoints || 0")
	title, _ := page.Title()

	uaOK := strings.Contains(fmt.Sprint(ua), "iPhone")
	viewportOK := intValue(width) > 0 && intValue(width) <= 500
	titleOK := strings.Contains(strings.ToLower(title), "google")

	exampleutil.AddCheck(&results, "UA 命中移动端", statusOf(uaOK), fmt.Sprintf("UA=%v", ua))
	exampleutil.AddCheck(&results, "视口为移动端宽度", statusOf(viewportOK), fmt.Sprintf("viewport=%vx%v dpr=%v", width, height, dpr))
	exampleutil.AddCheck(&results, "Google 页面打开成功", statusOf(titleOK), fmt.Sprintf("title=%s", title))
	exampleutil.AddCheck(&results, "语言与触摸信息", "成功", fmt.Sprintf("language=%v maxTouchPoints=%v", language, touchPoints))

	exampleutil.PrintChecks(results)
	return nil
}

func statusOf(ok bool) string {
	if ok {
		return "成功"
	}
	return "失败"
}

func intValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	default:
		return 0
	}
}
