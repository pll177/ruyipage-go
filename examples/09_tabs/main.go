package main

import (
	"fmt"
	"strings"
	"time"

	ruyipage "ruyipage-go"
	"ruyipage-go/examples/internal/exampleutil"
)

func main() {
	exampleutil.RunMain(run)
}

func run() error {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("测试9: 标签页管理")
	fmt.Println(strings.Repeat("=", 60))

	page, err := ruyipage.NewFirefoxPage(exampleutil.VisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		page.Wait().Sleep(2 * time.Second)
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

	fmt.Printf("\n1. 获取当前标签页数量:\n")
	fmt.Printf("   标签页数量: %d\n", page.TabsCount())

	tab2URL := "data:text/html,<html><head><title>Example Tab</title></head><body>tab2</body></html>"
	tab3URL := "data:text/html,<html><head><title>Wikipedia Tab</title></head><body>tab3</body></html>"

	fmt.Printf("\n2. 新建标签页:\n")
	tab2, err := page.NewTab(tab2URL, false)
	if err != nil {
		return err
	}
	page.Wait().Sleep(2 * time.Second)
	if tab2 == nil {
		return fmt.Errorf("新建标签页失败")
	}
	tab2Title, err := tab2.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ 新标签页已创建\n")
	fmt.Printf("   新标签页标题: %s\n", tab2Title)
	fmt.Printf("   当前标签页数量: %d\n", page.TabsCount())

	fmt.Printf("\n3. 再新建一个标签页:\n")
	tab3, err := page.NewTab(tab3URL, false)
	if err != nil {
		return err
	}
	page.Wait().Sleep(2 * time.Second)
	if tab3 == nil {
		return fmt.Errorf("创建第三个标签页失败")
	}
	tab3Title, err := tab3.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ 第三个标签页已创建\n")
	fmt.Printf("   标签页标题: %s\n", tab3Title)
	fmt.Printf("   当前标签页数量: %d\n", page.TabsCount())

	fmt.Printf("\n4. 获取所有标签页ID:\n")
	tabIDs := page.TabIDs()
	fmt.Printf("   标签页ID列表: %v\n", tabIDs)
	fmt.Printf("   当前页面 tab_id: %s\n", page.ContextID())

	fmt.Printf("\n5. 通过序号获取标签页:\n")
	firstTab, err := page.GetTab(1, "", "")
	if err != nil {
		return err
	}
	secondTab, err := page.GetTab(2, "", "")
	if err != nil {
		return err
	}
	if firstTab == nil || secondTab == nil {
		return fmt.Errorf("通过序号获取标签页失败")
	}
	firstTitle, err := firstTab.Title()
	if err != nil {
		return err
	}
	secondTitle, err := secondTab.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   第1个标签页标题: %s\n", firstTitle)
	fmt.Printf("   第2个标签页标题: %s\n", secondTitle)

	fmt.Printf("\n6. 切换到第一个标签页:\n")
	if err := firstTab.Get(testURL); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	firstTitle, err = firstTab.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   当前标题: %s\n", firstTitle)

	fmt.Printf("\n7. 获取最新的标签页:\n")
	latest := page.LatestTab()
	if latest == nil {
		return fmt.Errorf("未获取到最新标签页")
	}
	latestTitle, err := latest.Title()
	if err != nil {
		return err
	}
	fmt.Printf("   最新标签页标题: %s\n", latestTitle)

	fmt.Printf("\n8. 通过标题查找标签页:\n")
	exampleTab, err := page.GetTab(nil, "Example", "")
	if err != nil {
		return err
	}
	if exampleTab != nil {
		exampleTitle, titleErr := exampleTab.Title()
		if titleErr != nil {
			return titleErr
		}
		fmt.Printf("   找到标签页: %s\n", exampleTitle)
	} else {
		fmt.Printf("   未找到包含'Example'的标签页\n")
	}

	fmt.Printf("\n9. 通过URL查找标签页:\n")
	wikiTab, err := page.GetTab(nil, "Wikipedia", "")
	if err != nil {
		return err
	}
	if wikiTab != nil {
		wikiTitle, titleErr := wikiTab.Title()
		if titleErr != nil {
			return titleErr
		}
		fmt.Printf("   找到标签页: %s\n", wikiTitle)
	} else {
		fmt.Printf("   未找到包含'wikipedia'的标签页\n")
	}

	fmt.Printf("\n10. 获取所有标签页:\n")
	allTabs, err := page.GetTabs("", "")
	if err != nil {
		return err
	}
	fmt.Printf("   共有 %d 个标签页\n", len(allTabs))
	for index, tab := range allTabs {
		title, titleErr := tab.Title()
		if titleErr != nil {
			return titleErr
		}
		fmt.Printf("   标签页%d: %s\n", index+1, trimPreview(title, 50))
	}

	fmt.Printf("\n10.1 获取客户端窗口信息:\n")
	windowHandles := page.Browser().WindowHandles()
	fmt.Printf("   clientWindows 数量: %d\n", len(windowHandles))
	if len(windowHandles) > 0 {
		fmt.Printf("   首个窗口状态: %v\n", windowHandles[0]["state"])
	}

	fmt.Printf("\n10.2 获取 browsing context 树:\n")
	tree, err := page.Contexts().GetTree(nil, "")
	if err != nil {
		return err
	}
	fmt.Printf("   顶层 context 数量: %d\n", len(tree.Contexts))
	if len(tree.Contexts) > 0 {
		fmt.Printf("   第一个 context: %s\n", tree.Contexts[0].Context)
	}

	fmt.Printf("\n11. 关闭第二个标签页:\n")
	if err := secondTab.Close(false); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 标签页已关闭\n")
	fmt.Printf("   剩余标签页数量: %d\n", page.TabsCount())

	fmt.Printf("\n12. 关闭其他标签页:\n")
	if err := page.CloseOtherTabs(firstTab.ContextID()); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 其他标签页已关闭\n")
	fmt.Printf("   剩余标签页数量: %d\n", page.TabsCount())

	fmt.Printf("\n13. 新建后台标签页:\n")
	bgTab, err := page.NewTab(tab2URL, true)
	if err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	if bgTab == nil {
		return fmt.Errorf("创建后台标签页失败")
	}
	fmt.Printf("   ✓ 后台标签页已创建\n")
	fmt.Printf("   当前标签页数量: %d\n", page.TabsCount())

	fmt.Printf("\n14. 激活后台标签页:\n")
	if _, err := bgTab.Activate(); err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   ✓ 标签页已激活\n")

	fmt.Printf("\n15. 通过高层context API创建后台标签页:\n")
	bidiTabID, err := page.Contexts().CreateTab(true, "", page.ContextID())
	if err != nil {
		return err
	}
	page.Wait().Sleep(time.Second)
	fmt.Printf("   高层新标签页ID: %s\n", bidiTabID)
	fmt.Printf("   当前标签页数量: %d\n", page.TabsCount())
	if bidiTabID != "" {
		bidiTab, tabErr := page.GetTab(bidiTabID, "", "")
		if tabErr != nil {
			return tabErr
		}
		if bidiTab != nil {
			if err := bidiTab.Close(false); err != nil {
				return err
			}
			page.Wait().Sleep(time.Second)
			fmt.Printf("   ✓ 纯BiDi标签页已关闭\n")
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ 所有标签页管理测试通过！")
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func trimPreview(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
