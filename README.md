# ruyipage-go

> 基于 **Firefox + WebDriver BiDi** 的 Go 自动化库，适合页面分析、数据采集、网络监听、指纹浏览器接管与高层自动化封装。

## 仓库地址

- 当前 Go 仓库：<https://github.com/pll177/ruyipage-go>
- 上游 Python 仓库：<https://github.com/LoseNine/ruyipage>

## 项目说明

`ruyipage-go` 是 `ruyiPage` 的 Go 版本实现，延续 Firefox + BiDi 路线，提供更适合 Go 项目的页面、元素、标签页、frame、network、events、downloads、emulation 等高层 API。

### 当前能力

- Firefox 启动、接管、自动探测已打开浏览器
- 页面导航、标签页切换、iframe / nested iframe 访问
- 元素查找、点击、输入、拖拽、滚轮、原生动作链
- Shadow DOM、open / closed shadow root
- Cookies、下载、PDF、截图、用户提示框
- network / script / browsingContext / input / log 事件
- user context、viewport、storage、WebExtension
- XPath Picker 与 Go 代码自动生成

---

## 安装

### 远程安装 / 拉取依赖

如果你要在其他 Go 项目里直接引用：

```bash
go get github.com/pll177/ruyipage-go@latest
```

然后在代码里这样导入：

```go
import ruyipage "github.com/pll177/ruyipage-go"
```

> 注意：`go get` 使用的是 Go module 路径，正确写法是 `github.com/pll177/ruyipage-go`，不是带 `https://` 的 URL。

### 从源码使用

```bash
git clone https://github.com/pll177/ruyipage-go.git
cd ruyipage-go
go mod tidy
```

---

## 环境要求

- Windows
- Go 1.26+
- Firefox 或兼容 Firefox 内核浏览器

仓库示例默认使用：

- Firefox 路径：`C:\Users\pll177\Desktop\core\firefox.exe`
- 接管已打开浏览器命令：

```powershell
"C:\Users\pll177\Desktop\core\firefox.exe" -remote-debugging-port 9222
```

如果你的环境不同，可修改：

- `examples/internal/exampleutil/special_env.go`

---

## 快速开始

### 1）最简单启动

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
	defer page.Quit(0, false)

	if err := page.Get("https://example.com"); err != nil {
		panic(err)
	}
	page.Wait().Sleep(time.Second)

	title, _ := page.Title()
	fmt.Println(title)
}
```

### 2）自定义 `FirefoxOptions`

```go
package main

import (
	"fmt"

	ruyipage "github.com/pll177/ruyipage-go"
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

---

## 常见入口

| 能力 | 入口 |
| --- | --- |
| 页面导航 | `page.Get()` / `page.Refresh()` / `page.Back()` / `page.Forward()` |
| 元素查找 | `page.Ele()` / `page.Eles()` / `ele.Ele()` |
| 元素交互 | `ele.ClickSelf()` / `ele.Input()` / `ele.Text()` / `ele.Attr()` |
| 动作链 | `page.Actions()` |
| frame | `page.GetFrame()` / `frame.GetFrame()` |
| shadow root | `ele.ShadowRoot()` / `ele.ClosedShadowRoot()` |
| 下载 | `page.Downloads()` |
| 网络 | `page.Network()` / `page.Intercept()` |
| 事件 | `page.Events()` / `page.Navigation()` |
| 提示框 | `page.WaitPrompt()` / `page.AcceptPrompt()` |
| 上下文 | `page.Contexts()` / `page.BrowserTools()` |
| 模拟能力 | `page.Emulation()` |
| 扩展 | `page.Extensions()` |

---

## XPath Picker

启用后，页面右下角会出现半透明浮窗，支持：

- 显示元素名称、文本、绝对 XPath、相对 XPath
- 显示元素中心坐标
- 一键复制 XPath
- 自动生成对应的 Go 访问代码
- 支持 iframe、嵌套 iframe、open shadow root、closed shadow root

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

---

## 接管已打开浏览器

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

---

## 示例

建议优先看这些目录：

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
- `examples\39_attach_exist_browser`
- `examples\42_xpath_picker_complex_showcase`
- `examples\43_zhipin_xpath_picker`
- `examples\44_shopee_userdir_xpath`
- `examples\45_js_setter_untrusted_input`
- `examples\quickstart_bing_search`
- `examples\quickstart_cloudfare`
- `examples\quickstart_fingerprint_browser`
- `examples\quickstart_private_mode`

---

## 项目结构

- `examples`：完整示例
- `internal/adapter`：Firefox 接入与端口/进程探测
- `internal/base`：传输、dispatcher、driver 基础设施
- `internal/bidi`：BiDi 协议命令封装
- `internal/browser`：浏览器生命周期
- `internal/config`：配置实现
- `internal/elements`：元素实现
- `internal/pages`：page / tab / frame 实现
- `internal/support`：工具函数与通用支持
- `internal/units`：高层 manager 与组合能力
- `testdata/examples/test_pages`：本地测试页面

---

## 分支策略

本仓库只保留一个长期分支：

- `main`

版本点通过 tag 维护，例如当前已经有：

- `v1`

不会保留其他长期分支。
