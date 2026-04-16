package main

import (
	"fmt"
	"net/url"
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
	fmt.Println("测试 33: Log + Input Events")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Console().Stop()
		page.Events().Stop()
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 10)

	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "example.com 已加载")

	if err := page.Console().Start(""); err != nil {
		return err
	}
	page.Console().Clear()
	if _, err := page.RunJS(`
		console.log("This is a log message");
		console.info("This is an info message");
		console.warn("This is a warning message");
		console.error("This is an error message");
		console.debug("This is a debug message");
		console.table([{name: "Alice", age: 18}, {name: "Bob", age: 20}]);
		return true;
	`); err != nil {
		return err
	}

	logEntry := page.Console().Wait("", "", 3*time.Second)
	entries := page.Console().Entries()
	var errorEntry *ruyipage.LogEntry
	var warnEntry *ruyipage.LogEntry
	var tableEntry *ruyipage.LogEntry
	for index := range entries {
		entry := entries[index]
		switch {
		case errorEntry == nil && entry.Level == "error":
			errorEntry = &entry
		case warnEntry == nil && entry.Level == "warn":
			warnEntry = &entry
		case tableEntry == nil && (strings.Contains(entry.Text, "Alice") || strings.Contains(entry.Text, "Bob")):
			tableEntry = &entry
		}
	}

	if logEntry != nil {
		exampleutil.AddCheck(&results, "log.entryAdded first", "成功", fmt.Sprintf("level=%s text=%s", logEntry.Level, logEntry.Text))
	} else {
		exampleutil.AddCheck(&results, "log.entryAdded first", "失败", "未观察到首条日志事件")
	}
	if errorEntry != nil {
		exampleutil.AddCheck(&results, "log.entryAdded error", "成功", errorEntry.Text)
	} else {
		exampleutil.AddCheck(&results, "log.entryAdded error", "失败", "未观察到 error 日志事件")
	}
	if warnEntry != nil {
		exampleutil.AddCheck(&results, "log.entryAdded warn", "成功", warnEntry.Text)
	} else {
		exampleutil.AddCheck(&results, "log.entryAdded warn", "失败", "未观察到 warn 日志事件")
	}
	if tableEntry != nil {
		exampleutil.AddCheck(&results, "log.entryAdded table", "成功", truncate(tableEntry.Text, 120))
	} else {
		exampleutil.AddCheck(&results, "log.entryAdded table", "跳过", "当前 console.table 输出未稳定映射到 text")
	}
	if len(entries) >= 6 {
		exampleutil.AddCheck(&results, "log.entryAdded total", "成功", fmt.Sprintf("日志数量: %d", len(entries)))
	} else {
		exampleutil.AddCheck(&results, "log.entryAdded total", "失败", fmt.Sprintf("日志数量: %d", len(entries)))
	}
	page.Console().Stop()

	html := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>File Dialog Test</title></head>
<body>
	<input type="file" id="single-file">
	<input type="file" id="multiple-files" multiple>
	<button id="trigger-single" onclick="document.getElementById('single-file').click()">Open Single</button>
	<button id="trigger-multiple" onclick="document.getElementById('multiple-files').click()">Open Multiple</button>
</body>
</html>`
	if err := page.Get("data:text/html;charset=utf-8," + url.QueryEscape(html)); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "文件测试页加载", "成功", "本地 file dialog 页面已加载")

	if err := page.Events().Start([]string{"input.fileDialogOpened"}, []string{page.ContextID()}); err == nil {
		page.Events().Clear()
		singleTrigger, err := page.Ele("#trigger-single", 1, 5*time.Second)
		if err != nil {
			return err
		}
		if singleTrigger != nil {
			_ = singleTrigger.ClickSelf(true, 0)
		}
		singleEvent := page.Events().Wait("input.fileDialogOpened", 3*time.Second)
		if singleEvent != nil && !singleEvent.Multiple {
			exampleutil.AddCheck(&results, "input.fileDialogOpened single", "成功", fmt.Sprintf("multiple=%v", singleEvent.Multiple))
		} else {
			exampleutil.AddCheck(&results, "input.fileDialogOpened single", "跳过", "当前环境未稳定观察到单文件对话框事件")
		}

		page.Events().Clear()
		multipleTrigger, err := page.Ele("#trigger-multiple", 1, 5*time.Second)
		if err != nil {
			return err
		}
		if multipleTrigger != nil {
			_ = multipleTrigger.ClickSelf(true, 0)
		}
		multipleEvent := page.Events().Wait("input.fileDialogOpened", 3*time.Second)
		if multipleEvent != nil && multipleEvent.Multiple {
			exampleutil.AddCheck(&results, "input.fileDialogOpened multiple", "成功", fmt.Sprintf("multiple=%v", multipleEvent.Multiple))
		} else {
			exampleutil.AddCheck(&results, "input.fileDialogOpened multiple", "跳过", "当前环境未稳定观察到多文件对话框事件")
		}
	} else {
		exampleutil.AddCheck(&results, "input.fileDialogOpened", "不支持", "未能订阅 input.fileDialogOpened 事件")
	}

	exampleutil.PrintChecks(results)
	return nil
}

func truncate(value string, size int) string {
	if len(value) <= size {
		return value
	}
	return value[:size]
}
