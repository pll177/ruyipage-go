package main

import (
	"fmt"
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
	fmt.Println("测试 25: Browser 用户上下文与窗口管理")
	fmt.Println(strings.Repeat("=", 70))

	page, err := ruyipage.NewFirefoxPage(exampleutil.FixedVisibleOptions())
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	results := make([]exampleutil.CheckRow, 0, 8)
	userContextID := ""
	newTabContext := ""
	defer func() {
		if newTabContext != "" {
			_ = page.Contexts().Close(newTabContext, false)
		}
		if userContextID != "" {
			_ = page.Contexts().RemoveUserContext(userContextID)
		}
	}()

	if err := page.Get("https://www.example.com"); err != nil {
		return err
	}

	before, err := page.Contexts().GetUserContexts()
	if err != nil {
		return err
	}
	beforeIDs := userContextIDs(before)
	exampleutil.AddCheck(&results, "browser.getUserContexts before", "成功", fmt.Sprintf("创建前数量: %d", len(before)))

	userContextID, err = page.Contexts().CreateUserContext()
	if err != nil {
		return err
	}
	if userContextID != "" {
		exampleutil.AddCheck(&results, "browser.createUserContext", "成功", "userContext="+userContextID)
	} else {
		exampleutil.AddCheck(&results, "browser.createUserContext", "失败", "未返回 userContext ID")
	}

	afterCreate, err := page.Contexts().GetUserContexts()
	if err != nil {
		return err
	}
	afterCreateIDs := userContextIDs(afterCreate)
	if containsString(afterCreateIDs, userContextID) && len(afterCreate) == len(before)+1 {
		exampleutil.AddCheck(&results, "browser.getUserContexts after create", "成功", fmt.Sprintf("数量变为 %d", len(afterCreate)))
	} else {
		exampleutil.AddCheck(&results, "browser.getUserContexts after create", "失败", "创建后数量或 ID 校验不通过")
	}

	if userContextID != "" {
		newTabContext, err = page.Contexts().CreateTab(false, userContextID, "")
		if err != nil {
			return err
		}
		if newTabContext != "" {
			exampleutil.AddCheck(&results, "browsingContext.create tab in user context", "成功", "context="+newTabContext)
		} else {
			exampleutil.AddCheck(&results, "browsingContext.create tab in user context", "失败", "未返回新 tab context")
		}
	}

	if newTabContext != "" {
		if err := page.Contexts().Close(newTabContext, false); err != nil {
			exampleutil.AddCheck(&results, "browsingContext.close new tab", "跳过", err.Error())
		} else {
			exampleutil.AddCheck(&results, "browsingContext.close new tab", "成功", "新 tab 已关闭")
			newTabContext = ""
		}
	}

	if userContextID != "" {
		if err := page.Contexts().RemoveUserContext(userContextID); err != nil {
			return err
		}
		exampleutil.AddCheck(&results, "browser.removeUserContext", "成功", "已删除 "+userContextID)

		afterRemove, err := page.Contexts().GetUserContexts()
		if err != nil {
			return err
		}
		afterRemoveIDs := userContextIDs(afterRemove)
		if !containsString(afterRemoveIDs, userContextID) && len(afterRemove) == len(beforeIDs) {
			exampleutil.AddCheck(&results, "browser.getUserContexts after remove", "成功", fmt.Sprintf("数量回到 %d", len(afterRemove)))
		} else {
			exampleutil.AddCheck(&results, "browser.getUserContexts after remove", "失败", "删除后数量或 ID 校验不通过")
		}
		userContextID = ""
	}

	windows, err := page.Contexts().GetClientWindows()
	if err != nil {
		return err
	}
	if len(windows) == 0 {
		exampleutil.AddCheck(&results, "browser.getClientWindows", "失败", "未返回任何 client window")
		exampleutil.PrintChecks(results)
		return nil
	}
	exampleutil.AddCheck(&results, "browser.getClientWindows", "成功", fmt.Sprintf("窗口数量: %d", len(windows)))

	windowID := fmt.Sprint(windows[0]["clientWindow"])
	stateNotes := make([]string, 0, 5)
	stateFailed := false
	for _, state := range []string{"minimized", "normal", "maximized", "fullscreen", "normal"} {
		if err := page.Contexts().SetWindowState(windowID, state, nil, nil, nil, nil); err != nil {
			stateNotes = append(stateNotes, fmt.Sprintf("%s->error:%v", state, err))
			stateFailed = true
			continue
		}
		time.Sleep(400 * time.Millisecond)
		currentWindows, currentErr := page.Contexts().GetClientWindows()
		if currentErr != nil || len(currentWindows) == 0 {
			stateNotes = append(stateNotes, fmt.Sprintf("%s->unreadable", state))
			stateFailed = true
			continue
		}
		currentState := fmt.Sprint(currentWindows[0]["state"])
		stateNotes = append(stateNotes, fmt.Sprintf("%s->%s", state, currentState))
		if currentState != state {
			stateFailed = true
		}
	}
	if stateFailed {
		exampleutil.AddCheck(&results, "browser.setClientWindowState", "失败", strings.Join(stateNotes, ", "))
	} else {
		exampleutil.AddCheck(&results, "browser.setClientWindowState", "成功", strings.Join(stateNotes, ", "))
	}

	exampleutil.PrintChecks(results)
	return nil
}

func userContextIDs(rows []map[string]any) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, fmt.Sprint(row["userContext"]))
	}
	return ids
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
