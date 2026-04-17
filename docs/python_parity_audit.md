# ruyipage-go 对齐审计报告

> 注：本文件保留 2026-04-17 审计时的快照；若与当前仓库实现不一致，以当前代码为准。

审计日期：2026-04-17

对照基线：

- Python 主项目：`ruyipage-python`
- Go 翻译项目：`ruyipage-go`

审计方式：

- 静态比对公开 API、README、quickstart、examples、测试夹具
- 非侵入验证：
  - `go test ./...`
  - `python -m compileall ruyipage`

> 结论口径：本报告优先确认“公开面 + 示例”是否一致，并补充明显泄露、bug、验证缺口。  
> 未在本报告中标为“已确认缺失”的内容，不等于已经 100% 运行时对齐，只表示当前静态证据未直接证明其缺失。

---

## 1. 总结结论

### 1.1 已确认的大体对齐项

- 核心公开对象基本都有 Go 对应项：`FirefoxPage`、`FirefoxTab`、`FirefoxFrame`、`Firefox`、`FirefoxOptions`、`FirefoxElement`、`NoneElement`、`StaticElement`
- 顶层 attach / auto attach / find browser 入口基本有 Go 对应项
- 公共小类型与高层类型大部分已暴露：`Settings`、`Keys`、`By`、`ExtensionManager`、`BidiEvent`、`InterceptedRequest`、`DataPacket`、`DataCollector`、`NetworkData`、`CookieInfo`、`RealmInfo`、`ScriptRemoteValue`、`ScriptResult`、`PreloadScript`
- quickstart 数量已对齐：Python 4 个，Go 4 个
- 本地 HTML 夹具已对齐：`examples/test_pages/*.html` 与 `testdata/examples/test_pages/*.html` 文件名一一对应

### 1.2 已确认的不一致 / 缺口

1. **Go 的 `Launch()` 不是 Python 那种可带参数的快捷入口**
2. **Go 缺少两个 Python 已公开的错误类型：`IncorrectURLError`、`NetworkInterceptError`**
3. **Go 缺少 Python `examples/21-30` 共 11 个示例**
4. **Go 缺少 Python `examples/w3c_bidi/*` 参考资产**
5. **Go 仓库存在明显环境泄露：开发者本机路径、固定代理串、代理账号密码、固定 fpfile 路径**
6. **Go 项目当前没有任何 `_test.go` 回归测试文件，`go test ./...` 实际只有编译级检查**

### 1.3 风险判断

- **高优先级先处理泄露**：这些内容已经进入仓库源码和 README
- **其次补公开 API 缺口**：`Launch()` 便捷入口与错误类型不齐，会直接影响用户使用与兼容描述
- **再补 examples/资产**：这部分更偏“功能验证缺失”和“交付缺失”，不一定代表底层能力完全没有，但当前无法宣称与 Python 示例层完全一致

---

## 2. 公开 API 对齐

| 项目 | Python 基线 | Go 当前状态 | 结论 | 证据 |
| --- | --- | --- | --- | --- |
| 核心对象 | `FirefoxPage`/`FirefoxTab`/`FirefoxFrame`/`Firefox`/`FirefoxOptions`/元素对象 | 根包均有公开类型 | 基本对齐 | `ruyipage-python/ruyipage/__init__.py:22-54`；`ruyipage-go/firefox*.go`、`none_element.go`、`static_element.go` |
| 顶层 attach / find 入口 | `attach`、`attach_exist_browser`、`auto_attach_exist_browser`、`find_exist_browsers` 等 | `Attach`、`AttachExistingBrowser`、`AutoAttachExistingBrowser`、`FindExistingBrowsers` 等 | 基本对齐 | `ruyipage-python/ruyipage/__init__.py:228-451`；`ruyipage-go/entrypoints.go:40-279` |
| 版本公开面 | Python 公开 `__version__` | Go 公开 `Version` | 命名差异，可接受 | `ruyipage-python/ruyipage/version.py:1`；`ruyipage-go/version.go:3-4` |
| `launch()` 快捷入口 | Python `launch(...)` 支持 `browser_path`、`user_dir`、`headless`、`private`、`port`、超时等参数 | Go `Launch()` 为零参快捷入口 | **不一致 / 缺失** | `ruyipage-python/ruyipage/__init__.py:171-226`；`ruyipage-go/entrypoints.go:40-44`；`ruyipage-go/examples/41_private_mode_userdir/main.go:60-88` |
| 错误类型：`IncorrectURLError` | Python 已公开 | Go 未公开 | **缺失** | `ruyipage-python/ruyipage/errors.py:83-85`；`ruyipage-go/errors.go` |
| 错误类型：`NetworkInterceptError` | Python 已公开 | Go 未公开 | **缺失** | `ruyipage-python/ruyipage/errors.py:88-90`；`ruyipage-go/errors.go` |

### 2.1 公开 API 审计备注

- Go 侧核心公开对象总体不是空壳，结构上已经覆盖了 Python 主线对象。
- 真正明显的公开面缺口，审计时确认只有两类：
  - `Launch()` 快捷入口语义缺口
  - 错误类型公开面缺口

---

## 3. 示例、quickstart 与资产对齐

### 3.1 示例对齐结果

| 范围 | Python | Go | 结论 |
| --- | --- | --- | --- |
| `00-20` | 存在 | 存在 | 已覆盖 |
| `21-30` | 存在 | 不存在 | **缺失** |
| `31-45` | 存在 | 存在 | 已覆盖 |
| quickstart 4 个 | 存在 | 存在 | 已覆盖 |
| `examples/test_pages/*.html` | 11 个 | 11 个 | 已对齐 |
| `examples/w3c_bidi/*` | 存在 3 个文件 | 不存在 | **缺失** |

### 3.2 已确认缺失的示例

Go 当前缺少以下 Python 示例的一一对应目录：

- `21_emulation`
- `21_mobile_google_emulation`
- `22_web_extension`
- `23_download`
- `24_navigation_events`
- `25_browser_user_context`
- `26_browsing_context_advanced`
- `27_emulation_advanced`
- `28_network_data_collector`
- `29_script_input_advanced`
- `30_browsing_context_events`

对应基线路径：

- `ruyipage-python/examples/21_emulation.py`
- `ruyipage-python/examples/21_mobile_google_emulation.py`
- `ruyipage-python/examples/22_web_extension.py`
- `ruyipage-python/examples/23_download.py`
- `ruyipage-python/examples/24_navigation_events.py`
- `ruyipage-python/examples/25_browser_user_context.py`
- `ruyipage-python/examples/26_browsing_context_advanced.py`
- `ruyipage-python/examples/27_emulation_advanced.py`
- `ruyipage-python/examples/28_network_data_collector.py`
- `ruyipage-python/examples/29_script_input_advanced.py`
- `ruyipage-python/examples/30_browsing_context_events.py`

### 3.3 示例缺失的影响判断

- 这 **不自动等于底层 API 完全没有**，因为 Go README 和部分公开 manager 显示相关能力已有接口。
- 但这至少说明：
  - 示例层未完成一一迁移
  - 这些能力缺少同层级的可运行验证
  - 当前不能宣称“Go 示例覆盖面与 Python 一致”

### 3.4 `w3c_bidi` 资产缺失

Python 存在以下辅助资产：

- `ruyipage-python/examples/w3c_bidi/extract_w3c_bidi.py`
- `ruyipage-python/examples/w3c_bidi/generate_comparison.py`
- `ruyipage-python/examples/w3c_bidi/w3c_bidi_apis.json`

Go 当前不存在 `ruyipage-go/examples/w3c_bidi/`。

影响：

- 失去上游 BiDi 能力比对的参考材料
- 后续做协议覆盖核验时少了一套现成资产

---

## 4. 已确认的泄露、bug、缺失与风险

### LEAK-01：默认 Firefox 路径硬编码为开发者本机路径

- 严重级别：**High**
- 类型：泄露 + 功能偏差
- Python 基线：
  - Windows 默认路径是通用安装路径 `C:\Program Files\Mozilla Firefox\firefox.exe`
  - 还保留 macOS / Linux 分支
- Go 当前状态：
  - 默认路径被写死为 `C:\Users\<redacted>\Desktop\core\firefox.exe`
- 证据：
  - `ruyipage-python/ruyipage/_configs/firefox_options.py:34-38`
  - `ruyipage-go/internal/config/firefox_options.go:16`
  - `ruyipage-go/README.md:74,78`
- 影响：
  - 泄露开发者本机目录结构
  - 在其他机器上默认值大概率不可用
  - 与 Python 默认行为不一致
- 建议修复：
  - 把 Go 默认路径恢复为通用 Windows 默认安装路径，或改为运行时探测
  - README 中只保留占位路径，不保留个人目录

### LEAK-02：仓库内硬编码了真实风格代理串与账号密码

- 严重级别：**Critical**
- 类型：泄露
- Go 当前状态：
  - 示例辅助中硬编码 `proxy.example.com:8080:example-user:<redacted>`
  - 并可直接生成包含 `httpauth.username` / `httpauth.password` 的 fpfile
- 证据：
  - `ruyipage-go/examples/internal/exampleutil/special_env.go:16-20`
  - `ruyipage-go/examples/internal/exampleutil/special_env.go:57-91`
- 影响：
  - 泄露代理供应商、代理用户名、密码风格与使用方式
  - 后续示例可能误连真实外部代理
  - 容易被继续复制到 README、issue、日志输出
- 建议修复：
  - 全部替换为占位符或环境变量
  - 生成 fpfile 的逻辑保留，但内容来源改为环境变量或示例假值

### LEAK-03：指纹浏览器 fpfile 路径硬编码为开发者本机路径

- 严重级别：**High**
- 类型：泄露
- 证据：
  - `ruyipage-go/examples/quickstart_fingerprint_browser/main.go:11`
- 影响：
  - 再次暴露本机目录结构
  - 让示例默认依赖个人环境文件
- 建议修复：
  - 改成纯环境变量驱动
  - 文档中使用 `D:\path\to\profile.txt` 这类占位路径

### BUG-01：Go `Launch()` 缺少 Python 的参数化快捷入口

- 严重级别：**High**
- 类型：公开 API 缺失
- Python 基线：
  - `launch()` 可直接传 `browser_path`、`user_dir`、`headless`、`private`、`port`、窗口和超时参数
- Go 当前状态：
  - `Launch()` 固定零参，只能返回默认 quick start 页面
  - 为了对齐 Python `quickstart_private_mode.py`，Go 示例当时不得不自行实现本地 helper
- 证据：
  - `ruyipage-python/ruyipage/__init__.py:171-226`
  - `ruyipage-go/entrypoints.go:40-44`
  - `ruyipage-go/examples/41_private_mode_userdir/main.go:60-88`
- 影响：
  - 与 Python 快捷入口语义不一致
  - quickstart/README 需要额外解释绕路写法
  - 用户迁移成本升高
- 建议修复：
  - 增加与 Python 对应的参数化快捷入口
  - 或补一个明确公开的 `LaunchWithOptions(...)`/`LaunchQuick(...)` 并同步 README、examples、API 映射

### BUG-02：公开错误类型缺失 `IncorrectURLError`

- 严重级别：**Medium**
- 类型：公开 API 缺失
- 证据：
  - `ruyipage-python/ruyipage/errors.py:83-85`
  - `ruyipage-go/errors.go`
- 影响：
  - 无法与 Python 一样对 URL 错误做独立分类
- 建议修复：
  - 在 Go 公开错误层补齐别名/结构与内部抛出点

### BUG-03：公开错误类型缺失 `NetworkInterceptError`

- 严重级别：**Medium**
- 类型：公开 API 缺失
- 证据：
  - `ruyipage-python/ruyipage/errors.py:88-90`
  - `ruyipage-go/errors.go`
- 影响：
  - 网络拦截失败无法按 Python 语义区分
- 建议修复：
  - 在 Go 公开错误层补齐别名/结构与内部抛出点

### GAP-01：`examples/21-30` 整段缺失

- 严重级别：**High**
- 类型：功能验证缺失 / 示例缺失
- 证据：
  - Python 存在 `21-30`
  - Go `examples/` 目录直接从 `20_advanced_input` 跳到 `31_network_events`
- 影响：
  - emulation、web extension、download、navigation events、user context、advanced browsing context、advanced network collector、script input 等能力缺少同层级对照示例
  - 这些功能当前无法按“已示例对齐”来验收
- 建议修复：
  - 一一补齐 11 个示例
  - 每个示例保留 Python 的关键步骤、打印和验证点

### GAP-02：`examples/w3c_bidi/*` 资产缺失

- 严重级别：**Medium**
- 类型：资产缺失
- 证据：
  - Python 存在 `examples/w3c_bidi/*`
  - Go 不存在 `examples/w3c_bidi/`
- 影响：
  - 协议覆盖与上游能力比对材料缺失
- 建议修复：
  - 迁移为 `examples/w3c_bidi/` 或 `docs/` 下辅助资产

### RISK-01：当前没有自动化回归测试

- 严重级别：**High**
- 类型：质量风险
- 证据：
  - 审计时 `rg --files -g "*_test.go"` 无结果
  - `go test ./...` 输出均为 `[no test files]`
- 影响：
  - 当前“对齐 Python”的说法没有自动化兜底
  - 后续非常容易出现行为漂移却不自知
- 建议修复：
  - 至少补一层 smoke/regression：
    - 顶层入口
    - `Launch()` / attach
    - 错误类型
    - 迁移后的 `21-30` 示例

### DOC-01：README 仍要求直接修改源码文件适配本地环境

- 严重级别：**Medium**
- 类型：文档问题
- 证据：
  - `ruyipage-go/README.md:72-81`
  - README 明示若环境不同可修改 `examples/internal/exampleutil/special_env.go`
- 影响：
  - 诱导把个人环境再次写回仓库
  - 放大 LEAK-01 / LEAK-02 风险
- 建议修复：
  - 改成环境变量方案
  - README 仅说明 `.env` / 环境变量 / 本地未提交配置文件

---

## 5. 审计期间执行的验证

### 5.1 Go

执行：

```powershell
go test ./...
```

结果：

- 所有包可编译
- 但输出全部为 `[no test files]`
- 这说明当前只有“可编译性”通过，**不等于行为已经回归验证**

### 5.2 Python

执行：

```powershell
python -m compileall ruyipage
```

结果：

- Python 包可正常编译
- 可作为当前静态基线参考

---

## 6. 建议的下个会议实施顺序

### P0：先清泄露

1. 移除示例辅助中的真实代理串与账号密码
2. 移除所有 `C:\Users\<redacted>\...` 形式的默认路径
3. README 改为占位路径 + 环境变量说明
4. `quickstart_fingerprint_browser` 改成纯环境变量输入

### P1：补公开 API 缺口

1. 设计并实现与 Python 对应的参数化 `Launch()` 快捷入口
2. 补齐 `IncorrectURLError`
3. 补齐 `NetworkInterceptError`
4. 更新 README / quickstart / 示例调用方式

### P2：补示例与资产

1. 迁移 `examples/21-30`
2. 迁移 `examples/w3c_bidi/*`
3. 保持 Python 关键步骤、打印、验证点一致

### P3：补最小回归

1. 为新增/修复点补 `_test.go`
2. 至少覆盖：
   - 顶层入口
   - 错误类型公开面
   - 关键示例 smoke
   - 泄露项不再出现的静态检查

---

## 7. 最终判断

当前 `ruyipage-go` **不能认定为与 `ruyipage-python` 完全一致**。

更准确的判断是：

- **核心对象与主线能力大体已经搭起来了**
- **但公开快捷入口、错误公开面、示例层、参考资产层仍未完全对齐**
- **并且仓库内存在应优先处理的泄露问题**

如果下个会议要进入实现阶段，建议按本报告第 6 节顺序推进。
