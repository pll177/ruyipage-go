package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pll177/ruyipage-go/internal/support"
)

// ChromeDriverSession 描述一次 chromedriver + W3C session 启动结果。
type ChromeDriverSession struct {
	Process   *exec.Cmd
	Host      string
	Port      int
	SessionID string
	WSURL     string
}

// ChromeDriverConfig 控制 chromedriver 启动与 W3C session 创建。
type ChromeDriverConfig struct {
	ChromeDriverPath string
	ChromeBinary     string
	UserDataDir      string
	Host             string
	Port             int
	Args             []string
	Headless         bool
	// DebuggerAddress 指定一个已经用 --remote-debugging-port 启动的 Chrome
	// host:port（例如 "127.0.0.1:9222"）。设置后 chromedriver 不会再启动新的
	// Chrome 进程，而是走 goog:chromeOptions.debuggerAddress 接管它。
	DebuggerAddress string
	ReadyTimeout    time.Duration
	SessionTimeout  time.Duration
}

const (
	chromeDriverDefaultHost           = "127.0.0.1"
	chromeDriverDefaultReadyTimeout   = 20 * time.Second
	chromeDriverDefaultSessionTimeout = 30 * time.Second
	chromeDriverAutoPortStart         = 9515
	chromeDriverAutoPortEnd           = chromeDriverAutoPortStart + 200
)

// StartChromeDriverSession 启动 chromedriver 进程并创建启用 BiDi 的 W3C session。
//
// 返回值中的 WSURL 可直接喂给 base.NewBrowserBiDiDriver。调用方负责在结束时
// 调用 StopChromeDriverSession 清理 chromedriver 进程。
func StartChromeDriverSession(config ChromeDriverConfig) (*ChromeDriverSession, error) {
	if strings.TrimSpace(config.ChromeDriverPath) == "" {
		return nil, support.NewBrowserLaunchError(
			"未指定 chromedriver 路径，请通过 ChromeOptions.WithChromedriverPath 设置", nil,
		)
	}

	host := config.Host
	if host == "" {
		host = chromeDriverDefaultHost
	}

	port := config.Port
	if port <= 0 {
		free, err := support.FindFreePort(chromeDriverAutoPortStart, chromeDriverAutoPortEnd)
		if err != nil {
			return nil, support.NewBrowserLaunchError("无法为 chromedriver 分配可用端口", err)
		}
		port = free
	}

	readyTimeout := config.ReadyTimeout
	if readyTimeout <= 0 {
		readyTimeout = chromeDriverDefaultReadyTimeout
	}
	sessionTimeout := config.SessionTimeout
	if sessionTimeout <= 0 {
		sessionTimeout = chromeDriverDefaultSessionTimeout
	}

	args := []string{
		"--port=" + strconv.Itoa(port),
	}
	if host != chromeDriverDefaultHost {
		args = append(args, "--allowed-ips="+host)
	}

	cmd := exec.Command(config.ChromeDriverPath, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	if err := cmd.Start(); err != nil {
		return nil, support.NewBrowserLaunchError("启动 chromedriver 失败", err)
	}

	if !waitChromeDriverReady(host, port, readyTimeout) {
		_ = cmd.Process.Kill()
		return nil, support.NewBrowserConnectError(
			fmt.Sprintf("chromedriver 在 %s:%d 未就绪", host, port), nil,
		)
	}

	sessionID, wsURL, err := createChromeBiDiSession(host, port, config, sessionTimeout)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	return &ChromeDriverSession{
		Process:   cmd,
		Host:      host,
		Port:      port,
		SessionID: sessionID,
		WSURL:     wsURL,
	}, nil
}

// StopChromeDriverSession 结束受管 chromedriver 进程。
func StopChromeDriverSession(session *ChromeDriverSession) error {
	if session == nil || session.Process == nil || session.Process.Process == nil {
		return nil
	}
	return session.Process.Process.Kill()
}

func waitChromeDriverReady(host string, port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	statusURL := fmt.Sprintf("http://%s/status", net.JoinHostPort(host, strconv.Itoa(port)))
	client := &http.Client{Timeout: time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(statusURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

func createChromeBiDiSession(host string, port int, config ChromeDriverConfig, timeout time.Duration) (string, string, error) {
	chromeOpts := map[string]any{}
	debuggerAddress := strings.TrimSpace(config.DebuggerAddress)

	if debuggerAddress != "" {
		// 接管已打开 Chrome 时不要再传 args / binary / user-data-dir，chromedriver 会直接
		// 走 debuggerAddress 接入，传了反而可能被拒绝。
		chromeOpts["debuggerAddress"] = debuggerAddress
	} else {
		chromeArgs := append([]string{}, config.Args...)
		if config.Headless {
			chromeArgs = append(chromeArgs, "--headless=new")
		}
		if strings.TrimSpace(config.UserDataDir) != "" {
			chromeArgs = append(chromeArgs, "--user-data-dir="+config.UserDataDir)
		}
		if len(chromeArgs) > 0 {
			chromeOpts["args"] = chromeArgs
		}
		if strings.TrimSpace(config.ChromeBinary) != "" {
			chromeOpts["binary"] = config.ChromeBinary
		}
	}

	payload := map[string]any{
		"capabilities": map[string]any{
			"alwaysMatch": map[string]any{
				"browserName":  "chrome",
				"webSocketUrl": true,
				"goog:chromeOptions": func() map[string]any {
					if len(chromeOpts) == 0 {
						return map[string]any{}
					}
					return chromeOpts
				}(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", support.NewBrowserConnectError("序列化 chromedriver new session 载荷失败", err)
	}

	sessionURL := fmt.Sprintf("http://%s/session", net.JoinHostPort(host, strconv.Itoa(port)))
	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(sessionURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", "", support.NewBrowserConnectError("向 chromedriver 请求 new session 失败", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", support.NewBrowserConnectError("读取 chromedriver new session 响应失败", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", support.NewBrowserConnectError(
			fmt.Sprintf("chromedriver 返回非成功状态 %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))),
			nil,
		)
	}

	var parsed struct {
		Value struct {
			SessionID    string `json:"sessionId"`
			Capabilities struct {
				WebSocketURL string `json:"webSocketUrl"`
			} `json:"capabilities"`
		} `json:"value"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", "", support.NewBrowserConnectError("解析 chromedriver new session 响应失败", err)
	}
	if parsed.Value.SessionID == "" || parsed.Value.Capabilities.WebSocketURL == "" {
		return "", "", support.NewBrowserConnectError(
			"chromedriver 未返回 BiDi webSocketUrl（请确认 chromedriver 与 Chrome 版本匹配且支持 BiDi）", nil,
		)
	}
	return parsed.Value.SessionID, parsed.Value.Capabilities.WebSocketURL, nil
}
