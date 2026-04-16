package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
	"ruyipage-go/examples/internal/testserver"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("测试 31: Network Events")
	fmt.Println(strings.Repeat("=", 70))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(9330))
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
		page.Intercept().Stop()
		page.Events().Stop()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	baseURL := strings.TrimRight(server.GetURL(""), "/")

	if err := page.Get("about:blank"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "about:blank 已加载，服务地址: "+baseURL)

	if err := page.Events().Start(
		[]string{
			"network.beforeRequestSent",
			"network.responseStarted",
			"network.responseCompleted",
			"network.fetchError",
			"network.authRequired",
		},
		[]string{page.ContextID()},
	); err != nil {
		return err
	}

	page.Events().Clear()
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, baseURL+"/api/data"); err != nil {
		return err
	}

	beforeEvent := page.Events().Wait("network.beforeRequestSent", 5*time.Second)
	startedEvent := page.Events().Wait("network.responseStarted", 5*time.Second)
	completedEvent := page.Events().Wait("network.responseCompleted", 5*time.Second)

	if beforeEvent != nil && strings.HasSuffix(stringValue(beforeEvent.Request["url"]), "/api/data") {
		exampleutil.AddCheck(&results, "network.beforeRequestSent", "成功", stringValue(beforeEvent.Request["url"]))
	} else {
		exampleutil.AddCheck(&results, "network.beforeRequestSent", "失败", "未观察到 /api/data 请求开始事件")
	}

	if startedEvent != nil && intValue(startedEvent.Response["status"]) == 200 {
		exampleutil.AddCheck(&results, "network.responseStarted", "成功", fmt.Sprintf("status=%d", intValue(startedEvent.Response["status"])))
	} else {
		exampleutil.AddCheck(&results, "network.responseStarted", "失败", "未观察到 200 响应开始事件")
	}

	if completedEvent != nil && intValue(completedEvent.Response["status"]) == 200 {
		exampleutil.AddCheck(&results, "network.responseCompleted", "成功", fmt.Sprintf("status=%d", intValue(completedEvent.Response["status"])))
	} else {
		exampleutil.AddCheck(&results, "network.responseCompleted", "失败", "未观察到 200 响应完成事件")
	}

	page.Events().Clear()
	if _, err := page.RunJS(`fetch('http://127.0.0.1:9/').catch(function(){ return null; }); return true;`); err != nil {
		return err
	}
	errorEvent := page.Events().Wait("network.fetchError", 5*time.Second)
	if errorEvent != nil {
		exampleutil.AddCheck(&results, "network.fetchError", "成功", errorEvent.ErrorText)
	} else {
		exampleutil.AddCheck(&results, "network.fetchError", "跳过", "当前环境未稳定观察到 fetchError")
	}

	page.Events().Clear()
	if _, err := page.Intercept().Start(nil, []map[string]any{
		{"type": "string", "pattern": baseURL + "/api/auth"},
	}, []string{"authRequired"}); err != nil {
		return err
	}
	if err := page.Get(baseURL + "/api/auth"); err != nil {
		// 认证挑战前的 401 / 中断对本示例是预期路径。
	}
	authReq := page.Intercept().Wait(5 * time.Second)
	authEvent := page.Events().Wait("network.authRequired", 5*time.Second)
	if authReq != nil && authEvent != nil {
		exampleutil.AddCheck(&results, "network.authRequired", "成功", authReq.URL)
		if err := authReq.ContinueWithAuth("provideCredentials", "user", "pass"); err != nil {
			exampleutil.AddCheck(&results, "network.authRequired credentials", "失败", err.Error())
		} else {
			authDone := page.Events().Wait("network.responseCompleted", 5*time.Second)
			if authDone != nil && intValue(authDone.Response["status"]) == 200 {
				exampleutil.AddCheck(&results, "network.authRequired credentials", "成功", "提供凭证后认证通过")
			} else {
				exampleutil.AddCheck(&results, "network.authRequired credentials", "失败", "提供凭证后未观察到 200 完成事件")
			}
		}
	} else {
		exampleutil.AddCheck(&results, "network.authRequired", "跳过", "当前环境未稳定观察到 authRequired")
		exampleutil.AddCheck(&results, "network.authRequired credentials", "跳过", "未观察到 authRequired，跳过凭证验证")
	}

	exampleutil.PrintChecks(results)
	return nil
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
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
