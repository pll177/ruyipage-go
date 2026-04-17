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
	fmt.Println("测试 26: BrowsingContext 高级功能")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 7)
	childID := ""
	defer func() {
		if childID != "" {
			_ = page.Contexts().Close(childID, false)
		}
	}()

	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "example.com 已加载")

	tree, err := page.Contexts().GetTree(nil, "")
	if err != nil {
		return err
	}
	if len(tree.Contexts) > 0 {
		exampleutil.AddCheck(&results, "browsingContext.getTree", "成功", fmt.Sprintf("上下文数量: %d", len(tree.Contexts)))
	} else {
		exampleutil.AddCheck(&results, "browsingContext.getTree", "失败", "未返回任何上下文")
	}

	childID, err = page.Contexts().CreateTab(false, "", "")
	if err != nil {
		return err
	}
	if childID != "" {
		exampleutil.AddCheck(&results, "browsingContext.create", "成功", "新 context: "+childID)
		if err := page.Contexts().Close(childID, false); err != nil {
			exampleutil.AddCheck(&results, "browsingContext.close", "失败", err.Error())
		} else {
			exampleutil.AddCheck(&results, "browsingContext.close", "成功", "新 context 已关闭")
			childID = ""
		}
	} else {
		exampleutil.AddCheck(&results, "browsingContext.create", "失败", "未返回 context ID")
	}

	reloadResult, err := page.Contexts().Reload(false, "", "")
	if err != nil {
		return err
	}
	if fmt.Sprint(reloadResult["navigation"]) != "" {
		exampleutil.AddCheck(&results, "browsingContext.reload", "成功", fmt.Sprintf("navigation=%v", reloadResult["navigation"]))
	} else {
		exampleutil.AddCheck(&results, "browsingContext.reload", "跳过", "当前返回未包含 navigation")
	}

	ignoreCacheResult, ignoreCacheErr := page.Contexts().Reload(true, "", "")
	if ignoreCacheErr != nil {
		exampleutil.AddCheck(&results, "browsingContext.reload ignoreCache", "不支持", ignoreCacheErr.Error())
	} else {
		exampleutil.AddCheck(&results, "browsingContext.reload ignoreCache", "成功", fmt.Sprintf("navigation=%v", ignoreCacheResult["navigation"]))
	}

	if err := page.Contexts().SetBypassCSP(true, ""); err != nil {
		exampleutil.AddCheck(&results, "browsingContext.setBypassCSP", "不支持", err.Error())
	} else {
		exampleutil.AddCheck(&results, "browsingContext.setBypassCSP", "成功", "标准命令调用成功")
	}

	if err := page.Contexts().SetViewport(800, 600, nil, ""); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)
	viewport := page.Rect().ViewportSize()
	if viewport["width"] == 800 && viewport["height"] == 600 {
		exampleutil.AddCheck(&results, "browsingContext.setViewport 800x600", "成功", fmt.Sprintf("实际视口: %v", viewport))
	} else {
		exampleutil.AddCheck(&results, "browsingContext.setViewport 800x600", "失败", fmt.Sprintf("实际视口: %v", viewport))
	}

	if err := page.Contexts().SetViewport(375, 667, nil, ""); err != nil {
		return err
	}
	page.Wait().Sleep(200 * time.Millisecond)
	viewport = page.Rect().ViewportSize()
	if viewport["width"] == 375 && viewport["height"] == 667 {
		exampleutil.AddCheck(&results, "browsingContext.setViewport 375x667", "成功", fmt.Sprintf("实际视口: %v", viewport))
	} else {
		exampleutil.AddCheck(&results, "browsingContext.setViewport 375x667", "失败", fmt.Sprintf("实际视口: %v", viewport))
	}

	exampleutil.PrintChecks(results)
	return nil
}
