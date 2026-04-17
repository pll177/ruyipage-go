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
	fmt.Println("测试 21: Emulation 模块")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 13)
	if err := page.Get("data:text/html,<html><body><h1>Emulation Test</h1></body></html>"); err != nil {
		return err
	}

	customUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15"
	if err := page.Emulation().SetUserAgent(customUA, "iPhone"); err != nil {
		exampleutil.AddCheck(&results, "User-Agent 覆盖", "失败", err.Error())
	} else {
		_ = page.Refresh()
		newUA, readErr := page.RunJSExpr("navigator.userAgent")
		if readErr != nil {
			exampleutil.AddCheck(&results, "User-Agent 覆盖", "失败", readErr.Error())
		} else if strings.Contains(fmt.Sprint(newUA), customUA) {
			exampleutil.AddCheck(&results, "User-Agent 覆盖", "成功", "UA 命中自定义值")
		} else {
			exampleutil.AddCheck(&results, "User-Agent 覆盖", "失败", fmt.Sprintf("当前=%v", newUA))
		}
	}

	if err := page.Emulation().SetGeolocation(39.9042, 116.4074, 100); err != nil {
		exampleutil.AddCheck(&results, "地理位置覆盖", "失败", err.Error())
	} else {
		exampleutil.AddCheck(&results, "地理位置覆盖", "成功", "命令执行成功")
	}

	if err := page.Emulation().SetTimezone("America/New_York"); err != nil {
		exampleutil.AddCheck(&results, "时区覆盖", "失败", err.Error())
	} else {
		_ = page.Refresh()
		timezone, readErr := page.RunJSExpr("Intl.DateTimeFormat().resolvedOptions().timeZone")
		if readErr != nil {
			exampleutil.AddCheck(&results, "时区覆盖", "失败", readErr.Error())
		} else if strings.Contains(fmt.Sprint(timezone), "New_York") {
			exampleutil.AddCheck(&results, "时区覆盖", "成功", fmt.Sprintf("当前=%v", timezone))
		} else {
			exampleutil.AddCheck(&results, "时区覆盖", "失败", fmt.Sprintf("当前=%v", timezone))
		}
	}

	if err := page.Emulation().SetLocale([]string{"ja-JP", "ja"}); err != nil {
		exampleutil.AddCheck(&results, "语言覆盖", "失败", err.Error())
	} else {
		_ = page.Refresh()
		language, readErr := page.RunJSExpr("navigator.language")
		if readErr != nil {
			exampleutil.AddCheck(&results, "语言覆盖", "失败", readErr.Error())
		} else if strings.Contains(strings.ToLower(fmt.Sprint(language)), "ja") {
			exampleutil.AddCheck(&results, "语言覆盖", "成功", fmt.Sprintf("当前=%v", language))
		} else {
			exampleutil.AddCheck(&results, "语言覆盖", "失败", fmt.Sprintf("当前=%v", language))
		}
	}

	if err := page.Emulation().SetScreenOrientation("landscape-primary", 90); err != nil {
		exampleutil.AddCheck(&results, "屏幕方向覆盖", "失败", err.Error())
	} else {
		_ = page.Refresh()
		orientation, readErr := page.RunJSExpr("screen.orientation.type")
		if readErr != nil {
			exampleutil.AddCheck(&results, "屏幕方向覆盖", "失败", readErr.Error())
		} else if strings.Contains(fmt.Sprint(orientation), "landscape") {
			exampleutil.AddCheck(&results, "屏幕方向覆盖", "成功", fmt.Sprintf("当前=%v", orientation))
		} else {
			exampleutil.AddCheck(&results, "屏幕方向覆盖", "失败", fmt.Sprintf("当前=%v", orientation))
		}
	}

	dpr := 2.0
	if err := page.Emulation().SetScreenSize(1920, 1080, &dpr); err != nil {
		exampleutil.AddCheck(&results, "屏幕设置覆盖", "失败", err.Error())
	} else {
		_ = page.Refresh()
		sw, swErr := page.RunJSExpr("screen.width")
		sh, shErr := page.RunJSExpr("screen.height")
		if swErr != nil || shErr != nil {
			exampleutil.AddCheck(&results, "屏幕设置覆盖", "失败", fmt.Sprintf("width err=%v, height err=%v", swErr, shErr))
		} else if intValue(sw) == 1920 && intValue(sh) == 1080 {
			exampleutil.AddCheck(&results, "屏幕设置覆盖", "成功", fmt.Sprintf("当前=%vx%v", sw, sh))
		} else {
			exampleutil.AddCheck(&results, "屏幕设置覆盖", "失败", fmt.Sprintf("当前=%vx%v", sw, sh))
		}
	}

	offlineSupported, offlineErr := page.Emulation().SetNetworkOffline(true)
	if offlineErr != nil {
		exampleutil.AddCheck(&results, "网络条件模拟", "失败", offlineErr.Error())
	} else if offlineSupported {
		exampleutil.AddCheck(&results, "网络条件模拟", "成功", "离线模式")
	} else {
		exampleutil.AddCheck(&results, "网络条件模拟", "不支持", "当前 Firefox 未实现")
	}
	_, _ = page.Emulation().SetNetworkOffline(false)

	touchSupported, touchErr := page.Emulation().SetTouchEnabled(true, 5, "context")
	if touchErr != nil {
		exampleutil.AddCheck(&results, "触摸模拟", "失败", touchErr.Error())
	} else if touchSupported {
		exampleutil.AddCheck(&results, "触摸模拟", "成功", "启用触摸")
	} else {
		exampleutil.AddCheck(&results, "触摸模拟", "不支持", "当前 Firefox 未实现")
	}

	jsSupported, jsErr := page.Emulation().SetJavaScriptEnabled(true)
	if jsErr != nil {
		exampleutil.AddCheck(&results, "JavaScript 开关", "失败", jsErr.Error())
	} else if jsSupported {
		exampleutil.AddCheck(&results, "JavaScript 开关", "成功", "启用 JS")
	} else {
		exampleutil.AddCheck(&results, "JavaScript 开关", "不支持", "当前 Firefox 未实现")
	}

	scrollbarSupported, scrollbarErr := page.Emulation().SetScrollbarType("overlay")
	if scrollbarErr != nil {
		exampleutil.AddCheck(&results, "滚动条类型", "失败", scrollbarErr.Error())
	} else if scrollbarSupported {
		exampleutil.AddCheck(&results, "滚动条类型", "成功", "overlay")
	} else {
		exampleutil.AddCheck(&results, "滚动条类型", "不支持", "当前 Firefox 未实现")
	}

	forcedColorSupported, forcedColorErr := page.Emulation().SetForcedColorsMode("dark")
	if forcedColorErr != nil {
		exampleutil.AddCheck(&results, "强制颜色模式", "失败", forcedColorErr.Error())
	} else if forcedColorSupported {
		exampleutil.AddCheck(&results, "强制颜色模式", "成功", "dark")
	} else {
		exampleutil.AddCheck(&results, "强制颜色模式", "不支持", "当前 Firefox 未实现")
	}

	bypassSupported, bypassErr := page.Emulation().SetBypassCSP(true)
	if bypassErr != nil {
		exampleutil.AddCheck(&results, "CSP 绕过", "失败", bypassErr.Error())
	} else if bypassSupported {
		exampleutil.AddCheck(&results, "CSP 绕过", "成功", "enabled=true")
	} else {
		exampleutil.AddCheck(&results, "CSP 绕过", "不支持", "当前 Firefox 未实现")
	}

	enableTouch := true
	support := page.Emulation().ApplyMobilePreset(
		customUA,
		390,
		844,
		3.0,
		"portrait-primary",
		0,
		"en-US",
		"America/New_York",
		&enableTouch,
	)
	exampleutil.AddCheck(&results, "移动端预设", "成功", fmt.Sprint(support))

	exampleutil.PrintChecks(results)
	page.Wait().Sleep(300 * time.Millisecond)
	return nil
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
