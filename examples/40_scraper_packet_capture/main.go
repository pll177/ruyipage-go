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
	fmt.Println("Example 40: Scraper Packet Capture")
	fmt.Println(strings.Repeat("=", 70))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(9632))
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
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	if err := page.Get("about:blank"); err != nil {
		return err
	}

	collector, err := page.Network().AddDataCollector(
		[]string{"beforeRequestSent", "responseCompleted"},
		[]string{"request", "response"},
		0,
	)
	if err != nil {
		return err
	}
	defer func() {
		if collector != nil {
			_ = collector.Remove()
		}
	}()

	if err := page.Listen().Start("/api/data", false, "GET"); err != nil {
		return err
	}
	if _, err := page.Intercept().Start(nil, nil, []string{"beforeRequestSent"}); err != nil {
		return err
	}
	if _, err := page.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, server.GetURL("/api/data")); err != nil {
		return err
	}
	getReq := page.Intercept().Wait(8 * time.Second)
	getRequestID := ""
	if getReq != nil {
		getRequestID = getReq.RequestID
		_ = getReq.ContinueRequest("", "", nil, nil)
	}
	getPacket := page.Listen().Wait(8 * time.Second)
	page.Intercept().Stop()
	page.Listen().Stop()

	if getReq != nil && getReq.Method == "GET" {
		exampleutil.AddCheck(&results, "GET request captured", "成功", getReq.URL)
	} else {
		exampleutil.AddCheck(&results, "GET request captured", "失败", fmt.Sprintf("%v", getReq))
	}
	if getPacket != nil && getPacket.Status == 200 {
		exampleutil.AddCheck(&results, "GET response status", "成功", fmt.Sprintf("%d", getPacket.Status))
	} else {
		status := 0
		if getPacket != nil {
			status = getPacket.Status
		}
		exampleutil.AddCheck(&results, "GET response status", "失败", fmt.Sprintf("%d", status))
	}

	getResponseText := ""
	if collector != nil && getRequestID != "" {
		data, err := collector.Get(getRequestID, "response")
		if err == nil {
			getResponseText = exampleutil.DecodeNetworkText(data)
		}
	}
	if strings.Contains(getResponseText, `"status":"ok"`) || strings.Contains(getResponseText, `"status": "ok"`) {
		exampleutil.AddCheck(&results, "GET response body", "成功", truncate(getResponseText, 120))
	} else {
		exampleutil.AddCheck(&results, "GET response body", "失败", truncate(getResponseText, 120))
	}

	postBodies := make([]string, 0, 1)
	postRequestIDs := make([]string, 0, 1)
	postHandler := func(req *ruyipage.InterceptedRequest) {
		if strings.Contains(req.URL, "/api/echo") && req.Method == "POST" {
			postBodies = append(postBodies, req.Body())
			postRequestIDs = append(postRequestIDs, req.RequestID)
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}

	if err := page.Listen().Start("/api/echo", false, "POST"); err != nil {
		return err
	}
	if _, err := page.Intercept().StartRequests(postHandler, nil); err != nil {
		return err
	}
	postResult, err := page.RunJS(`function(url){
		return fetch(url, {
			method: "POST",
			headers: {"Content-Type": "application/json"},
			body: JSON.stringify({keyword: "ruyi", page: 2})
		})
			.then(function(response){ return response.json(); })
			.catch(function(error){ return {error: String(error)}; });
	}`, server.GetURL("/api/echo"))
	if err != nil {
		return err
	}
	postPacket := page.Listen().Wait(8 * time.Second)
	page.Intercept().Stop()
	page.Listen().Stop()

	postBody := ""
	if len(postBodies) > 0 {
		postBody = postBodies[0]
	}
	if postBody == `{"keyword":"ruyi","page":2}` {
		exampleutil.AddCheck(&results, "POST request body", "成功", postBody)
	} else {
		exampleutil.AddCheck(&results, "POST request body", "失败", postBody)
	}
	if postPacket != nil && postPacket.Status == 200 {
		exampleutil.AddCheck(&results, "POST response status", "成功", fmt.Sprintf("%d", postPacket.Status))
	} else {
		status := 0
		if postPacket != nil {
			status = postPacket.Status
		}
		exampleutil.AddCheck(&results, "POST response status", "失败", fmt.Sprintf("%d", status))
	}

	postResponseText := ""
	if collector != nil && len(postRequestIDs) > 0 {
		data, err := collector.Get(postRequestIDs[0], "response")
		if err == nil {
			postResponseText = exampleutil.DecodeNetworkText(data)
		}
	}
	if strings.Contains(postResponseText, `"body":"{\"keyword\":\"ruyi\",\"page\":2}"`) || strings.Contains(postResponseText, `"body": "{\"keyword\":\"ruyi\",\"page\":2}"`) {
		exampleutil.AddCheck(&results, "POST response body", "成功", truncate(postResponseText, 120))
	} else {
		exampleutil.AddCheck(&results, "POST response body", "失败", truncate(postResponseText, 120))
	}

	if postMap, ok := postResult.(map[string]any); ok && fmt.Sprint(postMap["body"]) == `{"keyword":"ruyi","page":2}` {
		exampleutil.AddCheck(&results, "POST page result", "成功", fmt.Sprint(postMap["body"]))
	} else {
		exampleutil.AddCheck(&results, "POST page result", "失败", fmt.Sprintf("%v", postResult))
	}

	exampleutil.PrintChecks(results)
	for _, row := range results {
		if row.Status == "失败" {
			return fmt.Errorf("存在失败项: %s", row.Item)
		}
	}
	return nil
}

func truncate(value string, size int) string {
	if len(value) <= size {
		return value
	}
	return value[:size]
}
