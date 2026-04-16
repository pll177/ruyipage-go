# ruyipage-go

`ruyipage-go` 是 `ruyipage-python` 的 Go 版本实现，核心方向保持一致：以 **Firefox + WebDriver BiDi** 为底层，提供更高层、可直接落地的自动化 API，适合页面分析、数据采集、事件监听、网络拦截、指纹浏览器接管等场景。

## 核心特性

- 基于 **Firefox + WebDriver BiDi**
- 面向 Go 的高层页面、元素、标签页、frame API
- 支持页面导航、元素查找、输入、点击、拖拽、滚轮等原生动作链
- 支持 Cookies、下载、PDF、截图、弹窗处理、事件监听、网络拦截
- 支持 user context、viewport、emulation、WebExtension、本地存储
- 支持接管已打开的 Firefox / Firefox 指纹浏览器
- 内置 **XPath Picker**，可生成包含 iframe / shadow root 访问链的 Go 代码片段

## 环境要求

- Windows
- Go 1.26+
- Firefox 或兼容的 Firefox 内核浏览器

仓库中的示例默认使用：

- Firefox 路径：`C:\Users\pll177\Desktop\core\firefox.exe`
- 已打开浏览器接管命令：

```powershell
"C:\Users\pll177\Desktop\core\firefox.exe" -remote-debugging-port 9222
```

如果你的环境不同，可自行修改 `examples/internal/exampleutil/special_env.go` 里的固定路径。

## 安装与初始化

当前仓库以源码方式使用：

```powershell
git clone <your-repo-url>
cd ruyipage-go
go mod tidy
```

## 最简单启动

最短路径可以直接使用 `Launch()`：

```go
package main

import (
	"fmt"
	"time"

	ruyipage "ruyipage-go"
)

func main() {
	page, err := ruyipage.Launch()
	if err != nil {
		panic(err)
	}
	defer page.Quit(0, false)

	if err := page.Get("https://example.com"); err != nil {
		panic(err)
	}
	page.Wait().Sleep(time.Second)

	title, _ := page.Title()
	fmt.Println(title)
}
```

直接运行仓库示例：

```powershell
go run .\examples\00_quickstart
```

## 使用 `FirefoxOptions` 自定义启动参数

```go
package main

import (
	"fmt"

	ruyipage "ruyipage-go"
)

func main() {
	opts := ruyipage.NewFirefoxOptions().
		WithBrowserPath(`D:\Firefox\firefox.exe`).
		WithUserDir(`D:\ruyipage_userdir`).
		EnableHeadless(false)

	page, err := ruyipage.NewFirefoxPage(opts)
	if err != nil {
		panic(err)
	}
	defer page.Quit(0, false)

	if err := page.Get("https://www.example.com"); err != nil {
		panic(err)
	}

	title, _ := page.Title()
	fmt.Println(title)
}
```

## 私密模式

```go
opts := ruyipage.NewFirefoxOptions().
	WithUserDir(`D:\ruyipage_userdir`).
	EnablePrivateMode(true)

page, err := ruyipage.NewFirefoxPage(opts)
```

对应示例：

```powershell
go run .\examples\quickstart_private_mode
go run .\examples\41_private_mode_userdir
```

## XPath Picker

启用后，页面右下角会出现浮窗，显示：

- 元素名称与文本
- XPath 绝对路径 / 相对路径
- 元素中心坐标
- 自动生成的 `ruyiPage` Go 代码

当前代码生成已经支持：

- 同源 iframe
- 嵌套 iframe
- open shadow root
- closed shadow root

最小示例：

```go
opts := ruyipage.NewFirefoxOptions().
	WithBrowserPath(`D:\Firefox\firefox.exe`).
	EnableHeadless(false).
	EnableXPathPicker(true).
	WithWindowSize(1600, 1100)

page, err := ruyipage.NewFirefoxPage(opts)
if err != nil {
	panic(err)
}
defer page.Quit(0, false)

_ = page.Get("https://example.com")
```

推荐直接运行：

```powershell
go run .\examples\42_xpath_picker_complex_showcase
```

## 接管已打开浏览器

如果浏览器已经手工打开，可以直接接管已有实例。

### 固定地址接管

```go
page, err := ruyipage.Attach("127.0.0.1:9222")
```

### 自动扫描端口接管

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

### 按进程特征自动接管

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

```powershell
go run .\examples\39_attach_exist_browser
```

## 常见能力

| 能力 | 入口 |
| --- | --- |
| 页面导航 | `page.Get()` / `page.Refresh()` / `page.Back()` / `page.Forward()` |
| 元素查找 | `page.Ele()` / `page.Eles()` / `ele.Ele()` |
| 元素交互 | `ele.ClickSelf()` / `ele.Input()` / `ele.Text()` / `ele.Attr()` |
| 动作链 | `page.Actions()` |
| 下载 | `page.Downloads()` |
| 网络 | `page.Network()` / `page.Intercept()` |
| 通用事件 | `page.Events()` |
| 导航事件 | `page.Navigation()` |
| 用户提示框 | `page.WaitPrompt()` / `page.AcceptPrompt()` |
| frame | `page.GetFrame()` / `frame.GetFrame()` |
| shadow root | `ele.ShadowRoot()` / `ele.ClosedShadowRoot()` |
| user context | `page.Contexts()` / `page.BrowserTools()` |
| 模拟能力 | `page.Emulation()` |
| 扩展 | `page.Extensions()` |

## 示例目录

仓库保留了完整示例，建议按编号阅读：

- `examples\00_quickstart`
- `examples\01_basic_navigation`
- `examples\02_element_finding`
- `examples\03_element_interaction`
- `examples\05_actions_chain`
- `examples\08_cookies`
- `examples\09_tabs`
- `examples\13_iframe`
- `examples\14_shadow_dom`
- `examples\17_user_prompts`
- `examples\18_advanced_network`
- `examples\31_network_events`
- `examples\32_script_events`
- `examples\35_native_bidi_drag`
- `examples\36_native_bidi_select`
- `examples\39_attach_exist_browser`
- `examples\42_xpath_picker_complex_showcase`
- `examples\43_zhipin_xpath_picker`
- `examples\44_shopee_userdir_xpath`
- `examples\45_js_setter_untrusted_input`
- `examples\quickstart_bing_search`
- `examples\quickstart_cloudfare`
- `examples\quickstart_fingerprint_browser`
- `examples\quickstart_private_mode`

## 项目结构

- `examples`：示例程序
- `internal/adapter`：Firefox 接入与进程/端口探测
- `internal/base`：传输、事件派发、driver 基础设施
- `internal/bidi`：BiDi 协议命令封装
- `internal/browser`：浏览器生命周期管理
- `internal/config`：配置实现
- `internal/elements`：元素内部实现
- `internal/pages`：page / tab / frame 实现
- `internal/support`：工具函数与辅助类型
- `internal/units`：高层 manager 与组合能力
- `testdata/examples/test_pages`：本地示例页面夹具

## 快速建议

- 想最快体验：先跑 `examples\00_quickstart`
- 想看 frame / shadow：跑 `examples\13_iframe`、`examples\14_shadow_dom`
- 想看 XPath Picker：跑 `examples\42_xpath_picker_complex_showcase`
- 想看接管已打开浏览器：跑 `examples\39_attach_exist_browser`
- 想看网络和事件：跑 `examples\18_advanced_network`、`examples\31_network_events`、`examples\32_script_events`

这个仓库当前仅保留 `main` 主分支作为唯一长期分支。
