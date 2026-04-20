package ruyipage

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pll177/ruyipage-go/internal/adapter"
	"github.com/pll177/ruyipage-go/internal/base"
	internalbrowser "github.com/pll177/ruyipage-go/internal/browser"
	internalpages "github.com/pll177/ruyipage-go/internal/pages"
	"github.com/pll177/ruyipage-go/internal/support"
)

// ChromeOptions 是 Chrome 启动的最小公开配置。
//
// 当前 Chrome 路线通过 chromedriver 提供 WebDriver BiDi 桥接，因此必须显式
// 指定一份与本机 Chrome 版本匹配的 chromedriver.exe。
type ChromeOptions struct {
	chromedriverPath    string
	chromedriverVersion string
	chromeBinary        string
	userDataDir         string
	host                string
	port                int
	headless            bool
	args                []string
	debuggerAddress     string
}

// NewChromeOptions 返回一个空白 ChromeOptions。
func NewChromeOptions() *ChromeOptions {
	return &ChromeOptions{}
}

// WithChromedriverPath 指定 chromedriver.exe 绝对路径。
//
// 与 WithChromedriverVersion 二选一；同时设置时 path 优先，不会触发自动下载。
func (o *ChromeOptions) WithChromedriverPath(path string) *ChromeOptions {
	o.chromedriverPath = path
	return o
}

// WithChromedriverVersion 开启 chromedriver 自动下载。
//
// version 支持：
//   - ""、"stable"、"latest"：使用 Chrome for Testing 的 LATEST_RELEASE_STABLE
//   - 大版本号，如 "131"：自动解析为匹配的精确版本
//   - 精确版本号，如 "131.0.6778.85"：直接使用
//
// 下载产物缓存在 %LOCALAPPDATA%\ruyipage\chromedriver\{version}\chromedriver.exe，
// 已存在时不会重复下载。仅支持 Windows x64。
func (o *ChromeOptions) WithChromedriverVersion(version string) *ChromeOptions {
	o.chromedriverVersion = version
	return o
}

// WithBrowserPath 指定 Chrome 可执行文件路径；不传时由 chromedriver 自动探测。
func (o *ChromeOptions) WithBrowserPath(path string) *ChromeOptions {
	o.chromeBinary = path
	return o
}

// WithUserDataDir 指定 Chrome user-data-dir，用于复用登录态与配置。
func (o *ChromeOptions) WithUserDataDir(path string) *ChromeOptions {
	o.userDataDir = path
	return o
}

// WithPort 指定 chromedriver 监听端口；不传时自动从 9515 起查找空闲端口。
func (o *ChromeOptions) WithPort(port int) *ChromeOptions {
	o.port = port
	return o
}

// WithHost 指定 chromedriver 监听 host；不传时使用 127.0.0.1。
func (o *ChromeOptions) WithHost(host string) *ChromeOptions {
	o.host = host
	return o
}

// Headless 设置是否启用 Chrome --headless=new。
func (o *ChromeOptions) Headless(headless bool) *ChromeOptions {
	o.headless = headless
	return o
}

// AddArgument 追加任意 Chrome 启动参数。
func (o *ChromeOptions) AddArgument(args ...string) *ChromeOptions {
	o.args = append(o.args, args...)
	return o
}

// WithDebuggerAddress 指定一个已经用 --remote-debugging-port 启动的 Chrome 地址
// （例如 "127.0.0.1:9222"）。设置后 chromedriver 不会再启动新的 Chrome 进程，
// 而是接管这个已存在的 Chrome，并对外提供 BiDi。
//
// 典型用法：
//
//	"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe" ^
//	    --remote-debugging-port=9222 ^
//	    --user-data-dir=D:\chrome_userdir
func (o *ChromeOptions) WithDebuggerAddress(address string) *ChromeOptions {
	o.debuggerAddress = address
	return o
}

// ChromePage 是基于 chromedriver + BiDi 的 Chrome 顶层页面对象。
//
// 内部复用了 Firefox 侧的 FirefoxPage / FirefoxBase，仅在生命周期上
// 额外负责清理 chromedriver.exe 进程。
type ChromePage struct {
	*FirefoxPage

	driverSession *adapter.ChromeDriverSession
}

// NewChromePage 启动 chromedriver、创建 BiDi session 并返回 ChromePage。
func NewChromePage(options *ChromeOptions) (*ChromePage, error) {
	if options == nil {
		options = NewChromeOptions()
	}

	chromedriverPath := options.chromedriverPath
	if strings.TrimSpace(chromedriverPath) == "" {
		resolved, err := adapter.EnsureChromedriver(options.chromedriverVersion)
		if err != nil {
			return nil, err
		}
		chromedriverPath = resolved
	}

	session, err := adapter.StartChromeDriverSession(adapter.ChromeDriverConfig{
		ChromeDriverPath: chromedriverPath,
		ChromeBinary:     options.chromeBinary,
		UserDataDir:      options.userDataDir,
		Host:             options.host,
		Port:             options.port,
		Headless:         options.headless,
		Args:             append([]string{}, options.args...),
		DebuggerAddress:  options.debuggerAddress,
	})
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%d", session.Host, session.Port)
	driver := base.NewBrowserBiDiDriver(address)
	if err := driver.Start(session.WSURL, 10*time.Second); err != nil {
		_ = adapter.StopChromeDriverSession(session)
		return nil, support.NewBrowserConnectError("连接 Chrome BiDi WebSocket 失败", err)
	}
	driver.SetSessionID(session.SessionID)

	info := &internalbrowser.ProbeInfo{
		Address:      address,
		Host:         session.Host,
		Port:         session.Port,
		Ready:        true,
		ProbeState:   internalbrowser.ProbeStateAttachable,
		WSURL:        session.WSURL,
		Driver:       driver,
		SessionID:    session.SessionID,
		SessionOwned: true,
	}

	firefox, err := internalbrowser.CreateFirefoxFromProbeInfo(info)
	if err != nil {
		_ = driver.Stop()
		_ = adapter.StopChromeDriverSession(session)
		return nil, err
	}

	innerPage, err := internalpages.NewFirefoxPageFromBrowser(firefox, "")
	if err != nil {
		_ = firefox.Quit(time.Second, true)
		_ = adapter.StopChromeDriverSession(session)
		return nil, err
	}

	firefoxPageRegistryMu.Lock()
	if existing := firefoxPageRegistry[address]; existing != nil {
		firefoxPageRegistryMu.Unlock()
		_ = adapter.StopChromeDriverSession(session)
		return &ChromePage{FirefoxPage: existing}, nil
	}
	page := newFirefoxPageFromInner(innerPage, address)
	firefoxPageRegistry[address] = page
	firefoxPageRegistryMu.Unlock()

	return &ChromePage{
		FirefoxPage:   page,
		driverSession: session,
	}, nil
}

// ChromedriverProcess 返回受管 chromedriver 进程，便于日志观察或信号控制。
func (p *ChromePage) ChromedriverProcess() *exec.Cmd {
	if p == nil || p.driverSession == nil {
		return nil
	}
	return p.driverSession.Process
}

// Quit 关闭浏览器并结束受管的 chromedriver 进程。
func (p *ChromePage) Quit(timeout time.Duration, force bool) error {
	if p == nil {
		return nil
	}

	var firstErr error
	if p.FirefoxPage != nil {
		if err := p.FirefoxPage.Quit(timeout, force); err != nil {
			firstErr = err
		}
	}
	if p.driverSession != nil {
		if err := adapter.StopChromeDriverSession(p.driverSession); err != nil && firstErr == nil {
			firstErr = err
		}
		p.driverSession = nil
	}
	return firstErr
}
