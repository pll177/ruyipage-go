package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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
	fmt.Println("测试 23: Download 下载管理")
	fmt.Println(strings.Repeat("=", 70))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(9530))
	if err := server.Start(); err != nil {
		return err
	}
	defer func() {
		_ = server.Stop()
	}()

	outputDir, err := exampleutil.OutputDir("23_download")
	if err != nil {
		return err
	}
	downloadPath := filepath.Join(outputDir, "downloads")
	textPath := filepath.Join(downloadPath, "test.txt")
	jsonPath := filepath.Join(downloadPath, "test.json")
	_ = os.RemoveAll(downloadPath)
	if err := os.MkdirAll(downloadPath, 0o755); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(downloadPath)
	}()

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Downloads().Stop()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 10)
	baseURL := strings.TrimRight(server.GetURL(""), "/")

	if err := page.Downloads().SetBehavior("allow", downloadPath, nil, nil); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "browser.setDownloadBehavior allow", "成功", "下载目录: "+downloadPath)

	if err := page.Get("data:text/html;charset=utf-8," + url.QueryEscape(buildDownloadPage(baseURL))); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "下载测试页加载", "成功", "服务地址: "+baseURL)

	if err := page.Downloads().Start(); err != nil {
		exampleutil.AddCheck(&results, "下载事件订阅", "失败", err.Error())
		exampleutil.PrintChecks(results)
		return nil
	}
	exampleutil.AddCheck(&results, "下载事件订阅", "成功", "已订阅 downloadWillBegin / downloadEnd")

	page.Downloads().Clear()
	if err := clickBySelector(page, "#download-text"); err != nil {
		return err
	}
	beginText, endText := page.Downloads().WaitChain("test.txt", 5*time.Second)
	if beginText != nil {
		exampleutil.AddCheck(&results, "downloadWillBegin text", "成功", "filename="+beginText.SuggestedFilename)
	} else {
		exampleutil.AddCheck(&results, "downloadWillBegin text", "失败", "5 秒内未收到开始事件")
	}
	if endText != nil {
		exampleutil.AddCheck(&results, "downloadEnd text", "成功", "status="+endText.Status)
	} else {
		exampleutil.AddCheck(&results, "downloadEnd text", "失败", "5 秒内未收到结束事件")
	}
	if page.Downloads().WaitFile(textPath, 3*time.Second, 1) {
		info, _ := os.Stat(textPath)
		exampleutil.AddCheck(&results, "text 文件落盘", "成功", fmt.Sprintf("%d bytes", info.Size()))
	} else {
		exampleutil.AddCheck(&results, "text 文件落盘", "失败", "事件触发后仍未观察到文件落盘")
	}

	page.Downloads().Clear()
	if err := clickBySelector(page, "#download-json"); err != nil {
		return err
	}
	beginJSON, endJSON := page.Downloads().WaitChain("test.json", 5*time.Second)
	if beginJSON != nil {
		exampleutil.AddCheck(&results, "downloadWillBegin json", "成功", "filename="+beginJSON.SuggestedFilename)
	} else {
		exampleutil.AddCheck(&results, "downloadWillBegin json", "失败", "5 秒内未收到开始事件")
	}
	if endJSON != nil {
		exampleutil.AddCheck(&results, "downloadEnd json", "成功", "status="+endJSON.Status)
	} else {
		exampleutil.AddCheck(&results, "downloadEnd json", "失败", "5 秒内未收到结束事件")
	}
	if page.Downloads().WaitFile(jsonPath, 3*time.Second, 1) {
		info, _ := os.Stat(jsonPath)
		exampleutil.AddCheck(&results, "json 文件落盘", "成功", fmt.Sprintf("%d bytes", info.Size()))
	} else {
		exampleutil.AddCheck(&results, "json 文件落盘", "失败", "事件触发后仍未观察到文件落盘")
	}

	page.Downloads().Clear()
	if err := page.Downloads().SetBehavior("deny", "", nil, nil); err != nil {
		exampleutil.AddCheck(&results, "browser.setDownloadBehavior deny", "失败", err.Error())
	} else {
		if err := clickBySelector(page, "#download-text"); err != nil {
			return err
		}
		deniedBegin := page.Downloads().Wait("browsingContext.downloadWillBegin", 2*time.Second, "test.txt", "")
		if deniedBegin == nil {
			exampleutil.AddCheck(&results, "browser.setDownloadBehavior deny", "成功", "deny 模式下未观察到下载开始事件")
		} else {
			exampleutil.AddCheck(&results, "browser.setDownloadBehavior deny", "失败", "deny 模式下仍出现下载开始事件")
		}
	}

	exampleutil.PrintChecks(results)
	return nil
}

func clickBySelector(page *ruyipage.FirefoxPage, selector string) error {
	element, err := page.Ele(selector, 1, 5*time.Second)
	if err != nil {
		return err
	}
	if element == nil {
		return fmt.Errorf("未找到元素: %s", selector)
	}
	return element.ClickSelf(false, 0)
}

func buildDownloadPage(baseURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>下载测试</title></head>
<body>
	<button id="download-text">下载文本文件</button>
	<button id="download-json">下载JSON文件</button>
	<div id="status"></div>
	<script>
		function triggerDownload(targetURL, statusText) {
			const a = document.createElement("a");
			a.href = targetURL;
			a.click();
			document.getElementById("status").textContent = statusText;
		}
		document.getElementById("download-text").onclick = function() {
			triggerDownload(%q, "文本文件下载已触发");
		};
		document.getElementById("download-json").onclick = function() {
			triggerDownload(%q, "JSON文件下载已触发");
		};
	</script>
</body>
</html>`, baseURL+"/download/text", baseURL+"/download/json")
}
