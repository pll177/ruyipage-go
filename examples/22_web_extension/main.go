package main

import (
	"archive/zip"
	"encoding/json"
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
	fmt.Println("测试 22: WebExtension 模块")
	fmt.Println(strings.Repeat("=", 70))

	outputDir, err := exampleutil.OutputDir("22_web_extension")
	if err != nil {
		return err
	}
	extDir := filepath.Join(outputDir, "test_extension")
	xpiPath := filepath.Join(outputDir, "test_extension.xpi")

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Extensions().UninstallAll()
		_ = page.Quit(0, false)
		_ = os.RemoveAll(extDir)
		_ = os.Remove(xpiPath)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	if err := createTestExtension(extDir); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "创建测试扩展", "成功", extDir)

	extID := safeInstall(page, extDir, "安装目录扩展", &results)
	if extID != "" {
		if err := page.Get("https://example.com"); err != nil {
			exampleutil.AddCheck(&results, "目录扩展生效验证", "失败", err.Error())
		} else {
			page.Wait().Sleep(1500 * time.Millisecond)
			marker, _ := page.RunJSExpr("document.documentElement.getAttribute('data-ruyi-extension')")
			if fmt.Sprint(marker) == "loaded" {
				exampleutil.AddCheck(&results, "目录扩展生效验证", "成功", "content script 已注入")
			} else {
				exampleutil.AddCheck(&results, "目录扩展生效验证", "失败", fmt.Sprintf("marker=%v", marker))
			}
		}
		if err := page.Extensions().Uninstall(extID); err != nil {
			exampleutil.AddCheck(&results, "卸载目录扩展", "失败", err.Error())
		} else {
			exampleutil.AddCheck(&results, "卸载目录扩展", "成功", extID)
		}
	} else {
		exampleutil.AddCheck(&results, "目录扩展生效验证", "跳过", "安装未成功")
		exampleutil.AddCheck(&results, "卸载目录扩展", "跳过", "安装未成功")
	}

	if err := packExtension(extDir, xpiPath); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "打包 XPI", "成功", xpiPath)

	xpiID := safeInstallArchive(page, xpiPath, &results)
	if xpiID != "" {
		if err := page.Get("https://example.com"); err != nil {
			exampleutil.AddCheck(&results, "XPI 扩展生效验证", "失败", err.Error())
		} else {
			page.Wait().Sleep(1500 * time.Millisecond)
			marker, _ := page.RunJSExpr("document.documentElement.getAttribute('data-ruyi-extension')")
			if fmt.Sprint(marker) == "loaded" {
				exampleutil.AddCheck(&results, "XPI 扩展生效验证", "成功", "content script 已注入")
			} else {
				exampleutil.AddCheck(&results, "XPI 扩展生效验证", "失败", fmt.Sprintf("marker=%v", marker))
			}
		}
		if err := page.Extensions().Uninstall(xpiID); err != nil {
			exampleutil.AddCheck(&results, "卸载 XPI 扩展", "失败", err.Error())
		} else {
			exampleutil.AddCheck(&results, "卸载 XPI 扩展", "成功", xpiID)
		}
	} else {
		exampleutil.AddCheck(&results, "XPI 扩展生效验证", "跳过", "安装未成功")
		exampleutil.AddCheck(&results, "卸载 XPI 扩展", "跳过", "安装未成功")
	}

	exampleutil.PrintChecks(results)
	return nil
}

func createTestExtension(extDir string) error {
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return err
	}
	manifest := map[string]any{
		"manifest_version": 2,
		"name":             "RuyiPage Go Test Extension",
		"version":          "1.0.0",
		"description":      "测试扩展",
		"browser_specific_settings": map[string]any{
			"gecko": map[string]any{
				"id":                 "test-go@ruyipage.local",
				"strict_min_version": "109.0",
			},
		},
		"background": map[string]any{"scripts": []string{"background.js"}},
		"content_scripts": []map[string]any{
			{
				"matches": []string{"<all_urls>"},
				"js":      []string{"content.js"},
				"run_at":  "document_end",
			},
		},
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(extDir, "manifest.json"), manifestBytes, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(extDir, "background.js"), []byte("console.log('RuyiPage Go Test Extension loaded!');\n"), 0o644); err != nil {
		return err
	}
	return os.WriteFile(
		filepath.Join(extDir, "content.js"),
		[]byte("document.documentElement.setAttribute('data-ruyi-extension', 'loaded');\n"),
		0o644,
	)
}

func packExtension(extDir string, xpiPath string) error {
	file, err := os.Create(xpiPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	archive := zip.NewWriter(file)
	defer func() {
		_ = archive.Close()
	}()

	for _, name := range []string{"manifest.json", "background.js", "content.js"} {
		sourcePath := filepath.Join(extDir, name)
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		writer, err := archive.Create(name)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}
	return archive.Close()
}

func safeInstall(page *ruyipage.FirefoxPage, path string, label string, results *[]exampleutil.CheckRow) string {
	extID, err := page.Extensions().InstallDir(path)
	if err == nil && extID != "" {
		exampleutil.AddCheck(results, label, "成功", "extension_id="+extID)
		return extID
	}
	if err == nil {
		exampleutil.AddCheck(results, label, "不支持", "当前 Firefox 未返回 extension id")
		return ""
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "unknown command") || strings.Contains(message, "not supported") {
		exampleutil.AddCheck(results, label, "不支持", "当前 Firefox 不支持 webExtension.install")
		return ""
	}
	exampleutil.AddCheck(results, label, "失败", err.Error())
	return ""
}

func safeInstallArchive(page *ruyipage.FirefoxPage, path string, results *[]exampleutil.CheckRow) string {
	extID, err := page.Extensions().InstallArchive(path)
	if err == nil && extID != "" {
		exampleutil.AddCheck(results, "安装 XPI 扩展", "成功", "extension_id="+extID)
		return extID
	}
	if err == nil {
		exampleutil.AddCheck(results, "安装 XPI 扩展", "不支持", "当前 Firefox 未返回 extension id")
		return ""
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "unknown command") || strings.Contains(message, "not supported") {
		exampleutil.AddCheck(results, "安装 XPI 扩展", "不支持", "当前 Firefox 不支持 webExtension.install")
		return ""
	}
	exampleutil.AddCheck(results, "安装 XPI 扩展", "失败", err.Error())
	return ""
}
