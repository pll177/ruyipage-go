package main

import (
	"fmt"
	"sort"
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
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试11: 网络拦截")
	fmt.Println(strings.Repeat("=", 60))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(8888))
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
		page.Wait().Sleep(time.Second)
		_ = page.Quit(0, false)
	}()

	if err := page.Get("about:blank"); err != nil {
		return err
	}
	page.Wait().Sleep(500 * time.Millisecond)

	fmt.Println("\n1. 拦截并 Mock 本地 API:")
	_, err = page.Intercept().Start(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/data") {
			fmt.Printf("   拦截到请求: %s %s\n", req.Method, req.URL)
			_ = req.Mock(
				`{"status":"ok","data":{"message":"mocked-by-interceptor"}}`,
				200,
				map[string]string{
					"content-type":                "application/json",
					"access-control-allow-origin": "*",
				},
				"",
			)
			fmt.Println("   ✓ 已返回 Mock 响应")
			return
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil, []string{"beforeRequestSent"})
	if err != nil {
		return err
	}
	mocked, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(response){ return response.json(); })
			.then(function(data){ return data.data.message; })
			.catch(function(error){ return String(error); });
	}`, server.GetURL("/api/data"))
	if err != nil {
		return err
	}
	if fmt.Sprint(mocked) != "mocked-by-interceptor" {
		return fmt.Errorf("mock 结果异常: %v", mocked)
	}
	fmt.Printf("   mock结果: %v\n", mocked)
	page.Intercept().Stop()

	fmt.Println("\n2. 阻止本地 API 请求:")
	_, err = page.Intercept().Start(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/data") {
			fmt.Printf("   阻止请求: %s\n", req.URL)
			_ = req.Fail()
			return
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil, []string{"beforeRequestSent"})
	if err != nil {
		return err
	}
	blocked, err := page.RunJS(`function(url){
		return fetch(url)
			.then(function(){ return "unexpected-success"; })
			.catch(function(error){ return "blocked:" + error.name; });
	}`, server.GetURL("/api/data"))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(fmt.Sprint(blocked), "blocked:") {
		return fmt.Errorf("fail 结果异常: %v", blocked)
	}
	fmt.Printf("   fail结果: %v\n", blocked)
	page.Intercept().Stop()

	fmt.Println("\n3. 修改请求头并继续:")
	headerErrs := make(chan error, 1)
	_, err = page.Intercept().Start(func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/headers") {
			headers := mergeRequestHeaders(req.Headers, map[string]string{
				"X-Ruyi-Demo": "yes",
				"User-Agent":  "RuyiPage/Example11",
			})
			if err := req.ContinueRequest("", "", headers, nil); err != nil {
				select {
				case headerErrs <- err:
				default:
				}
				return
			}
			fmt.Printf("   ✓ 已注入自定义请求头（method=%s）\n", req.Method)
			return
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil, []string{"beforeRequestSent"})
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
		return fmt.Errorf("continue_request 修改请求头失败: %w", err)
	default:
	}
	headers, ok := headersValue.(map[string]any)
	if !ok {
		return fmt.Errorf("请求头结果类型异常: %T", headersValue)
	}
	xHeader := valueFromMap(headers, "X-Ruyi-Demo", "x-ruyi-demo")
	if xHeader != "yes" {
		return fmt.Errorf("X-Ruyi-Demo 异常: %q, 全量响应=%v", xHeader, headers)
	}
	fmt.Printf("   服务器看到的 X-Ruyi-Demo: %s\n", xHeader)
	page.Intercept().Stop()

	fmt.Println("\n4. 队列模式 wait() 捕获拦截请求:")
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
	req := page.Intercept().Wait(5 * time.Second)
	if req == nil {
		return fmt.Errorf("wait 未在超时内捕获请求")
	}
	fmt.Printf("   wait捕获: %s %s\n", req.Method, req.URL)
	if err := req.ContinueRequest("", "", nil, nil); err != nil {
		return err
	}
	page.Intercept().Stop()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 网络拦截测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func valueFromMap(values map[string]any, keys ...string) string {
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
