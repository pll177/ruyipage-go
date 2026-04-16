package adapter

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"

	"ruyipage-go/internal/support"
)

const remoteAgentDefaultTimeout = 30 * time.Second

var (
	remoteAgentPollInterval = 500 * time.Millisecond
	startFirefoxProcess     = defaultStartFirefoxProcess
)

// GetBiDiWSURL 从 Firefox Remote Agent 获取 BiDi WebSocket 地址。
//
// 探测顺序与 Python 版保持一致：
//  1. 先试直连根路径 `ws://host:port`
//  2. 优先轮询 `http://host:port/json` 解析 `webSocketDebuggerUrl`
//  3. 同轮快速探测 `ws://host:port/session`，避免在 `/json` 不可用时白等完整超时
func GetBiDiWSURL(host string, port int, timeout time.Duration) (string, error) {
	if host == "" || port < 1 || port > 65535 {
		return "", support.NewBrowserConnectError("Firefox Remote Agent 地址无效", nil)
	}

	resolvedTimeout := resolveRemoteAgentTimeout(timeout)
	probeTimeout := minDuration(resolvedTimeout, 3*time.Second)

	address := net.JoinHostPort(host, strconv.Itoa(port))
	directWS := "ws://" + address
	sessionWS := directWS + "/session"
	jsonURL := "http://" + address + "/json"

	if probeRemoteAgentWSURL(directWS, probeTimeout) {
		return directWS, nil
	}

	deadline := time.Now().Add(resolvedTimeout)
	var lastErr error

	for {
		body, err := fetchRemoteAgentJSON(jsonURL, probeTimeout)
		if err == nil {
			if wsURL := extractRemoteAgentWSURL(body); wsURL != "" {
				return wsURL, nil
			}
		} else {
			lastErr = err
		}

		if probeRemoteAgentWSURL(sessionWS, probeTimeout) {
			return sessionWS, nil
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(minDuration(remoteAgentPollInterval, time.Until(deadline)))
	}

	return "", support.NewBrowserConnectError(
		fmt.Sprintf("无法从 %s 获取 Firefox BiDi WebSocket 地址", jsonURL),
		lastErr,
	)
}

// WaitForFirefox 等待 Firefox Remote Agent 端口就绪。
func WaitForFirefox(host string, port int, timeout time.Duration) bool {
	if host == "" || port < 1 || port > 65535 {
		return false
	}

	resolvedTimeout := resolveRemoteAgentTimeout(timeout)
	_, matched, err := support.WaitUntil(func() (struct{}, bool, error) {
		return struct{}{}, support.IsPortOpen(host, port, time.Second), nil
	}, resolvedTimeout, 300*time.Millisecond)
	return err == nil && matched
}

// LaunchFirefox 启动 Firefox 进程。
func LaunchFirefox(command []string, env map[string]string) (*exec.Cmd, error) {
	if len(command) == 0 {
		return nil, support.NewBrowserLaunchError("Firefox 启动命令不能为空", nil)
	}

	cmd, err := startFirefoxProcess(command, env)
	if err != nil {
		return nil, support.NewBrowserLaunchError("启动 Firefox 失败", err)
	}
	return cmd, nil
}

func fetchRemoteAgentJSON(url string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}

	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("HTTP %d", response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

func extractRemoteAgentWSURL(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || trimmed == "null" {
		return ""
	}

	if strings.HasPrefix(trimmed, "{") {
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			return ""
		}
		return readRemoteAgentWSURL(payload)
	}

	if strings.HasPrefix(trimmed, "[") {
		var payload []map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			return ""
		}
		for _, item := range payload {
			if wsURL := readRemoteAgentWSURL(item); wsURL != "" {
				return wsURL
			}
		}
	}

	return ""
}

func readRemoteAgentWSURL(payload map[string]any) string {
	if payload == nil {
		return ""
	}

	wsURL, _ := payload["webSocketDebuggerUrl"].(string)
	return wsURL
}

func probeRemoteAgentWSURL(wsURL string, timeout time.Duration) bool {
	if wsURL == "" {
		return false
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: timeout,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func resolveRemoteAgentTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return remoteAgentDefaultTimeout
	}
	return timeout
}

func minDuration(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

func defaultStartFirefoxProcess(command []string, env map[string]string) (*exec.Cmd, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if env != nil {
		cmd.Env = buildRemoteAgentEnv(env)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000,
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func buildRemoteAgentEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}

	values := os.Environ()
	filtered := make([]string, 0, len(values)+len(env))
	seen := make(map[string]struct{}, len(env))
	for key := range env {
		seen[strings.ToUpper(key)] = struct{}{}
	}

	for _, item := range values {
		index := strings.Index(item, "=")
		if index <= 0 {
			filtered = append(filtered, item)
			continue
		}
		key := strings.ToUpper(item[:index])
		if _, ok := seen[key]; ok {
			continue
		}
		filtered = append(filtered, item)
	}

	for key, value := range env {
		filtered = append(filtered, key+"="+value)
	}
	return filtered
}
