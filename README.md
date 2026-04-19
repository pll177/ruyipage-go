# ruyipage-go

> 基于 **Firefox + WebDriver BiDi** 的 Go 自动化库，适合页面分析、数据采集、网络监听、浏览器接管、指纹浏览器协同与高层自动化封装。
>
> 这份 README 参考 Python 版 `ruyiPage` README 的组织方式重写，但所有代码示例、方法名和目录路径都以 **当前 Go 仓库实际 API** 为准。

## 仓库关系

- Go 实现：<https://github.com/pll177/ruyipage-go>
- Python 基线：<https://github.com/LoseNine/ruyipage>

`ruyipage-go` 延续了 `ruyiPage` 的 Firefox + BiDi 路线，并在 Go 侧暴露了一组高层对象：

- `FirefoxPage`
- `FirefoxTab`
- `FirefoxFrame`
- `Firefox`
- `FirefoxOptions`
- `FirefoxElement`
- `NoneElement`
- `StaticElement`

---

## v1 最新修复：`AutoPortEnabled(true)` 并发启动端口冲突

已修复同一 Go 进程内并发启动多个 Firefox 实例时，`AutoPortEnabled(true)` 看起来已经开启，但最终仍可能只启动出一个浏览器的问题。

这次问题的真实原因有两层：

- 底层 `internal/browser` 里，自动端口以前只是“启动前扫描空闲端口”，并发时存在撞端口窗口
- 更关键的是顶层 `firefox_page.go` 在真正分配出自动端口之前，就先用默认地址（通常是 `127.0.0.1:9222`）做了 page 单例缓存，导致多个 `NewFirefoxPage(opts)` 调用直接复用成同一个页面对象

所以实际现象会变成：

- 即使你已经设置了 `AutoPortEnabled(true)`
- 如果没有显式 `WithAddress(...)`
- 多次并发创建时仍可能被外层 page cache 合并，最后看起来只用了同一个浏览器

当前行为：

- 同一 Go 进程内并发启动时，底层自动端口分配会先做进程内唯一预留
- 启动失败或连接失败时，会自动换一个新端口继续重试
- 浏览器退出后，端口会释放，后续实例可以再次复用
- 顶层 `FirefoxPage` 现在不会再在自动端口真正解析前，错误地按默认地址提前复用 page cache
- 如果你显式调用 `WithAddress("127.0.0.1:xxxx")`，则仍以显式地址为准，不走自动端口分配

这次修复后，`AutoPortEnabled(true)` 才真正具备了你预期的效果：**不手动指定 `WithAddress(...)` 时，也能在同进程并发场景下启动多个独立浏览器实例，并在退出后复用端口。**

---

## 更新到最新版本

当前推荐安装版本：`v1.0.1`

如果你是新项目，请直接显式安装当前版本：

```bash
go get github.com/pll177/ruyipage-go@v1.0.1
go mod tidy
```

如果你项目里已经安装过旧版本，请在你的项目目录执行：

```bash
go get github.com/pll177/ruyipage-go@v1.0.1
go mod tidy
```

如果本机 Go proxy / module cache 还没刷新，先清缓存再更新：

```bash
go clean -modcache
go get github.com/pll177/ruyipage-go@v1.0.1
go mod tidy
```

说明：

- 目前不建议依赖 `@latest`，因为 Go proxy 刷新 tag 可能有延迟
- 新安装、老项目升级，都建议显式写 `@当前版本`
- 后续每次发布都会递增小版本号，例如 `v1.0.2`、`v1.0.3`
- 你只需要把 README 里的版本号替换成最新发布版本即可
- 不需要再手动写 commit hash，也不需要再临时写 `@main`

---

## v1 最新更新：`WithAutoFPFile()` 先看这里

`WithAutoFPFile()` 现在是**无参数** API，用来根据你**前面已经设置好的配置**自动生成临时 fpfile。

当前规则：

- `WithAutoFPFile()` 会读取调用前已经设置好的状态
- 代理是可选的：传了 `WithProxy(...)` 就走代理，不传就走本机网络
- 窗口大小优先读取前面设置的 `WithWindowSize(...)`
- 如果前面没设置窗口大小，则默认使用 `1280 x 800`
- 一旦调用了 `WithAutoFPFile()`，后面就不要再继续调用配置类方法；否则会在 `Validate()` / 启动阶段报错
- 框架内部自动补的临时 profile、自动端口不算用户配置，不会误触发这条限制

### 示例 1：本地直连，不走代理

```go
opts := ruyipage.NewFirefoxOptions().
	Headless(false).
	WithBrowserPath(`C:\Program Files\Mozilla Firefox\firefox.exe`).
	WithWindowSize(1280, 800).
	CloseBrowserOnExitEnabled(true)

if _, err := opts.WithAutoFPFile(); err != nil {
	panic(err)
}

page, err := ruyipage.NewFirefoxPage(opts)
if err != nil {
	panic(err)
}
defer func() {
	_ = page.Quit(0, false)
}()
```

### 示例 2：先设代理，再生成自动 fpfile

```go
opts := ruyipage.NewFirefoxOptions().
	Headless(false).
	WithBrowserPath(`C:\Program Files\Mozilla Firefox\firefox.exe`).
	WithProxy("http://user:pass@proxy.example.com:7878").
	WithWindowSize(1440, 900)

if _, err := opts.WithAutoFPFile(); err != nil {
	panic(err)
}

page, err := ruyipage.NewFirefoxPage(opts)
if err != nil {
	panic(err)
}
defer func() {
	_ = page.Quit(0, false)
}()
```

### 示例 3：错误写法（不要把配置放到后面）

```go
opts := ruyipage.NewFirefoxOptions().
	WithBrowserPath(`C:\Program Files\Mozilla Firefox\firefox.exe`)

if _, err := opts.WithAutoFPFile(); err != nil {
	panic(err)
}

// 这里会导致后续校验/启动报错
opts.WithWindowSize(1600, 900)
opts.WithProxy("http://proxy.example.com:7878")
```

如果你已经有现成 fpfile，仍然可以继续使用 `WithFPFile(...)` 手工传入。

---

## 安装与使用

### 安装

```bash
go get github.com/pll177/ruyipage-go@v1.0.1
go mod tidy
```

导入方式：

```go
import ruyipage "github.com/pll177/ruyipage-go"
```

老项目更新：

```bash
go get github.com/pll177/ruyipage-go@v1.0.1
go mod tidy
```

安装后可确认模块版本：

```bash
go list -m github.com/pll177/ruyipage-go
```

### 环境要求

- Windows
- Go 1.26+
- Firefox 或兼容 Firefox 内核浏览器

示例默认使用的 Firefox 路径：

```powershell
C:\Program Files\Mozilla Firefox\firefox.exe
```

如果你准备直接运行 `examples/` 里的示例，建议优先通过环境变量覆盖本机环境差异：

| 环境变量 | 用途 |
| --- | --- |
| `RUYIPAGE_EXAMPLE_FIREFOX_PATH` | 指定示例所用 Firefox 路径 |
| `RUYIPAGE_EXAMPLE_ATTACH_COMMAND` | 覆盖“接管已打开浏览器”示例中的手工启动命令 |
| `RUYIPAGE_EXAMPLE_USER_DIR` | 为 userdir / 私密模式 / XPath 示例指定独立目录 |
| `RUYIPAGE_EXAMPLE_FPFILE` | 指纹浏览器 quickstart 示例使用的 fpfile |
| `RUYIPAGE_EXAMPLE_PROXY_HOST` | 代理示例主机 |
| `RUYIPAGE_EXAMPLE_PROXY_PORT` | 代理示例端口 |
| `RUYIPAGE_EXAMPLE_PROXY_USERNAME` | 代理认证用户名 |
| `RUYIPAGE_EXAMPLE_PROXY_PASSWORD` | 代理认证密码 |

### 最简单启动

```go
package main

import (
	"fmt"
	"time"

	ruyipage "github.com/pll177/ruyipage-go"
)

func main() {
	page, err := ruyipage.Launch()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	if err := page.Get("https://example.com"); err != nil {
		panic(err)
	}
	page.Wait().Sleep(time.Second)

	title, err := page.Title()
	if err != nil {
		panic(err)
	}
	fmt.Println(title)
}
```

### JS 事件 `isTrusted` 对比能力

`ruyipage-go` 不只支持原生点击、输入、拖拽、滚轮、键盘等高 `isTrusted` 动作，也支持在多种 JS 事件构造里附加 `ruyi: true`，让事件行为更贴近真实交互。

例如：

```javascript
new Event("change", { bubbles: true, ruyi: true })
new InputEvent("input", { bubbles: true, data: "A", inputType: "insertText", ruyi: true })
new MouseEvent("click", { bubbles: true, clientX: 12, clientY: 24, ruyi: true })
new KeyboardEvent("keydown", { bubbles: true, key: "Enter", code: "Enter", ruyi: true })
```

可直接运行综合示例：

```bash
go run ./examples/45_js_setter_untrusted_input
```

该示例会对比普通 JS 事件与 `ruyi: true` 事件在 `isTrusted` 上的差异，覆盖：

- `Event`
- `InputEvent`
- `KeyboardEvent`
- `MouseEvent`
- `FocusEvent`
- `CustomEvent`
- `PointerEvent`
- `WheelEvent`

### 指定 Firefox 路径和 userdir

```go
package main

import (
	"fmt"

	ruyipage "github.com/pll177/ruyipage-go"
)

func main() {
	opts := ruyipage.NewFirefoxOptions().
		WithBrowserPath(`D:\Firefox\firefox.exe`).
		WithUserDir(`D:\ruyipage_userdir`)

	page, err := ruyipage.NewFirefoxPage(opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	if err := page.Get("https://www.example.com"); err != nil {
		panic(err)
	}

	title, _ := page.Title()
	fmt.Println(title)
}
```

其中：

- `browser_path` 对应 Go 的 `WithBrowserPath(...)`
- `user_dir` 对应 Go 的 `WithUserDir(...)`
- 不传 `user_dir` 时，通常会走临时 profile，更适合一次性测试
- 需要复用登录态、Cookie、扩展、首选项时，建议显式指定 `user_dir`

### 更适合新手的 `Launch`

`Launch()` 是 Go 版对齐 Python `launch(...)` 的快捷入口：

- `Launch()`：使用默认 quick start 预设
- `Launch(FirefoxQuickStartOptions{...})`：带参数启动

```go
package main

import (
	"fmt"

	ruyipage "github.com/pll177/ruyipage-go"
)

func main() {
	page, err := ruyipage.Launch(ruyipage.FirefoxQuickStartOptions{
		BrowserPath:  `D:\Firefox\firefox.exe`,
		UserDir:      `D:\ruyipage_userdir`,
		Private:      true,
		Headless:     false,
		XPathPicker:  false,
		ActionVisual: false,
		Port:         9333,
		WindowWidth:  1440,
		WindowHeight: 960,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = page.Quit(0, false)
	}()

	_ = page.Get("https://www.example.com")
	title, _ := page.Title()
	fmt.Println(title)
}
```

`FirefoxQuickStartOptions` 常用字段：

| 字段 | 说明 |
| --- | --- |
| `BrowserPath` | Firefox 可执行文件路径 |
| `UserDir` | profile / userdir |
| `Private` | 是否私密模式 |
| `Headless` | 是否无头 |
| `XPathPicker` | 是否启用 XPath Picker |
| `ActionVisual` | 是否启用鼠标行为可视化调试 |
| `Port` | 调试端口 |
| `WindowWidth` / `WindowHeight` | 启动窗口大小 |
| `TimeoutBase` / `TimeoutPageLoad` / `TimeoutScript` | 三类超时，单位秒 |

### 开启隐私模式

```go
package main

import ruyipage "github.com/pll177/ruyipage-go"

func main() {
	// 方式一：配置对象
	opts := ruyipage.NewFirefoxOptions().
		PrivateMode(true)
	page1, err := ruyipage.NewFirefoxPage(opts)
	if err != nil {
		panic(err)
	}
	_ = page1.Quit(0, false)

	// 方式二：Launch 快捷入口
	page2, err := ruyipage.Launch(ruyipage.FirefoxQuickStartOptions{
		Private: true,
	})
	if err != nil {
		panic(err)
	}
	_ = page2.Quit(0, false)
}
```

说明：

- `PrivateMode(true)` 会给 Firefox 添加私密浏览参数
- 私密模式与“临时 profile”不是同一个概念
- 如果只是一次性会话，不复用历史数据，也可以只是不传 `user_dir`
- 完整示例见：`examples/quickstart_private_mode` 和 `examples/41_private_mode_userdir`

### 启用 XPath Picker

```go
opts := ruyipage.NewFirefoxOptions().
	WithBrowserPath(`D:\Firefox\firefox.exe`).
	XPathPickerEnabled(true).
	WithWindowSize(1600, 1100)

page, err := ruyipage.NewFirefoxPage(opts)
if err != nil {
	panic(err)
}
defer func() {
	_ = page.Quit(0, false)
}()

_ = page.Get("https://example.com")
```

启用后，页面右下角会出现半透明浮窗，支持：

- 显示元素名称、文本、绝对 XPath、相对 XPath
- 显示元素中心坐标
- 一键复制 XPath
- 自动生成对应的 Go 访问代码
- 元素组捕获、组策略切换、组 XPath / 组 CSS 复制
- iframe、嵌套 iframe、open shadow root、closed shadow root 场景

推荐直接运行：

```bash
go run ./examples/42_xpath_picker_complex_showcase
go run ./examples/43_zhipin_xpath_picker
go run ./examples/44_shopee_userdir_xpath
```

### 鼠标行为可视化调试

`ActionVisualEnabled(true)` 会把动作链轨迹、点击圈、目标高亮显示出来，适合排查：

- 鼠标轨迹是否跑偏
- 点击坐标是否命中目标
- `ClickSelf(true)` / `Input(..., byJS=true)` 是否作用在预期元素上

```go
opts := ruyipage.NewFirefoxOptions().
	WithBrowserPath(`D:\Firefox\firefox.exe`).
	ActionVisualEnabled(true).
	WithWindowSize(1400, 900)
```

推荐直接运行：

```bash
go run ./examples/42_2_action_visual_showcase
```

### 接管已打开的浏览器

先手工启动 Firefox，并打开调试端口：

```powershell
"C:\Program Files\Mozilla Firefox\firefox.exe" -remote-debugging-port 9222
```

固定地址接管：

```go
page, err := ruyipage.Attach("127.0.0.1:9222")
```

按端口范围自动接管：

```go
page, err := ruyipage.AutoAttachExistingBrowser(
	"",
	"127.0.0.1",
	9222,
	65535,
	200*time.Millisecond,
	64,
	1,
	true,
)
```

按进程特征自动接管：

```go
page, err := ruyipage.AutoAttachExistingBrowserByProcess(
	"",
	200*time.Millisecond,
	16,
	1,
	true,
)
```

推荐示例：

```bash
go run ./examples/39_attach_exist_browser
```

---

## 项目定位与技术路线

`ruyipage-go` 不是简单的底层 BiDi 命令转发器，而是一个面向 Go 业务项目的高层 Firefox 自动化封装。

技术路线：

- 基于 **Firefox + WebDriver BiDi**
- 保留 `Page / Tab / Frame / Element / Browser / Options` 这一层高层抽象
- 兼顾页面自动化、网络抓包、下载、事件、上下文、模拟能力、扩展管理
- 支持 **新启动浏览器** 和 **接管已打开浏览器**
- 支持与 Firefox 指纹浏览器一起使用

### 高风控场景推荐

如果你的目标是：

- 登录流
- 验证码前后页面
- 强交互页面
- 复杂 iframe / shadow DOM 页面
- 需要更真实交互轨迹的采集或分析流程

那这条 Firefox + BiDi 路线通常比“只会普通 DOM click/input”的轻量方案更合适。

---

## 能力总览

| 能力 | 主要入口 | 代表示例 |
| --- | --- | --- |
| 页面启动 / attach | `Launch()` / `NewFirefoxPage()` / `Attach()` | `00_quickstart`、`39_attach_exist_browser` |
| 页面导航 | `Get()` / `Refresh()` / `Back()` / `Forward()` | `01_basic_navigation` |
| 元素查找 / 交互 | `Ele()` / `Eles()` / `ClickSelf()` / `Input()` | `02_element_finding`、`03_element_interaction` |
| 等待机制 | `Wait()` / `WaitLoadComplete()` / `WaitURLContains()` | `04_wait_conditions` |
| 动作链 / 触摸 | `Actions()` / `Touch()` | `05_actions_chain`、`20_advanced_input` |
| 截图 / PDF | `Screenshot()` / `PDF()` | `06_screenshot`、`19_pdf_printing` |
| JavaScript / script | `RunJS()` / `RunJSExpr()` / `EvalHandle()` / `AddPreloadScript()` | `07_javascript`、`29_script_input_advanced`、`32_script_events` |
| Cookies / Storage | `Cookies()` / `SetCookies()` / `DeleteCookies()` / `LocalStorage()` / `SessionStorage()` | `08_cookies` |
| 标签页 / frame | `NewTab()` / `GetTab()` / `GetFrame()` | `09_tabs`、`13_iframe` |
| Shadow DOM | `ShadowRoot()` / `ClosedShadowRoot()` | `14_shadow_dom` |
| 提示框 | `WaitPrompt()` / `HandlePrompt()` | `17_user_prompts` |
| 网络监听 / 拦截 | `Listen()` / `Intercept()` / `Network()` | `11_network_intercept`、`18_advanced_network`、`28_network_data_collector` |
| 下载 | `Downloads()` | `23_download` |
| 导航 / 通用事件 | `Navigation()` / `Events()` | `24_navigation_events`、`30_browsing_context_events`、`31_network_events` |
| 浏览上下文 | `Contexts()` | `25_browser_user_context`、`26_browsing_context_advanced` |
| 模拟能力 | `Emulation()` | `21_emulation`、`21_mobile_google_emulation`、`27_emulation_advanced` |
| WebExtension | `Extensions()` | `22_web_extension` |
| XPath Picker | `XPathPickerEnabled(true)` | `42_xpath_picker_complex_showcase` |
| 行为可视化 | `ActionVisualEnabled(true)` | `42_2_action_visual_showcase` |

---

## 根目录快速开始示例

### 1. Bing 搜索示例

```bash
go run ./examples/quickstart_bing_search
```

### 2. Cloudflare / Copilot 示例

```bash
go run ./examples/quickstart_cloudfare
```

### 3. 指纹浏览器示例

先准备 fpfile，并通过环境变量指定：

```powershell
$env:RUYIPAGE_EXAMPLE_FPFILE='D:\path\to\profile.txt'
go run ./examples/quickstart_fingerprint_browser
```

### 4. 私密模式 + userdir 示例

```powershell
$env:RUYIPAGE_EXAMPLE_USER_DIR='D:\ruyipage_userdir'
go run ./examples/quickstart_private_mode
```

---

## 最常用 API 文档

## 1. 页面对象：`FirefoxPage`

### 创建页面

```go
// 方式一：新手友好入口
page, err := ruyipage.Launch()

// 方式二：完整配置
opts := ruyipage.NewFirefoxOptions().
	WithBrowserPath(`D:\Firefox\firefox.exe`).
	WithUserDir(`D:\ruyipage_userdir`)
page, err := ruyipage.NewFirefoxPage(opts)

// 方式三：接管已打开浏览器
page, err := ruyipage.Attach("127.0.0.1:9222")
```

### 常用属性

| API | 说明 |
| --- | --- |
| `page.URL()` | 当前页面 URL |
| `page.Title()` | 当前标题 |
| `page.HTML()` | 当前页面 HTML |
| `page.ContextID()` | 当前 browsing context id |
| `page.Browser()` | 获取关联的 `Firefox` 对象 |
| `page.TabIDs()` | 当前可见 tab id 列表 |
| `page.TabsCount()` | 标签页数量 |

### 常用导航

| API | 说明 |
| --- | --- |
| `page.Get(url)` | 跳转页面 |
| `page.Refresh()` | 刷新 |
| `page.Back()` / `page.Forward()` | 前进后退 |
| `page.WaitLoadComplete(timeout)` | 等待页面完成加载 |
| `page.WaitURLContains(fragment, timeout)` | 等待 URL 包含指定片段 |
| `page.WaitTitleContains(fragment, timeout)` | 等待标题包含指定片段 |
| `page.Wait()` | 获取高层等待器 |
| `page.Quit(timeout, force)` | 关闭页面 / 浏览器会话 |

### `page.Wait()` 常用入口

| API | 说明 |
| --- | --- |
| `page.Wait().Sleep(duration)` | 简单 sleep |
| `page.Wait().Ele(locator, timeout)` | 等待元素出现 |
| `page.Wait().EleDisplayed(locator, timeout)` | 等待元素显示 |
| `page.Wait().EleHidden(locator, timeout)` | 等待元素隐藏 |
| `page.Wait().EleDeleted(locator, timeout)` | 等待元素删除 |
| `page.Wait().DocLoaded(timeout)` | 等待 DOMContentLoaded / 文档就绪 |
| `page.Wait().LoadComplete(timeout)` | 等待页面完整加载 |
| `page.Wait().URLContains(fragment, timeout)` | 等待 URL 命中 |
| `page.Wait().TitleContains(fragment, timeout)` | 等待标题命中 |

### 标签页与 frame

| API | 说明 |
| --- | --- |
| `page.NewTab(url, background)` | 新建标签页 |
| `page.GetTab(idOrNum, title, url)` | 按 id / 序号 / 标题 / URL 获取 tab |
| `page.GetTabs(title, url)` | 获取多个标签页 |
| `page.LatestTab()` | 获取最新 tab |
| `page.CloseOtherTabs(tabOrIDs)` | 关闭其他标签页 |
| `tab.Activate()` | 激活指定 tab |
| `tab.Close(others)` | 关闭当前 tab |
| `page.GetFrame(locatorOrIndexOrContext)` | 获取 frame |
| `page.GetFrames()` | 获取全部 frame |

---

## 2. 元素查找：`Ele()` / `Eles()`

### `page.Ele(locator, index, timeout)`

最常见的定位写法就是字符串定位器：

```go
h1, _ := page.Ele("tag:h1", 1, 5*time.Second)
submit, _ := page.Ele("#submit", 1, 5*time.Second)
loginBtn, _ := page.Ele("text:登录", 1, 5*time.Second)
link, _ := page.Ele(`xpath://a[@id="test-link"]`, 1, 5*time.Second)
```

常见 locator 形式：

- CSS：`#kw`、`.item`、`[data-id="1"]`
- XPath：`xpath://button[@type="submit"]`
- 文本：`text:提交`
- 标签：`tag:h1`

### `page.Eles(locator, timeout)`

```go
items, err := page.Eles(".result-item", 5*time.Second)
if err != nil {
	panic(err)
}
for _, item := range items {
	text, _ := item.Text()
	fmt.Println(text)
}
```

### 在元素内部继续查找

```go
form, _ := page.Ele("#login-form", 1, 5*time.Second)
username, _ := form.Ele(`input[name="username"]`, 1, 5*time.Second)
password, _ := form.Ele(`input[type="password"]`, 1, 5*time.Second)
```

### 常用元素 API

| API | 说明 |
| --- | --- |
| `ele.Text()` | 读取文本 |
| `ele.HTML()` | 读取 outer HTML |
| `ele.Tag()` | 标签名 |
| `ele.Value()` | value 值 |
| `ele.Attr(name)` / `ele.Attrs()` | 属性读取 |
| `ele.Link()` / `ele.Src()` | 读取链接 / 资源地址 |
| `ele.IsDisplayed()` / `ele.IsEnabled()` | 状态判断 |
| `ele.Size()` / `ele.Location()` | 尺寸和坐标 |
| `ele.ClickSelf(byJS, timeout)` | 点击 |
| `ele.Input(text, clear, byJS)` | 输入 |
| `ele.Clear()` | 清空输入框 |
| `ele.Hover()` | hover |
| `ele.DragTo(target, duration)` | 拖拽 |
| `ele.ShadowRoot()` / `ele.ClosedShadowRoot()` | 获取 shadow root |
| `ele.Ele()` / `ele.Eles()` | 元素内继续查找 |

---

## 3. 动作链：`page.Actions()`

动作链适合做更真实的鼠标 / 键盘交互。

```go
err := page.Actions().
	MoveTo(map[string]int{"x": 300, "y": 220}, 0, 0, 120*time.Millisecond, nil).
	Click(nil, 1).
	Perform()
if err != nil {
	panic(err)
}
```

### 常见写法

| API | 说明 |
| --- | --- |
| `MoveTo(target, dx, dy, duration, origin)` | 鼠标移动到坐标或元素 |
| `HumanMove(target, origin)` | 拟人移动 |
| `Click(on, count)` | 左键点击 |
| `DBClick(on)` | 双击 |
| `RClick(on)` | 右键点击 |
| `Hold(on, button)` / `Release(on)` | 按下 / 释放 |
| `Scroll(dx, dy, origin, deltaOrigin)` | 滚轮滚动 |
| `Type(text, delay)` | 输入文本 |
| `KeyDown(key)` / `KeyUp(key)` | 键盘按下 / 抬起 |
| `Wait(duration)` | 在动作链中暂停 |
| `Perform()` | 执行动作链 |

如果你要做触摸输入，可继续看 `page.Touch()` 相关示例：

```bash
go run ./examples/20_advanced_input
```

---

## 4. Cookies 与 Storage

### 获取 / 设置 / 删除 Cookie

```go
cookies, _ := page.Cookies(true)

_ = page.SetCookies(map[string]any{
	"name":   "session_id",
	"value":  "abc123",
	"domain": "example.com",
	"path":   "/",
})

_ = page.DeleteCookies(map[string]any{
	"name": "session_id",
})
```

如果你偏好“setter 风格”：

```go
_ = page.CookiesSetter().Set(map[string]any{
	"name":   "demo_cookie",
	"value":  "demo_value",
	"domain": "example.com",
	"path":   "/",
})

_ = page.CookiesSetter().Remove("demo_cookie", "example.com")
```

### `localStorage` / `sessionStorage`

```go
_ = page.LocalStorage().Set("token", "abc")
token, _ := page.LocalStorage().Get("token")
items, _ := page.LocalStorage().Items()

_ = page.SessionStorage().Set("page", "home")
count, _ := page.SessionStorage().Len()

fmt.Println(token, items, count)
```

### 相关 API

| API | 说明 |
| --- | --- |
| `page.Cookies(allInfo)` | 获取 cookie 列表 |
| `page.SetCookie(cookie)` / `page.SetCookies(cookies)` | 设置 cookie |
| `page.DeleteCookies(filter)` | 删除 cookie |
| `page.CookiesSetter().Set(...)` | setter 风格设置 |
| `page.CookiesSetter().Remove(name, domain)` | 删除单个 cookie |
| `page.CookiesSetter().Clear()` | 清空 cookie |
| `page.LocalStorage()` / `page.SessionStorage()` | 存储管理器 |

完整示例：

```bash
go run ./examples/08_cookies
```

---

## 5. 下载

```go
downloadDir := `D:\ruyipage_downloads`

if err := page.Downloads().SetBehavior("allow", downloadDir, nil, nil); err != nil {
	panic(err)
}
if err := page.Downloads().Start(); err != nil {
	panic(err)
}

begin, end := page.Downloads().WaitChain("test.txt", 5*time.Second)
ok := page.Downloads().WaitFile(downloadDir+`\\test.txt`, 3*time.Second, 1)

fmt.Println(begin != nil, end != nil, ok)
```

### 典型下载流程

1. `SetBehavior("allow", path, nil, nil)`
2. `Start()` 开启下载事件监听
3. 触发下载动作
4. `WaitChain(filename, timeout)` 等待开始 / 结束事件
5. `WaitFile(path, timeout, minSize)` 确认文件落盘

完整示例：

```bash
go run ./examples/23_download
```

---

## 6. 导航事件

```go
if err := page.Navigation().Start([]string{
	"browsingContext.navigationStarted",
	"browsingContext.load",
}); err != nil {
	panic(err)
}
defer page.Navigation().Stop()

_ = page.Get("https://example.com")
event := page.Navigation().WaitForLoad(5 * time.Second)
fmt.Println(event != nil)
```

`page.Navigation()` 适合：

- 精确观察某次跳转的开始 / 完成
- 需要判断 URL 片段变化
- 需要把普通 `Get()` 和页面内跳转区分开来

完整示例：

```bash
go run ./examples/24_navigation_events
```

---

## 7. 通用事件监听

```go
if err := page.Events().Start([]string{"network.fetchError"}, []string{page.ContextID()}); err != nil {
	panic(err)
}
defer page.Events().Stop()

event := page.Events().Wait("network.fetchError", 5*time.Second)
fmt.Println(event != nil)
```

`page.Events()` 适合：

- 一次性观察某类 BiDi 事件
- 验证某个操作是否触发了脚本 / 网络 / 输入 / 日志事件
- 做调试型、审计型、回归型验证

推荐示例：

```bash
go run ./examples/30_browsing_context_events
go run ./examples/31_network_events
go run ./examples/32_script_events
go run ./examples/33_log_input_events
```

---

## 8. 网络能力

### 请求监听

```go
if err := page.Listen().Start("/api/data", false, "GET"); err != nil {
	panic(err)
}
defer page.Listen().Stop()

packet := page.Listen().Wait(5 * time.Second)
if packet != nil {
	fmt.Println(packet.URL, packet.Status)
}
```

### 请求 / 响应拦截

```go
_, err := page.Intercept().StartRequests(func(req *ruyipage.InterceptedRequest) {
	if req.Method == "GET" {
		_ = req.ContinueRequest("", "", nil, nil)
		return
	}
	_ = req.Fail()
}, nil)
if err != nil {
	panic(err)
}
defer page.Intercept().Stop()
```

### `page.Network()` 高层能力

| API | 说明 |
| --- | --- |
| `page.Network().SetExtraHeaders(headers)` | 设置额外请求头 |
| `page.Network().ClearExtraHeaders()` | 清空额外请求头 |
| `page.Network().SetCacheBehavior(behavior)` | 设置缓存行为 |
| `page.Network().AddDataCollector(events, dataTypes, maxSize)` | 注册 data collector |
| `collector.Get(requestID, dataType)` | 读取网络数据 |
| `collector.Disown(requestID, dataType)` | 释放数据 |
| `collector.Remove()` | 移除 collector |

完整示例：

```bash
go run ./examples/11_network_intercept
go run ./examples/18_advanced_network
go run ./examples/28_network_data_collector
go run ./examples/40_scraper_packet_capture
go run ./examples/40_1_request_header_capture
```

---

## 9. 浏览上下文

```go
tree, err := page.Contexts().GetTree(nil, "")
if err != nil {
	panic(err)
}
fmt.Println(len(tree.Contexts))

userContextID, _ := page.Contexts().CreateUserContext()
fmt.Println(userContextID)
```

`page.Contexts()` 常用入口：

| API | 说明 |
| --- | --- |
| `GetTree(maxDepth, root)` | 获取 browsing context 树 |
| `CreateTab(background, userContext, referenceContext)` | 创建 tab |
| `CreateWindow(background, userContext)` | 创建窗口 |
| `Close(context, promptUnload)` | 关闭 context |
| `Reload(ignoreCache, wait, context)` | 重新加载 context |
| `SetViewport(width, height, dpr, context)` | 设置 viewport |
| `SetBypassCSP(enabled, context)` | 设置 CSP bypass |
| `CreateUserContext()` | 创建 user context |
| `GetUserContexts()` | 获取 user context 列表 |
| `RemoveUserContext(userContext)` | 删除 user context |
| `GetClientWindows()` | 获取客户端窗口信息 |

完整示例：

```bash
go run ./examples/25_browser_user_context
go run ./examples/26_browsing_context_advanced
go run ./examples/37_three_isolated_user_context_tabs
```

---

## 10. Script 能力

```go
value, _ := page.RunJS("return document.title")
expr, _ := page.RunJSExpr("navigator.userAgent")
result, _ := page.EvalHandle(`JSON.stringify({title: document.title})`, true)
preload, _ := page.AddPreloadScript(`() => { window.__ruyiReady = true }`)
_ = page.RemovePreloadScript(preload.ID)

fmt.Println(value, expr, result.Type)
```

常用入口：

| API | 说明 |
| --- | --- |
| `page.RunJS(script, args...)` | 执行脚本 |
| `page.RunJSExpr(expression, args...)` | 执行表达式 |
| `page.RunJSExprInSandbox(expression, sandbox, args...)` | 在 sandbox 中执行表达式 |
| `page.GetRealms(typeName)` | 获取 realms |
| `page.EvalHandle(expression, awaitPromise)` | 获取 handle / remote value |
| `page.AddPreloadScript(script)` | 注册 preload script |
| `page.RemovePreloadScript(scriptID)` | 删除 preload script |
| `page.DisownHandles(handles)` | 释放 handle |

推荐示例：

```bash
go run ./examples/07_javascript
go run ./examples/29_script_input_advanced
go run ./examples/32_script_events
```

---

## 11. 弹窗

```go
prompt, err := page.WaitPrompt(2 * time.Second)
if err != nil {
	panic(err)
}
fmt.Println(prompt["message"])

text := "alice"
if err := page.HandlePrompt(true, &text, 2*time.Second); err != nil {
	panic(err)
}
```

如果你想提前声明默认策略，可以在 `FirefoxOptions` 上设置：

```go
opts := ruyipage.NewFirefoxOptions().WithUserPromptHandler(map[string]string{
	"alert":   "accept",
	"confirm": "accept",
	"prompt":  "ignore",
	"default": "accept",
})
```

完整示例：

```bash
go run ./examples/17_user_prompts
```

---

## 12. Emulation

```go
_ = page.Emulation().SetUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)", "iPhone")
_ = page.Emulation().SetTimezone("America/New_York")
_ = page.Emulation().SetLocale([]string{"en-US", "en"})
_ = page.Emulation().SetGeolocation(39.9042, 116.4074, 100)

enableTouch := true
support := page.Emulation().ApplyMobilePreset(
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
	390,
	844,
	3.0,
	"portrait-primary",
	0,
	"en-US",
	"America/New_York",
	&enableTouch,
)
fmt.Println(support)
```

常用入口：

| API | 说明 |
| --- | --- |
| `SetUserAgent(userAgent, platform)` | 覆盖 UA |
| `SetTimezone(timezoneID)` | 覆盖时区 |
| `SetLocale(locales)` | 覆盖语言 |
| `SetGeolocation(lat, lng, accuracy)` | 覆盖定位 |
| `SetScreenOrientation(type, angle)` | 覆盖屏幕方向 |
| `SetScreenSize(width, height, dpr)` | 覆盖屏幕尺寸 |
| `SetNetworkOffline(enabled)` | 设置网络离线 |
| `SetTouchEnabled(enabled, maxTouchPoints, scope)` | 设置触摸能力 |
| `SetJavaScriptEnabled(enabled)` | 开关 JS |
| `SetScrollbarType(type)` | 设置滚动条类型 |
| `SetForcedColorsMode(mode)` | 设置强制颜色模式 |
| `SetBypassCSP(enabled)` | 设置 CSP bypass |
| `ApplyMobilePreset(...)` | 一次性应用移动端预设 |

推荐示例：

```bash
go run ./examples/21_emulation
go run ./examples/21_mobile_google_emulation
go run ./examples/27_emulation_advanced
```

---

## 13. WebExtension

```go
extID, err := page.Extensions().InstallDir(`D:\path\to\extension_dir`)
if err != nil {
	panic(err)
}
defer func() {
	_ = page.Extensions().Uninstall(extID)
}()
```

常用入口：

| API | 说明 |
| --- | --- |
| `Install(path)` | 自动识别安装目录或压缩包 |
| `InstallDir(path)` | 安装扩展目录 |
| `InstallArchive(path)` | 安装 `.xpi` / 压缩扩展 |
| `Uninstall(extensionID)` | 卸载指定扩展 |
| `UninstallAll()` | 卸载全部扩展 |
| `InstalledExtensions()` | 查看已安装扩展 |

完整示例：

```bash
go run ./examples/22_web_extension
```

---

## 14. 常见公开错误类型

Go 版已经公开了和 Python 对齐的一组主要错误类型，常见的有：

- `ruyipage.RuyiPageError`
- `ruyipage.ElementNotFoundError`
- `ruyipage.ElementLostError`
- `ruyipage.ContextLostError`
- `ruyipage.BiDiError`
- `ruyipage.PageDisconnectedError`
- `ruyipage.JavaScriptError`
- `ruyipage.BrowserConnectError`
- `ruyipage.BrowserLaunchError`
- `ruyipage.AlertExistsError`
- `ruyipage.WaitTimeoutError`
- `ruyipage.NoRectError`
- `ruyipage.CanNotClickError`
- `ruyipage.LocatorError`
- `ruyipage.IncorrectURLError`
- `ruyipage.NetworkInterceptError`

如果你在业务侧需要区分 URL 格式错误或网络拦截失败，可以直接按这些类型做分类处理。

---

## 15. 代表性示例

### 入门

- `examples/00_quickstart`
- `examples/01_basic_navigation`
- `examples/02_element_finding`
- `examples/03_element_interaction`
- `examples/04_wait_conditions`
- `examples/05_actions_chain`

### 页面、脚本、标签页

- `examples/06_screenshot`
- `examples/07_javascript`
- `examples/08_cookies`
- `examples/09_tabs`
- `examples/10_scrolling`
- `examples/13_iframe`
- `examples/14_shadow_dom`
- `examples/15_comprehensive`
- `examples/19_pdf_printing`
- `examples/20_advanced_input`

### 网络、下载、事件

- `examples/11_network_intercept`
- `examples/12_console_listener`
- `examples/18_advanced_network`
- `examples/23_download`
- `examples/24_navigation_events`
- `examples/28_network_data_collector`
- `examples/30_browsing_context_events`
- `examples/31_network_events`
- `examples/32_script_events`
- `examples/33_log_input_events`
- `examples/40_scraper_packet_capture`
- `examples/40_1_request_header_capture`

### 高级能力

- `examples/17_user_prompts`
- `examples/21_emulation`
- `examples/21_mobile_google_emulation`
- `examples/22_web_extension`
- `examples/25_browser_user_context`
- `examples/26_browsing_context_advanced`
- `examples/27_emulation_advanced`
- `examples/29_script_input_advanced`
- `examples/34_remaining_commands`
- `examples/35_native_bidi_drag`
- `examples/36_native_bidi_select`
- `examples/37_three_isolated_user_context_tabs`
- `examples/38_proxy_auth_ipinfo`
- `examples/39_attach_exist_browser`
- `examples/41_private_mode_userdir`
- `examples/42_2_action_visual_showcase`
- `examples/42_xpath_picker_complex_showcase`
- `examples/43_zhipin_xpath_picker`
- `examples/44_shopee_userdir_xpath`
- `examples/45_js_setter_untrusted_input`

### 协议对照辅助资产

- `examples/w3c_bidi/extract_w3c_bidi.py`
- `examples/w3c_bidi/generate_comparison.py`
- `examples/w3c_bidi/w3c_bidi_apis.json`

---

如果你准备从 Python 项目迁移到 Go 项目，建议阅读顺序是：

1. `Launch()` / `FirefoxQuickStartOptions`
2. `FirefoxOptions`
3. `FirefoxPage`、`FirefoxElement`
4. `Actions()`、`Listen()`、`Intercept()`、`Events()`
5. `Contexts()`、`Emulation()`、`Extensions()`
6. 对照 `examples/` 按编号逐步跑通
