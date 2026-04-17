package main

import (
	"fmt"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
)

func main() {
	opt := ruyipage.NewFirefoxOptions()
	opt.Headless(false)

	opt.CloseBrowserOnExitEnabled(true)
	page, err := ruyipage.NewFirefoxPage(opt)
	if err != nil {
		panic(err)
	}
	defer page.Quit(0, false)

	if err := page.Get("https://example.com"); err != nil {
		panic(err)
	}
	page.Wait().Sleep(time.Second * 100000)

	title, _ := page.Title()
	fmt.Println(title)
}
