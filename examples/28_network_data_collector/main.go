package main

import (
	"fmt"
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
	fmt.Println("测试 28: Network Data Collector")
	fmt.Println(strings.Repeat("=", 70))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(9550))
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
		page.Listen().Stop()
		page.Intercept().Stop()
		_ = page.Network().ClearExtraHeaders()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 10)
	var collector *ruyipage.DataCollector
	defer func() {
		if collector != nil {
			_ = collector.Remove()
		}
	}()

	baseURL := strings.TrimRight(server.GetURL(""), "/")
	if err := page.Get("about:blank"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "about:blank 已加载，服务地址: "+baseURL)

	if err := page.Network().SetExtraHeaders(map[string]string{"X-Test-Collector": "yes"}); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "network.setExtraHeaders", "成功", "已设置 X-Test-Collector")

	if err := page.Network().SetCacheBehavior("bypass"); err != nil {
		exampleutil.AddCheck(&results, "network.setCacheBehavior", "不支持", err.Error())
	} else {
		exampleutil.AddCheck(&results, "network.setCacheBehavior", "成功", "缓存行为已设为 bypass")
	}

	collector, err = page.Network().AddDataCollector([]string{"responseCompleted"}, []string{"response"}, 0)
	if err != nil {
		return err
	}
	if collector.ID != "" {
		exampleutil.AddCheck(&results, "network.addDataCollector", "成功", "collector="+collector.ID)
	} else {
		exampleutil.AddCheck(&results, "network.addDataCollector", "失败", "未返回 collector ID")
	}

	if _, err := page.Intercept().StartRequests(nil, nil); err != nil {
		return err
	}
	if err := page.Listen().Start("/api/collector", false, "GET"); err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, baseURL+"/api/collector"); err != nil {
		return err
	}

	req := page.Intercept().Wait(8 * time.Second)
	if req != nil {
		hasHeader := false
		for key := range req.Headers {
			if strings.EqualFold(key, "X-Test-Collector") {
				hasHeader = true
				break
			}
		}
		exampleutil.AddCheck(&results, "request header injected", statusOf(hasHeader), fmt.Sprintf("request_id=%s, X-Test-Collector=%v", req.RequestID, hasHeader))
		if err := req.ContinueRequest("", "", nil, nil); err != nil {
			exampleutil.AddCheck(&results, "continue intercepted request", "失败", err.Error())
		}
	} else {
		exampleutil.AddCheck(&results, "request header injected", "失败", "未捕获到 beforeRequestSent 请求")
	}

	packet := page.Listen().Wait(8 * time.Second)
	if packet != nil {
		exampleutil.AddCheck(&results, "network.responseCompleted observed", "成功", fmt.Sprintf("status=%d url=%s", packet.Status, packet.URL))
	} else {
		exampleutil.AddCheck(&results, "network.responseCompleted observed", "失败", "未在超时内观察到 responseCompleted")
	}

	if collector != nil && req != nil {
		data, getErr := collector.Get(req.RequestID, "response")
		if getErr != nil {
			exampleutil.AddCheck(&results, "network.getData response", "失败", getErr.Error())
		} else if data != nil && data.HasData() {
			exampleutil.AddCheck(&results, "network.getData response", "成功", truncate(exampleutil.DecodeNetworkText(data), 120))
			if err := collector.Disown(req.RequestID, "response"); err != nil {
				exampleutil.AddCheck(&results, "network.disownData", "失败", err.Error())
			} else {
				exampleutil.AddCheck(&results, "network.disownData", "成功", "已释放 response 数据")
				dataAfter, afterErr := collector.Get(req.RequestID, "response")
				if afterErr != nil || dataAfter == nil || !dataAfter.HasData() {
					note := "释放后已无可用数据"
					if afterErr != nil {
						note = "释放后读取报错: " + truncate(afterErr.Error(), 100)
					}
					exampleutil.AddCheck(&results, "network.getData after disown", "成功", note)
				} else {
					exampleutil.AddCheck(&results, "network.getData after disown", "跳过", truncate(exampleutil.DecodeNetworkText(dataAfter), 120))
				}
			}
		} else {
			exampleutil.AddCheck(&results, "network.getData response", "失败", fmt.Sprintf("raw=%v", data))
			exampleutil.AddCheck(&results, "network.disownData", "跳过", "未成功拿到 response 数据，跳过释放验证")
			exampleutil.AddCheck(&results, "network.getData after disown", "跳过", "未成功拿到 response 数据，跳过释放后验证")
		}
	}

	if collector != nil {
		if err := collector.Remove(); err != nil {
			exampleutil.AddCheck(&results, "network.removeDataCollector", "失败", err.Error())
		} else {
			exampleutil.AddCheck(&results, "network.removeDataCollector", "成功", "已移除 "+collector.ID)
			collector = nil
		}
	}

	if err := page.Network().ClearExtraHeaders(); err != nil {
		exampleutil.AddCheck(&results, "clear extra headers", "失败", err.Error())
	} else {
		exampleutil.AddCheck(&results, "clear extra headers", "成功", "额外请求头已清理")
	}

	exampleutil.PrintChecks(results)
	return nil
}

func statusOf(ok bool) string {
	if ok {
		return "成功"
	}
	return "失败"
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
