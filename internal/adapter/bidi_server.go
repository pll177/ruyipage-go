package adapter

import (
	"fmt"
	"os/exec"
	"time"

	"ruyipage-go/internal/base"
	"ruyipage-go/internal/bidi"
	"ruyipage-go/internal/config"
	"ruyipage-go/internal/support"
)

const (
	defaultBiDiServerReadyTimeout   = 30 * time.Second
	defaultBiDiServerConnectTimeout = 10 * time.Second
	defaultBiDiServerCommandTimeout = 10 * time.Second
)

var (
	bidiServerWaitForFirefox = WaitForFirefox
	bidiServerGetBiDiWSURL   = GetBiDiWSURL
	bidiServerLaunchFirefox  = LaunchFirefox
	bidiServerFindFreePort   = support.FindFreePort
)

// BiDiServer 封装 Firefox 启动到 BiDi 会话就绪的辅助流程。
type BiDiServer struct {
	options *config.FirefoxOptions

	driver       *base.BrowserBiDiDriver
	process      *exec.Cmd
	registry     *ContextRegistry
	adapter      *ContextEventAdapter
	sessionID    string
	sessionOwned bool
}

// NewBiDiServer 创建一个新的 BiDiServer 门面。
func NewBiDiServer(options *config.FirefoxOptions) *BiDiServer {
	if options == nil {
		options = config.NewFirefoxOptions()
	}
	return &BiDiServer{
		options: options.Clone(),
	}
}

// Driver 返回当前连接的 BrowserBiDiDriver。
func (s *BiDiServer) Driver() *base.BrowserBiDiDriver {
	if s == nil {
		return nil
	}
	return s.driver
}

// Process 返回当前受管 Firefox 进程。
func (s *BiDiServer) Process() *exec.Cmd {
	if s == nil {
		return nil
	}
	return s.process
}

// ContextRegistry 返回当前 context 注册表。
func (s *BiDiServer) ContextRegistry() *ContextRegistry {
	if s == nil {
		return nil
	}
	return s.registry
}

// SessionID 返回当前持有的 session id。
func (s *BiDiServer) SessionID() string {
	if s == nil {
		return ""
	}
	return s.sessionID
}

// Connect 连接 Firefox BiDi Server；launch=true 时允许启动 Firefox。
func (s *BiDiServer) Connect(launch bool) (*base.BrowserBiDiDriver, error) {
	if s == nil || s.options == nil {
		return nil, support.NewBrowserConnectError("BiDiServer 未初始化", nil)
	}
	if s.driver != nil && s.driver.IsRunning() {
		return s.driver, nil
	}

	if err := s.prepareOptions(); err != nil {
		return nil, err
	}
	if err := s.options.WritePrefsToProfile(); err != nil {
		return nil, support.NewBrowserLaunchError("写入 Firefox profile 首选项失败", err)
	}

	if launch && !s.options.IsExistingOnly() {
		command, err := s.options.BuildCommand()
		if err != nil {
			return nil, support.NewBrowserLaunchError("构建 Firefox 启动命令失败", err)
		}
		process, err := bidiServerLaunchFirefox(command, nil)
		if err != nil {
			return nil, err
		}
		s.process = process
	}

	if !bidiServerWaitForFirefox(s.options.Host(), s.options.Port(), defaultBiDiServerReadyTimeout) {
		return nil, support.NewBrowserConnectError(
			fmt.Sprintf("Firefox Remote Agent 未就绪 (%s)", s.options.Address()),
			nil,
		)
	}

	wsURL, err := bidiServerGetBiDiWSURL(s.options.Host(), s.options.Port(), defaultBiDiServerConnectTimeout)
	if err != nil {
		return nil, err
	}

	driver := base.NewBrowserBiDiDriver(s.options.Address())
	if err := driver.Start(wsURL, defaultBiDiServerConnectTimeout); err != nil {
		return nil, err
	}

	sessionID, owned, err := s.createSession(driver)
	if err != nil {
		_ = driver.Stop()
		return nil, err
	}

	registry := NewContextRegistry()
	adapter := NewContextEventAdapter(driver, registry)
	if err := adapter.Start(); err != nil {
		if owned {
			_ = bidi.End(driver, defaultBiDiServerCommandTimeout)
		}
		_ = driver.Stop()
		return nil, err
	}

	if err := s.syncContexts(driver, registry); err != nil {
		adapter.Stop()
		if owned {
			_ = bidi.End(driver, defaultBiDiServerCommandTimeout)
		}
		_ = driver.Stop()
		return nil, err
	}

	s.driver = driver
	s.registry = registry
	s.adapter = adapter
	s.sessionID = sessionID
	s.sessionOwned = owned
	return driver, nil
}

// GetTopContext 返回当前第一个顶层 context id。
func (s *BiDiServer) GetTopContext() string {
	if s == nil || s.registry == nil {
		return ""
	}

	for _, contextID := range s.registry.AllIDs() {
		info, exists := s.registry.Get(contextID)
		if exists && info.Parent == "" {
			return contextID
		}
	}

	ids := s.registry.AllIDs()
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

// Disconnect 断开 BiDi 连接，并在受管进程存在时结束该进程。
func (s *BiDiServer) Disconnect() error {
	if s == nil {
		return nil
	}

	var firstErr error
	if s.adapter != nil {
		s.adapter.Stop()
		s.adapter = nil
	}
	if s.driver != nil {
		if s.sessionOwned {
			if err := bidi.End(s.driver, defaultBiDiServerCommandTimeout); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if err := s.driver.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.driver = nil
	}
	if s.process != nil && s.process.Process != nil {
		if err := s.process.Process.Kill(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	s.process = nil
	s.registry = nil
	s.sessionID = ""
	s.sessionOwned = false
	return firstErr
}

func (s *BiDiServer) prepareOptions() error {
	if err := s.options.Validate(); err != nil {
		return err
	}
	if !s.options.AutoPortEnabled() {
		return nil
	}

	start := s.options.AutoPortStart()
	if start <= 0 {
		start = s.options.Port()
	}
	port, err := bidiServerFindFreePort(start, start+100)
	if err != nil {
		return support.NewBrowserLaunchError(
			fmt.Sprintf("在端口范围 %d-%d 中找不到可用调试端口", start, start+99),
			err,
		)
	}
	s.options.WithPort(port)
	return nil
}

func (s *BiDiServer) createSession(driver *base.BrowserBiDiDriver) (string, bool, error) {
	status, err := bidi.Status(driver, defaultBiDiServerCommandTimeout)
	if err != nil {
		return "", false, support.NewBrowserConnectError("查询 Firefox BiDi session 状态失败", err)
	}

	if !status.Ready {
		_ = bidi.End(driver, defaultBiDiServerCommandTimeout)
	}

	result, err := bidi.New(driver, map[string]any{}, s.options.UserPromptHandler(), defaultBiDiServerCommandTimeout)
	if err != nil {
		return "", false, support.NewBrowserConnectError("创建 Firefox BiDi session 失败", err)
	}
	return result.SessionID, true, nil
}

func (s *BiDiServer) syncContexts(driver *base.BrowserBiDiDriver, registry *ContextRegistry) error {
	maxDepth := 0
	result, err := bidi.GetTree(driver, &maxDepth, "", defaultBiDiServerCommandTimeout)
	if err != nil {
		return err
	}
	registry.SyncFromTree(readContextTreeSlice(result["contexts"]))
	return nil
}

func readContextTreeSlice(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		result := make([]map[string]any, len(typed))
		copy(result, typed)
		return result
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if mapped, ok := item.(map[string]any); ok {
				result = append(result, mapped)
			}
		}
		return result
	default:
		return []map[string]any{}
	}
}
