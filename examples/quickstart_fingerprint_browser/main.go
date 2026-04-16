package main

import (
	"fmt"
	"os"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const defaultFPFile = `C:\Users\pll177\Desktop\core\profile1.txt`

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fpfile := exampleutil.ResolveEnvPath("RUYIPAGE_EXAMPLE_FPFILE", defaultFPFile)
	if _, err := os.Stat(fpfile); err != nil {
		return fmt.Errorf("请准备指纹浏览器 fpfile，并通过 RUYIPAGE_EXAMPLE_FPFILE 指定；当前路径=%s: %w", fpfile, err)
	}

	opts := exampleutil.FixedVisibleOptions().WithFPFile(fpfile)
	page, err := ruyipage.NewFirefoxPage(opts)
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	if err := page.Get("https://www.browserscan.net/zh"); err != nil {
		return err
	}

	if err := page.Emulation().SetGeolocation(39.9042, 116.4074, 100); err != nil {
		return err
	}
	if err := page.Emulation().SetTimezone("Asia/Tokyo"); err != nil {
		return err
	}
	if err := page.Emulation().SetLocale([]string{"ja-JP", "ja"}); err != nil {
		return err
	}
	if err := page.Network().SetExtraHeaders(map[string]string{
		"Accept-Language": "ja-JP,ja;q=0.9",
	}); err != nil {
		return err
	}

	dpr := 2.0
	if err := page.Emulation().SetScreenSize(1366, 768, &dpr); err != nil {
		return err
	}
	if err := page.Refresh(); err != nil {
		return err
	}

	screenWidth, err := page.RunJS(`return screen.width`)
	if err != nil {
		return err
	}
	screenHeight, err := page.RunJS(`return screen.height`)
	if err != nil {
		return err
	}
	fmt.Printf("Firefox: %s\n", exampleutil.FixedFirefoxPath)
	fmt.Printf("fpfile: %s\n", fpfile)
	fmt.Printf("屏幕设置覆盖,当前=%vx%v\n", screenWidth, screenHeight)
	return nil
}
