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
	fmt.Println("测试 27: Emulation 高级能力")
	fmt.Println(strings.Repeat("=", 70))
	opt := ruyipage.NewFirefoxOptions().
		WithBrowserPath("C:\\Users\\pll177\\Desktop\\core\\firefox.exe").
		Headless(false)

	page, err := ruyipage.NewFirefoxPage(opt)
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 9)
	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "example.com 已加载")

	forcedActive, activeErr := page.Emulation().SetForcedColorsMode("active")
	addSupportedCheck(&results, "emulation.setForcedColorsModeThemeOverride active", forcedActive, activeErr)

	forcedNone, noneErr := page.Emulation().SetForcedColorsMode("none")
	addSupportedCheck(&results, "emulation.setForcedColorsModeThemeOverride none", forcedNone, noneErr)

	sizeResult := validateScreenSize(page, 1920, 1080, 2.0)
	exampleutil.AddCheck(&results, "emulation.setScreenSettingsOverride 1920x1080@2", sizeResult.status, sizeResult.note)

	sizeResult = validateScreenSize(page, 375, 812, 3.0)
	exampleutil.AddCheck(&results, "emulation.setScreenSettingsOverride 375x812@3", sizeResult.status, sizeResult.note)

	orientationResult := validateOrientation(page, "portrait-primary", 0)
	exampleutil.AddCheck(&results, "emulation.setScreenOrientationOverride portrait", orientationResult.status, orientationResult.note)

	orientationResult = validateOrientation(page, "landscape-primary", 90)
	exampleutil.AddCheck(&results, "emulation.setScreenOrientationOverride landscape", orientationResult.status, orientationResult.note)

	jsOff, jsOffErr := page.Emulation().SetJavaScriptEnabled(false)
	addSupportedCheck(&results, "emulation.setScriptingEnabled false", jsOff, jsOffErr)
	jsOn, jsOnErr := page.Emulation().SetJavaScriptEnabled(true)
	addSupportedCheck(&results, "emulation.setScriptingEnabled true", jsOn, jsOnErr)

	scrollbarNone, scrollbarNoneErr := page.Emulation().SetScrollbarType("none")
	addSupportedCheck(&results, "emulation.setScrollbarTypeOverride none", scrollbarNone, scrollbarNoneErr)
	scrollbarStandard, scrollbarStandardErr := page.Emulation().SetScrollbarType("standard")
	addSupportedCheck(&results, "emulation.setScrollbarTypeOverride standard", scrollbarStandard, scrollbarStandardErr)
	scrollbarOverlay, scrollbarOverlayErr := page.Emulation().SetScrollbarType("overlay")
	addSupportedCheck(&results, "emulation.setScrollbarTypeOverride overlay", scrollbarOverlay, scrollbarOverlayErr)

	exampleutil.PrintChecks(results)
	page.Wait().Sleep(300 * time.Millisecond)
	return nil
}

type checkResult struct {
	status string
	note   string
}

func validateScreenSize(page *ruyipage.FirefoxPage, width int, height int, dprValue float64) checkResult {
	if page == nil {
		return checkResult{status: "失败", note: "page 为空"}
	}
	if err := page.Emulation().SetScreenSize(width, height, &dprValue); err != nil {
		return checkResult{status: "失败", note: err.Error()}
	}
	_ = page.Refresh()
	page.Wait().Sleep(200 * time.Millisecond)

	value, err := page.RunJS(`return [screen.width, screen.height, window.devicePixelRatio]`)
	if err != nil {
		return checkResult{status: "失败", note: err.Error()}
	}
	items := anySlice(value)
	if len(items) < 3 {
		return checkResult{status: "失败", note: fmt.Sprintf("unexpected=%v", value)}
	}
	if intValue(items[0]) == width && intValue(items[1]) == height && floatValue(items[2]) == dprValue {
		return checkResult{status: "成功", note: fmt.Sprint(items)}
	}
	if intValue(items[0]) == width && intValue(items[1]) == height {
		return checkResult{status: "跳过", note: fmt.Sprintf("尺寸生效，但 DPR 未生效: %v", items)}
	}
	return checkResult{status: "失败", note: fmt.Sprint(items)}
}

func validateOrientation(page *ruyipage.FirefoxPage, orientationType string, angle int) checkResult {
	if page == nil {
		return checkResult{status: "失败", note: "page 为空"}
	}
	if err := page.Emulation().SetScreenOrientation(orientationType, angle); err != nil {
		return checkResult{status: "失败", note: err.Error()}
	}
	_ = page.Refresh()
	page.Wait().Sleep(200 * time.Millisecond)

	value, err := page.RunJS(`return [screen.orientation.type, screen.orientation.angle]`)
	if err != nil {
		return checkResult{status: "失败", note: err.Error()}
	}
	items := anySlice(value)
	if len(items) < 2 {
		return checkResult{status: "失败", note: fmt.Sprintf("unexpected=%v", value)}
	}
	if fmt.Sprint(items[0]) == orientationType && intValue(items[1]) == angle {
		return checkResult{status: "成功", note: fmt.Sprint(items)}
	}
	if fmt.Sprint(items[0]) == orientationType {
		return checkResult{status: "跳过", note: fmt.Sprintf("方向类型生效，但角度未生效: %v", items)}
	}
	return checkResult{status: "失败", note: fmt.Sprint(items)}
}

func addSupportedCheck(results *[]exampleutil.CheckRow, item string, supported bool, err error) {
	if err != nil {
		exampleutil.AddCheck(results, item, "失败", err.Error())
		return
	}
	if supported {
		exampleutil.AddCheck(results, item, "成功", "标准命令已实现")
		return
	}
	exampleutil.AddCheck(results, item, "不支持", "当前 Firefox 未实现该命令")
}

func anySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []string:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, item)
		}
		return result
	default:
		return nil
	}
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

func floatValue(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}
