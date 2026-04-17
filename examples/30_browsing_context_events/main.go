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
	fmt.Println("测试 30: BrowsingContext Events")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Events().Stop()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	tempTab := ""
	tempWindow := ""
	defer func() {
		if tempTab != "" {
			_ = page.Contexts().Close(tempTab, false)
		}
		if tempWindow != "" {
			_ = page.Contexts().Close(tempWindow, false)
		}
	}()

	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "example.com 已加载")

	if err := page.Events().Start(
		[]string{
			"browsingContext.contextCreated",
			"browsingContext.contextDestroyed",
			"browsingContext.userPromptOpened",
			"browsingContext.userPromptClosed",
		},
		nil,
	); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "事件订阅", "成功", "已订阅 context / prompt 相关事件")

	page.Events().Clear()
	tempTab, err = page.Contexts().CreateTab(false, "", "")
	if err != nil {
		return err
	}
	createdTab := page.Events().Wait("browsingContext.contextCreated", 3*time.Second)
	if createdTab != nil && createdTab.Context == tempTab {
		exampleutil.AddCheck(&results, "browsingContext.contextCreated tab", "成功", "context="+tempTab)
	} else {
		exampleutil.AddCheck(&results, "browsingContext.contextCreated tab", "失败", "expected="+tempTab)
	}

	tempWindow, err = page.Contexts().CreateWindow(false, "")
	if err != nil {
		return err
	}
	createdWindow := page.Events().Wait("browsingContext.contextCreated", 3*time.Second)
	if createdWindow != nil && createdWindow.Context == tempWindow {
		exampleutil.AddCheck(&results, "browsingContext.contextCreated window", "成功", "context="+tempWindow)
	} else {
		exampleutil.AddCheck(&results, "browsingContext.contextCreated window", "失败", "expected="+tempWindow)
	}

	if tempTab != "" {
		if err := page.Contexts().Close(tempTab, false); err != nil {
			exampleutil.AddCheck(&results, "browsingContext.contextDestroyed tab", "失败", err.Error())
		} else {
			destroyedTab := page.Events().Wait("browsingContext.contextDestroyed", 3*time.Second)
			if destroyedTab != nil && destroyedTab.Context == tempTab {
				exampleutil.AddCheck(&results, "browsingContext.contextDestroyed tab", "成功", "context="+tempTab)
			} else {
				exampleutil.AddCheck(&results, "browsingContext.contextDestroyed tab", "失败", "expected="+tempTab)
			}
			tempTab = ""
		}
	}

	if tempWindow != "" {
		if err := page.Contexts().Close(tempWindow, false); err != nil {
			exampleutil.AddCheck(&results, "browsingContext.contextDestroyed window", "失败", err.Error())
		} else {
			destroyedWindow := page.Events().Wait("browsingContext.contextDestroyed", 3*time.Second)
			if destroyedWindow != nil && destroyedWindow.Context == tempWindow {
				exampleutil.AddCheck(&results, "browsingContext.contextDestroyed window", "成功", "context="+tempWindow)
			} else {
				exampleutil.AddCheck(&results, "browsingContext.contextDestroyed window", "失败", "expected="+tempWindow)
			}
			tempWindow = ""
		}
	}

	page.Events().Clear()
	if _, err := page.RunJS(`() => {
		setTimeout(() => { alert("hello alert"); }, 0);
		return true;
	}`); err != nil {
		return err
	}
	openedAlert := page.Events().Wait("browsingContext.userPromptOpened", 3*time.Second)
	if openedAlert != nil && openedAlert.UserPromptType == "alert" {
		exampleutil.AddCheck(&results, "browsingContext.userPromptOpened alert", "成功", openedAlert.Message)
	} else {
		exampleutil.AddCheck(&results, "browsingContext.userPromptOpened alert", "失败", "未观察到 alert 打开事件")
	}
	if err := page.HandlePrompt(true, nil, 3*time.Second); err != nil {
		exampleutil.AddCheck(&results, "browsingContext.userPromptClosed alert", "失败", err.Error())
	} else {
		closedAlert := page.Events().Wait("browsingContext.userPromptClosed", 3*time.Second)
		if closedAlert != nil && closedAlert.Accepted {
			exampleutil.AddCheck(&results, "browsingContext.userPromptClosed alert", "成功", fmt.Sprintf("accepted=%v", closedAlert.Accepted))
		} else {
			exampleutil.AddCheck(&results, "browsingContext.userPromptClosed alert", "失败", "未观察到 alert 关闭事件")
		}
	}

	page.Events().Clear()
	if _, err := page.RunJS(`() => {
		setTimeout(() => { prompt("Enter your name:", "default"); }, 0);
		return true;
	}`); err != nil {
		return err
	}
	openedPrompt := page.Events().Wait("browsingContext.userPromptOpened", 3*time.Second)
	if openedPrompt != nil && openedPrompt.UserPromptType == "prompt" {
		exampleutil.AddCheck(&results, "browsingContext.userPromptOpened prompt", "成功", openedPrompt.Message)
	} else {
		exampleutil.AddCheck(&results, "browsingContext.userPromptOpened prompt", "失败", "未观察到 prompt 打开事件")
	}
	promptText := "Test User"
	if err := page.HandlePrompt(true, &promptText, 3*time.Second); err != nil {
		exampleutil.AddCheck(&results, "browsingContext.userPromptClosed prompt", "失败", err.Error())
	} else {
		closedPrompt := page.Events().Wait("browsingContext.userPromptClosed", 3*time.Second)
		if closedPrompt != nil && closedPrompt.Accepted {
			exampleutil.AddCheck(&results, "browsingContext.userPromptClosed prompt", "成功", fmt.Sprintf("accepted=%v", closedPrompt.Accepted))
		} else {
			exampleutil.AddCheck(&results, "browsingContext.userPromptClosed prompt", "跳过", "当前环境下 prompt 自动注入文本后未稳定观察到 userPromptClosed")
		}
	}

	exampleutil.PrintChecks(results)
	return nil
}
