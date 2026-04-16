package main

import (
	"fmt"
	"sort"
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
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试18: 高级网络功能")
	fmt.Println(strings.Repeat("=", 60))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(8889))
	if err := server.Start(); err != nil {
		return err
	}
	defer func() {
		_ = server.Stop()
	}()

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Intercept().Stop()
		page.Events().Stop()
		page.Wait().Sleep(time.Second)
		_ = page.Quit(0, false)
	}()

	if err := page.Get("about:blank"); err != nil {
		return err
	}
	page.Wait().Sleep(600 * time.Millisecond)

	fmt.Println("\n1. beforeRequestSent 拦截 + 继续请求:")
	intercepted := make([]string, 0, 1)
	_, err = page.Intercept().StartRequests(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/data") {
			intercepted = append(intercepted, req.URL)
			fmt.Printf("   拦截到请求: %s %s\n", req.Method, req.URL)
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil)
	if err != nil {
		return err
	}
	requestResult, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(response){ return response.json(); })
			.then(function(data){ return data.status; })
			.catch(function(error){ return String(error); });
	}`, server.GetURL("/api/data"))
	if err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   请求结果: %v\n", requestResult)
	fmt.Printf("   拦截计数: %d\n", len(intercepted))
	page.Intercept().Stop()

	fmt.Println("\n2. 修改请求头:")
	headerErrs := make(chan error, 1)
	_, err = page.Intercept().StartRequests(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/headers") {
			headers := mergeRequestHeaders(req.Headers, map[string]string{
				"X-Ruyi-Demo": "yes",
				"User-Agent":  "RuyiPage/Example18",
			})
			if err := req.ContinueRequest("", "", headers, nil); err != nil {
				select {
				case headerErrs <- err:
				default:
				}
				return
			}
			fmt.Printf("   ✓ 已注入请求头（method=%s）\n", req.Method)
			return
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil)
	if err != nil {
		return err
	}
	headersValue, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(response){ return response.json(); })
			.catch(function(error){ return {error: String(error)}; });
	}`, server.GetURL("/api/headers"))
	if err != nil {
		return err
	}
	select {
	case err := <-headerErrs:
		return fmt.Errorf("修改请求头失败: %w", err)
	default:
	}
	headers, ok := headersValue.(map[string]any)
	if !ok {
		return fmt.Errorf("请求头结果类型异常: %T", headersValue)
	}
	fmt.Printf("   X-Ruyi-Demo: %s\n", mapValue(headers, "X-Ruyi-Demo", "x-ruyi-demo"))
	page.Intercept().Stop()

	fmt.Println("\n3. mock 响应:")
	_, err = page.Intercept().StartRequests(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/data") {
			_ = req.Mock(
				`{"status":"mocked","data":{"message":"这是Mock数据"}}`,
				200,
				map[string]string{
					"content-type":                "application/json",
					"access-control-allow-origin": "*",
				},
				"",
			)
			fmt.Println("   ✓ 已返回Mock响应")
			return
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil)
	if err != nil {
		return err
	}
	mockMessage, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(response){ return response.json(); })
			.then(function(data){ return data.status + ":" + data.data.message; })
			.catch(function(error){ return String(error); });
	}`, server.GetURL("/api/data"))
	if err != nil {
		return err
	}
	fmt.Printf("   Mock结果: %v\n", mockMessage)
	page.Intercept().Stop()

	fmt.Println("\n4. 阻止请求:")
	blockHandler := func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/error") {
			_ = req.Fail()
			fmt.Println("   ✓ 请求已阻止")
			return
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}
	_, err = page.Intercept().StartRequests(blockHandler, nil)
	if err != nil {
		return err
	}
	blocked, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(){ return "unexpected-success"; })
			.catch(function(error){ return "blocked:" + error.name; });
	}`, server.GetURL("/api/error"))
	if err != nil {
		return err
	}
	fmt.Printf("   阻止结果: %v\n", blocked)
	page.Intercept().Stop()

	fmt.Println("\n5. responseStarted 阶段修改响应状态码:")
	_, err = page.Intercept().StartResponses(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/data") {
			statusCode := 299
			_ = req.ContinueResponse(nil, "RuyiModified", &statusCode)
			fmt.Println("   ✓ 响应状态码已改为 299")
			return
		}
		_ = req.ContinueResponse(nil, "", nil)
	}, nil)
	if err != nil {
		return err
	}
	modifiedStatus, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(response){ return response.status + ":" + response.statusText; })
			.catch(function(error){ return String(error); });
	}`, server.GetURL("/api/data"))
	if err != nil {
		return err
	}
	fmt.Printf("   响应状态: %v\n", modifiedStatus)
	page.Intercept().Stop()

	fmt.Println("\n6. 队列模式 wait() 手动处理:")
	_, err = page.Intercept().Start(nil, nil, []string{"beforeRequestSent"})
	if err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return "sent";
	}`, server.GetURL("/api/data")); err != nil {
		return err
	}
	queued := page.Intercept().Wait(5 * time.Second)
	if queued == nil {
		return fmt.Errorf("队列模式未捕获到请求")
	}
	fmt.Printf("   wait捕获: %s %s\n", queued.Method, queued.URL)
	if err := queued.ContinueRequest("", "", nil, nil); err != nil {
		return err
	}
	page.Intercept().Stop()

	fmt.Println("\n7. fetchError 事件监听:")
	if err := page.Events().Start([]string{"network.fetchError"}, []string{page.ContextID()}); err != nil {
		return err
	}
	page.Events().Clear()
	_, err = page.Intercept().StartRequests(blockHandler, nil)
	if err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return "sent";
	}`, server.GetURL("/api/error")); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	page.Intercept().Stop()
	fetchErrorCount := 0
	for {
		event := page.Events().Wait("network.fetchError", 200*time.Millisecond)
		if event == nil {
			break
		}
		fetchErrorCount++
	}
	fmt.Printf("   fetchError 事件数量: %d\n", fetchErrorCount)
	page.Events().Stop()

	fmt.Println("\n8. 直接读取请求体 req.body:")
	capturedBodies := make([]string, 0, 1)
	_, err = page.Intercept().StartRequests(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/echo") && req.Method == "POST" {
			body := req.Body()
			capturedBodies = append(capturedBodies, body)
			fmt.Printf("   捕获 body: %s\n", body)
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil)
	if err != nil {
		return err
	}
	echoed, err := page.RunJS(`function(url){
		return fetch(url, {
			method: "POST",
			headers: {"Content-Type": "application/json"},
			body: JSON.stringify({message: "hello-body"})
		})
			.then(function(response){ return response.json(); })
			.then(function(data){ return data.body; })
			.catch(function(error){ return String(error); });
	}`, server.GetURL("/api/echo"))
	if err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	fmt.Printf("   服务端收到: %v\n", echoed)
	if len(capturedBodies) > 0 {
		fmt.Printf("   拦截侧读取: %s\n", capturedBodies[len(capturedBodies)-1])
	}
	page.Intercept().Stop()

	fmt.Println("\n9. GET 请求高频字段读取:")
	capturedGet := make([]*ruyipage.InterceptedRequest, 0, 1)
	_, err = page.Intercept().StartRequests(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/headers") {
			capturedGet = append(capturedGet, req)
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil)
	if err != nil {
		return err
	}
	getResponse, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(response){ return response.json(); })
			.catch(function(error){ return {error: String(error)}; });
	}`, server.GetURL("/api/headers"))
	if err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	page.Intercept().Stop()
	if len(capturedGet) > 0 {
		req := capturedGet[0]
		fmt.Printf("   GET字段: method=%s, request_id=%s, body=%s\n", req.Method, req.RequestID, req.Body())
		fmt.Printf("   GET headers Accept: %s\n", req.Headers["Accept"])
	} else {
		fmt.Println("   ⚠ 未捕获到 GET 请求")
	}
	fmt.Printf("   服务端返回类型: %T\n", getResponse)

	fmt.Println("\n10. POST 请求 wait() 模式读取 body:")
	_, err = page.Intercept().Start(nil, nil, []string{"beforeRequestSent"})
	if err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url, {
			method: "POST",
			headers: {"Content-Type": "application/json"},
			body: JSON.stringify({mode: "queue", value: 99})
		}).catch(function(){ return null; });
		return true;
	}`, server.GetURL("/api/echo")); err != nil {
		return err
	}
	queuedPost := page.Intercept().Wait(8 * time.Second)
	if queuedPost == nil {
		return fmt.Errorf("POST 队列模式未捕获到请求")
	}
	fmt.Printf("   wait捕获POST: %s %s body=%s\n", queuedPost.Method, queuedPost.URL, queuedPost.Body())
	if err := queuedPost.ContinueRequest("", "", nil, nil); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)
	page.Intercept().Stop()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有高级网络功能测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func mapValue(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return fmt.Sprint(value)
		}
	}
	return ""
}

func mergeRequestHeaders(current map[string]string, overrides map[string]string) []map[string]any {
	merged := make(map[string]string, len(current)+len(overrides))
	names := make(map[string]string, len(current)+len(overrides))
	for name, value := range current {
		lower := strings.ToLower(name)
		merged[lower] = value
		names[lower] = name
	}
	for name, value := range overrides {
		lower := strings.ToLower(name)
		merged[lower] = value
		names[lower] = name
	}

	keys := make([]string, 0, len(merged))
	for lower := range merged {
		keys = append(keys, lower)
	}
	sort.Strings(keys)

	rows := make([]map[string]any, 0, len(keys))
	for _, lower := range keys {
		rows = append(rows, map[string]any{
			"name":  names[lower],
			"value": map[string]any{"type": "string", "value": merged[lower]},
		})
	}
	return rows
}
