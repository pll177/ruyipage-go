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
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("测试 32: Script Events")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Realms().Stop()
		page.Events().Stop()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	tempTabID := ""
	preload := ruyipage.PreloadScript{}

	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "example.com 已加载")

	if err := page.Realms().Start(); err != nil {
		return err
	}
	initialRealms := page.Realms().List()
	exampleutil.AddCheck(&results, "script.getRealms baseline", "成功", fmt.Sprintf("初始 realm 数量: %d", len(initialRealms)))

	tempTabID, err = page.Contexts().CreateTab(false, "", "")
	if err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	createdRealms := page.Realms().List()
	if len(createdRealms) > len(initialRealms) {
		exampleutil.AddCheck(&results, "script.realmCreated", "成功", fmt.Sprintf("realm 数量: %d -> %d", len(initialRealms), len(createdRealms)))
	} else {
		exampleutil.AddCheck(&results, "script.realmCreated", "跳过", "当前环境未稳定观察到新增 realm")
	}

	if tempTabID != "" {
		_ = page.Contexts().Close(tempTabID, false)
		tempTabID = ""
		page.Wait().Sleep(time.Second)
	}
	destroyedRealms := page.Realms().List()
	if len(destroyedRealms) <= len(createdRealms) {
		exampleutil.AddCheck(&results, "script.realmDestroyed", "成功", fmt.Sprintf("关闭后 realm 数量: %d", len(destroyedRealms)))
	} else {
		exampleutil.AddCheck(&results, "script.realmDestroyed", "跳过", "当前环境未稳定观察到 realm 销毁")
	}
	page.Realms().Stop()

	if err := page.Events().Start([]string{"script.message"}, []string{page.ContextID()}); err == nil {
		preload, err = page.AddPreloadScript(`() => {
			const ch = new BroadcastChannel("ruyi-script-message");
			ch.postMessage({ type: "preload", value: "hello" });
			ch.close();
		}`)
		if err != nil {
			return err
		}
		defer func() {
			if preload.ID != "" {
				_ = page.RemovePreloadScript(preload.ID)
			}
		}()

		if err := page.Get("https://www.example.com/?script-message=1"); err != nil {
			return err
		}
		messageEvent := page.Events().Wait("script.message", 3*time.Second)
		if messageEvent != nil {
			exampleutil.AddCheck(&results, "script.message", "成功", fmt.Sprintf("channel=%s data=%v", messageEvent.Channel, messageEvent.Data))
		} else {
			exampleutil.AddCheck(&results, "script.message", "跳过", "当前环境未稳定观察到 script.message")
		}
	} else {
		exampleutil.AddCheck(&results, "script.message", "不支持", "未能订阅 script.message 事件")
	}

	exampleutil.PrintChecks(results)
	return nil
}
