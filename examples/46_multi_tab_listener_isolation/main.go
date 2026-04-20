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
	fmt.Println("测试 46: Multi-tab Listener Isolation")
	fmt.Println(strings.Repeat("=", 70))

	server := testserver.New("127.0.0.1", exampleutil.ServerPort(9660))
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
		_ = page.Quit(0, false)
	}()

	tab1, err := page.NewTab("about:blank", false)
	if err != nil {
		return err
	}
	tab2, err := page.NewTab("about:blank", false)
	if err != nil {
		return err
	}

	defer tab1.Listen().Stop()
	defer tab2.Listen().Stop()

	baseURL := strings.TrimRight(server.GetURL(""), "/")
	url1 := baseURL + "/api/data?tab=one"
	url2 := baseURL + "/api/data?tab=two"
	results := make([]exampleutil.CheckRow, 0, 8)

	if err := tab1.Listen().Start("tab=one", false, "GET"); err != nil {
		return err
	}
	if err := tab2.Listen().Start("tab=two", false, "GET"); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "listener start", "成功", "tab-1 与 tab-2 已分别启动监听")

	if _, err := tab1.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, url1); err != nil {
		return err
	}
	packet1 := tab1.Listen().Wait(8 * time.Second)
	exampleutil.AddCheck(&results, "tab-1 initial listen", statusOf(packet1 != nil), packetNote(packet1, "首次监听 tab-1"))

	if err := tab1.Close(false); err != nil {
		return err
	}
	exampleutil.AddCheck(&results, "close tab-1", "成功", "已关闭第一个 tab，并保留 tab-2")

	if _, err := tab2.RunJS(`function(url){
		fetch(url).catch(function(){ return null; });
		return true;
	}`, url2); err != nil {
		return err
	}
	packet2 := tab2.Listen().Wait(8 * time.Second)
	exampleutil.AddCheck(&results, "tab-2 listen after tab-1 close", statusOf(packet2 != nil), packetNote(packet2, "关闭 tab-1 后 tab-2 继续监听"))

	if packet2 == nil {
		exampleutil.PrintChecks(results)
		return fmt.Errorf("关闭 tab-1 后，tab-2 未继续收到监听数据包，bug 仍存在")
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

func packetNote(packet *ruyipage.DataPacket, fallback string) string {
	if packet == nil {
		return fallback
	}
	return fmt.Sprintf("status=%d url=%s", packet.Status, packet.URL)
}
