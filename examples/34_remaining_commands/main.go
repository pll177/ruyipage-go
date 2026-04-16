package main

import (
	"encoding/base64"
	"fmt"
	"os"
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
	fmt.Println("测试 34: Remaining Commands")
	fmt.Println(strings.Repeat("=", 70))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(9430))
	if err := server.Start(); err != nil {
		return err
	}
	defer func() {
		_ = server.Stop()
	}()

	outputDir, err := exampleutil.OutputDir("34_remaining_commands")
	if err != nil {
		return err
	}
	screenshotPath := outputDir + string(os.PathSeparator) + "test_screenshot.png"

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Intercept().Stop()
		page.Events().Stop()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 12)
	baseURL := strings.TrimRight(server.GetURL(""), "/")

	if err := page.Get(baseURL + "/nav/basic"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "服务地址: "+baseURL)

	dataBytes, err := page.Screenshot("", false)
	if err != nil {
		return err
	}
	if len(dataBytes) > 100 {
		exampleutil.AddCheck(&results, "browsingContext.captureScreenshot viewport", "成功", fmt.Sprintf("截图字节数: %d", len(dataBytes)))
	} else {
		exampleutil.AddCheck(&results, "browsingContext.captureScreenshot viewport", "失败", "截图数据过短")
	}

	if _, err := page.Screenshot(screenshotPath, false); err != nil {
		return err
	}
	if info, err := os.Stat(screenshotPath); err == nil && info.Size() > 0 {
		exampleutil.AddCheck(&results, "screenshot save file", "成功", screenshotPath)
	} else {
		exampleutil.AddCheck(&results, "screenshot save file", "失败", "未保存截图文件")
	}

	if err := page.Get(baseURL + "/nav/basic?a=1"); err != nil {
		return err
	}
	if err := page.Get(baseURL + "/nav/basic?a=2"); err != nil {
		return err
	}
	if err := page.Get(baseURL + "/nav/basic?a=3"); err != nil {
		return err
	}
	_ = page.Back()
	backURL, _ := page.URL()
	_ = page.Back()
	back2URL, _ := page.URL()
	_ = page.Forward()
	forwardURL, _ := page.URL()
	if strings.HasSuffix(backURL, "?a=2") && strings.HasSuffix(back2URL, "?a=1") && strings.HasSuffix(forwardURL, "?a=2") {
		exampleutil.AddCheck(&results, "browsingContext.traverseHistory", "成功", fmt.Sprintf("back=%s, back2=%s, forward=%s", backURL, back2URL, forwardURL))
	} else {
		exampleutil.AddCheck(&results, "browsingContext.traverseHistory", "失败", fmt.Sprintf("back=%s, back2=%s, forward=%s", backURL, back2URL, forwardURL))
	}

	if _, err := page.Intercept().StartResponses(nil, []map[string]any{{"type": "string", "pattern": baseURL + "/api/data"}}); err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, baseURL+"/api/data"); err != nil {
		return err
	}
	respReq := page.Intercept().Wait(5 * time.Second)
	if respReq != nil {
		if err := respReq.ContinueResponse(nil, "", nil); err != nil {
			return err
		}
		exampleutil.AddCheck(&results, "network.continueResponse", "成功", respReq.URL)
	} else {
		exampleutil.AddCheck(&results, "network.continueResponse", "跳过", "未稳定捕获到 responseStarted 拦截")
	}
	page.Intercept().Stop()

	if _, err := page.Intercept().StartRequests(nil, []map[string]any{{"type": "string", "pattern": baseURL + "/api/mock-source"}}); err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, baseURL+"/api/mock-source"); err != nil {
		return err
	}
	mockReq := page.Intercept().Wait(5 * time.Second)
	if mockReq != nil {
		if err := mockReq.Mock(`{"status":"mocked"}`, 200, map[string]string{"content-type": "application/json", "access-control-allow-origin": "*"}, ""); err != nil {
			return err
		}
		exampleutil.AddCheck(&results, "network.provideResponse", "成功", mockReq.URL)
	} else {
		exampleutil.AddCheck(&results, "network.provideResponse", "跳过", "未稳定捕获到 beforeRequestSent 拦截")
	}
	page.Intercept().Stop()

	if _, err := page.Intercept().StartRequests(nil, []map[string]any{{"type": "string", "pattern": baseURL + "/api/slow"}}); err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, baseURL+"/api/slow"); err != nil {
		return err
	}
	failReq := page.Intercept().Wait(5 * time.Second)
	if failReq != nil {
		if err := failReq.Fail(); err != nil {
			return err
		}
		exampleutil.AddCheck(&results, "network.failRequest", "成功", failReq.URL)
	} else {
		exampleutil.AddCheck(&results, "network.failRequest", "跳过", "未稳定捕获到 beforeRequestSent 拦截")
	}
	page.Intercept().Stop()

	if err := page.Events().Start([]string{"network.authRequired"}, []string{page.ContextID()}); err != nil {
		return err
	}
	if _, err := page.Intercept().Start(nil, []map[string]any{{"type": "string", "pattern": baseURL + "/api/auth"}}, []string{"authRequired"}); err != nil {
		return err
	}
	_ = page.Get(baseURL + "/api/auth")
	authReq := page.Intercept().Wait(5 * time.Second)
	authEvent := page.Events().Wait("network.authRequired", 5*time.Second)
	if authReq != nil && authEvent != nil {
		if err := authReq.ContinueWithAuth("provideCredentials", "user", "pass"); err != nil {
			return err
		}
		exampleutil.AddCheck(&results, "network.continueWithAuth", "成功", authReq.URL)
	} else {
		exampleutil.AddCheck(&results, "network.continueWithAuth", "跳过", "当前环境未稳定观察到 authRequired")
	}
	page.Intercept().Stop()
	page.Events().Stop()

	actionPage := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Actions Test</title></head>
<body>
	<input id="input-box" autofocus>
	<button id="btn" onclick="window.btnClicked = true">Click Me</button>
	<script>
		window.btnClicked = false;
		window.lastKey = "";
		document.getElementById("input-box").addEventListener("keydown", function(event) {
			window.lastKey = event.key;
		});
	</script>
</body>
</html>`
	actionURL := "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString([]byte(actionPage))
	if err := page.Get(actionURL); err != nil {
		return err
	}
	inputBox, err := page.Ele("#input-box", 1, 5*time.Second)
	if err != nil {
		return err
	}
	button, err := page.Ele("#btn", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if inputBox == nil || button == nil {
		return fmt.Errorf("未找到动作测试页元素")
	}
	if err := page.Actions().MoveTo(inputBox, 0, 0, 0, nil).Click(nil, 1).Press("a").Perform(); err != nil {
		return err
	}
	lastKey, err := page.RunJS(`return window.lastKey`)
	if err != nil {
		return err
	}
	if err := page.Actions().MoveTo(button, 0, 0, 0, nil).Click(nil, 1).Perform(); err != nil {
		return err
	}
	btnClicked, err := page.RunJS(`return window.btnClicked`)
	if err != nil {
		return err
	}
	if err := page.Actions().ReleaseAll(); err != nil {
		return err
	}
	if fmt.Sprint(lastKey) == "a" {
		exampleutil.AddCheck(&results, "input.performActions keyboard", "成功", "lastKey=a")
	} else {
		exampleutil.AddCheck(&results, "input.performActions keyboard", "失败", fmt.Sprintf("lastKey=%v", lastKey))
	}
	if clicked, _ := btnClicked.(bool); clicked {
		exampleutil.AddCheck(&results, "input.performActions pointer", "成功", "按钮点击事件已触发")
	} else {
		exampleutil.AddCheck(&results, "input.performActions pointer", "失败", fmt.Sprintf("btnClicked=%v", btnClicked))
	}
	exampleutil.AddCheck(&results, "input.releaseActions", "成功", "动作状态已释放")

	exampleutil.PrintChecks(results)
	return nil
}
