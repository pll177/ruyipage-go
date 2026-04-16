package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

const exampleName = "19_pdf_printing"

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试19: PDF打印功能")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(time.Second)
		_ = page.Quit(0, false)
	}()

	testURL, err := exampleutil.TestPageURL("test_page.html")
	if err != nil {
		return err
	}
	if err := page.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)

	outputDir, err := exampleutil.OutputDir(exampleName)
	if err != nil {
		return err
	}

	fmt.Println("\n1. 基本PDF打印:")
	basicPDF := filepath.Join(outputDir, "basic.pdf")
	if _, err := page.PDF(basicPDF, nil); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已保存: %s (%d bytes)\n", basicPDF, fileSize(basicPDF))

	fmt.Println("\n2. A4 + 背景:")
	a4PDF := filepath.Join(outputDir, "a4_bg.pdf")
	if _, err := page.PDF(a4PDF, map[string]any{
		"page":       map[string]any{"width": 21.0, "height": 29.7},
		"background": true,
	}); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已保存: %s (%d bytes)\n", a4PDF, fileSize(a4PDF))

	fmt.Println("\n3. 横向 + 缩放 + 页边距:")
	landscapePDF := filepath.Join(outputDir, "landscape_scaled.pdf")
	if _, err := page.PDF(landscapePDF, map[string]any{
		"orientation": "landscape",
		"scale":       0.9,
		"margin": map[string]any{
			"top":    1.2,
			"bottom": 1.2,
			"left":   1.0,
			"right":  1.0,
		},
	}); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已保存: %s (%d bytes)\n", landscapePDF, fileSize(landscapePDF))

	fmt.Println("\n4. 指定页范围 + shrinkToFit:")
	rangePDF := filepath.Join(outputDir, "page_ranges_shrink.pdf")
	if _, err := page.PDF(rangePDF, map[string]any{
		"pageRanges":  []string{"1-2"},
		"shrinkToFit": true,
	}); err != nil {
		return err
	}
	fmt.Printf("   ✓ 已保存: %s (%d bytes)\n", rangePDF, fileSize(rangePDF))

	fmt.Println("\n5. 获取 bytes / base64:")
	pdfBytes, err := page.PDF("", map[string]any{
		"pageRanges": []string{"1"},
	})
	if err != nil {
		return err
	}
	pdfBytesNoShrink, err := page.PDF("", map[string]any{
		"shrinkToFit": false,
	})
	if err != nil {
		return err
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytesNoShrink)
	fmt.Printf("   bytes长度: %d\n", len(pdfBytes))
	fmt.Printf("   base64长度: %d\n", len(pdfBase64))
	if len(pdfBytes) <= 100 || len(pdfBase64) <= 100 {
		return fmt.Errorf("PDF 输出长度异常")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有PDF打印测试通过！")
	fmt.Printf("PDF保存在: %s\n", outputDir)
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
