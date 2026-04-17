package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	fmt.Println("测试 29: Script + Input 高级能力")
	fmt.Println(strings.Repeat("=", 70))

	outputDir, err := exampleutil.OutputDir("29_script_input_advanced")
	if err != nil {
		return err
	}
	htmlPath := filepath.Join(outputDir, "test_file_input.html")
	file1 := filepath.Join(outputDir, "test1.txt")
	file2 := filepath.Join(outputDir, "test2.txt")
	defer func() {
		_ = os.Remove(htmlPath)
		_ = os.Remove(file1)
		_ = os.Remove(file2)
	}()

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "页面加载", "成功", "example.com 已加载")

	realms, err := page.GetRealms("")
	if err != nil {
		return err
	}
	if len(realms) > 0 {
		exampleutil.AddCheck(&results, "script.getRealms all", "成功", fmt.Sprintf("realm 数量: %d", len(realms)))
	} else {
		exampleutil.AddCheck(&results, "script.getRealms all", "失败", "未返回任何 realm")
	}

	windowRealms, err := page.GetRealms("window")
	if err != nil {
		return err
	}
	if len(windowRealms) > 0 {
		exampleutil.AddCheck(&results, "script.getRealms window", "成功", fmt.Sprintf("window realm 数量: %d", len(windowRealms)))
	} else {
		exampleutil.AddCheck(&results, "script.getRealms window", "失败", "未返回 window realm")
	}

	singleHandle, err := page.EvalHandle(`({data: "test", array: [1, 2, 3]})`, true)
	if err != nil {
		return err
	}
	if singleHandle.Success() && singleHandle.Result.Handle != "" {
		if err := page.DisownHandles([]string{singleHandle.Result.Handle}); err != nil {
			exampleutil.AddCheck(&results, "script.disown single", "失败", err.Error())
		} else {
			exampleutil.AddCheck(&results, "script.disown single", "成功", "handle="+singleHandle.Result.Handle)
		}
	} else {
		exampleutil.AddCheck(&results, "script.disown single", "跳过", "脚本结果未返回 handle")
	}

	handles := make([]string, 0, 3)
	for index := 0; index < 3; index++ {
		result, evalErr := page.EvalHandle(fmt.Sprintf(`({id: %d, value: "test%d"})`, index, index), true)
		if evalErr != nil {
			return evalErr
		}
		if result.Success() && result.Result.Handle != "" {
			handles = append(handles, result.Result.Handle)
		}
	}
	if len(handles) > 0 {
		if err := page.DisownHandles(handles); err != nil {
			exampleutil.AddCheck(&results, "script.disown batch", "失败", err.Error())
		} else {
			exampleutil.AddCheck(&results, "script.disown batch", "成功", fmt.Sprintf("句柄数量: %d", len(handles)))
		}
	} else {
		exampleutil.AddCheck(&results, "script.disown batch", "跳过", "未拿到可用 handle")
	}

	if err := os.WriteFile(htmlPath, []byte(fileInputHTML), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(file1, []byte("Test file 1 content"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(file2, []byte("Test file 2 content"), 0o644); err != nil {
		return err
	}

	fileURL, err := exampleutil.FileURLFromPath(htmlPath)
	if err != nil {
		return err
	}
	if err := page.Get(fileURL); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "文件测试页加载", "成功", fileURL)

	singleInput, err := page.Ele("#single-file", 1, 5*time.Second)
	if err != nil {
		return err
	}
	multipleInput, err := page.Ele("#multiple-files", 1, 5*time.Second)
	if err != nil {
		return err
	}
	if singleInput == nil || multipleInput == nil {
		return fmt.Errorf("未找到 file input 元素")
	}

	if err := singleInput.UploadFiles(file1); err != nil {
		exampleutil.AddCheck(&results, "input.setFiles single", "失败", err.Error())
	} else {
		resultText, _ := page.RunJSExpr(`document.getElementById("result").textContent`)
		if strings.Contains(fmt.Sprint(resultText), "test1.txt") {
			exampleutil.AddCheck(&results, "input.setFiles single", "成功", fmt.Sprint(resultText))
		} else {
			exampleutil.AddCheck(&results, "input.setFiles single", "失败", fmt.Sprint(resultText))
		}
	}

	if err := multipleInput.UploadFiles(file1, file2); err != nil {
		exampleutil.AddCheck(&results, "input.setFiles multiple", "失败", err.Error())
	} else {
		resultText, _ := page.RunJSExpr(`document.getElementById("result").textContent`)
		if strings.Contains(fmt.Sprint(resultText), "Multiple files: 2") {
			exampleutil.AddCheck(&results, "input.setFiles multiple", "成功", fmt.Sprint(resultText))
		} else {
			exampleutil.AddCheck(&results, "input.setFiles multiple", "失败", fmt.Sprint(resultText))
		}
	}

	exampleutil.PrintChecks(results)
	return nil
}

const fileInputHTML = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>File Input Test</title></head>
<body>
	<input type="file" id="single-file">
	<input type="file" id="multiple-files" multiple>
	<div id="result"></div>
	<script>
		document.getElementById("single-file").addEventListener("change", function(event) {
			document.getElementById("result").textContent = "Single file: " + event.target.files[0].name;
		});
		document.getElementById("multiple-files").addEventListener("change", function(event) {
			document.getElementById("result").textContent = "Multiple files: " + event.target.files.length;
		});
	</script>
</body>
</html>`
