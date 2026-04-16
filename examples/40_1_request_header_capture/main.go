package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
	"github.com/pll177/ruyipage-go/examples/internal/exampleutil"
)

const requestCaptureURL = "https://www.posindonesia.co.id/en/tracking/LZ027513746CN"

type capturedPost struct {
	URL     string
	Headers map[string]string
	Body    string
}

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("Example 40_1: Request Header Capture")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Intercept().Stop()
		_ = page.Quit(0, false)
	}()

	capturedPosts := make([]capturedPost, 0, 8)
	if err := page.Get("about:blank"); err != nil {
		return err
	}

	if _, err := page.Intercept().StartRequests(func(req *ruyipage.InterceptedRequest) {
		if req.Method == "POST" {
			body := req.Body()
			headers := cloneHeaders(req.Headers)
			capturedPosts = append(capturedPosts, capturedPost{
				URL:     req.URL,
				Headers: headers,
				Body:    body,
			})
			fmt.Println("\n[POST]")
			fmt.Printf("url: %s\n", req.URL)
			fmt.Printf("headers: %v\n", headers)
			fmt.Printf("body: %s\n", truncate(body, 500))
		}
		_ = req.ContinueRequest("", "", nil, nil)
	}, nil); err != nil {
		return err
	}

	if err := page.Navigate(requestCaptureURL, "none"); err != nil {
		return err
	}
	page.Wait().Sleep(3 * time.Second)

	passed := page.HandleCloudflareChallenge(20*time.Second, 2*time.Second)
	fmt.Printf("cloudflare passed: %v\n", passed)
	page.Wait().Sleep(20 * time.Second)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("captured POST count: %d\n", len(capturedPosts))
	for index, item := range capturedPosts {
		if index >= 10 {
			break
		}
		fmt.Printf("%d. %s\n", index+1, item.URL)
		fmt.Printf("   content-type: %s\n", item.Headers["Content-Type"])
		fmt.Printf("   origin: %s\n", item.Headers["Origin"])
		fmt.Printf("   body preview: %s\n", truncate(item.Body, 200))
	}
	fmt.Println(strings.Repeat("=", 70))
	return nil
}

func cloneHeaders(headers map[string]string) map[string]string {
	result := make(map[string]string, len(headers))
	for key, value := range headers {
		result[key] = value
	}
	return result
}

func truncate(value string, size int) string {
	if len(value) <= size {
		return value
	}
	return value[:size]
}
