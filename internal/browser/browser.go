package browser

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ruyipage-go/internal/adapter"
	"ruyipage-go/internal/base"
	"ruyipage-go/internal/bidi"
	"ruyipage-go/internal/support"
)

// ProbeState 表示单个地址的探测结论。
type ProbeState string

const (
	// ProbeStateAttachable 表示当前地址可以创建并接管新的 BiDi session。
	ProbeStateAttachable ProbeState = "attachable"
	// ProbeStateOccupied 表示当前地址像是 Firefox BiDi，但唯一 session 已被占用。
	ProbeStateOccupied ProbeState = "occupied"
)

var (
	defaultFirefoxProcessNamePatterns = []string{
		"firefox.exe",
		"flowerbrowser.exe",
		"adsbrowser.exe",
	}
	defaultFirefoxCommandlinePatterns = []string{
		"--remote-debugging-port",
		"--marionette",
		"-contentproc",
		"-isforbrowser",
		"adspower",
		"flower",
		"firefox",
	}

	probeBiDiAddressForScan = ProbeBiDiAddress
	runPowerShellCommand    = defaultRunPowerShellCommand
)

// ProbeContextInfo 是 live probe 中保留的最小 context 信息。
type ProbeContextInfo struct {
	Context        string
	URL            string
	UserContext    string
	OriginalOpener any
}

// ProbeInfo 是单个地址探测完成后的复用结果。
type ProbeInfo struct {
	Address       string
	Host          string
	Port          int
	Ready         bool
	Message       string
	StatusMessage string
	ErrorMessage  string
	ProbeState    ProbeState
	WSURL         string
	Driver        *base.BrowserBiDiDriver
	SessionID     string
	SessionOwned  bool
	WindowCount   int
	TabCount      int
	ClientWindows []map[string]any
	Contexts      []ProbeContextInfo
	ScannedPorts  []int

	probeTimeout time.Duration
}

// IsAttachable 返回当前探测结果是否可直接复用为 attach 目标。
func (info *ProbeInfo) IsAttachable() bool {
	return info != nil && info.ProbeState == ProbeStateAttachable
}

// IsOccupied 返回当前探测结果是否命中了“端口有效但唯一 session 被占用”的语义。
func (info *ProbeInfo) IsOccupied() bool {
	return info != nil && info.ProbeState == ProbeStateOccupied
}

// HasLiveDriver 返回当前结果是否保留了可复用的 live probe driver。
func (info *ProbeInfo) HasLiveDriver() bool {
	return info != nil && info.Driver != nil && info.IsAttachable()
}

// ProbeBiDiAddress 探测单个地址是否为可接管的 Firefox BiDi 实例。
//
// 返回 nil, nil 表示该地址不可达、不是 BiDi 实例或探测失败；
// 返回 occupied 表示端口像是 Firefox BiDi，但当前无法创建新的 session；
// 返回 attachable 表示已经拿到可直接复用的探测信息。
func ProbeBiDiAddress(address string, timeout time.Duration, keepDriver bool) (*ProbeInfo, error) {
	host, port, err := splitProbeAddress(address)
	if err != nil {
		return nil, nil
	}

	resolvedTimeout := resolveProbeTimeout(timeout)
	if !support.IsPortOpen(host, port, resolvedTimeout) {
		return nil, nil
	}

	wsURL, err := adapter.GetBiDiWSURL(host, port, resolvedTimeout)
	if err != nil || wsURL == "" {
		return nil, nil
	}

	driver := base.NewBrowserBiDiDriver(address)
	if err := driver.Start(wsURL, resolvedTimeout); err != nil {
		stopProbeDriver(driver)
		return nil, nil
	}

	cleanup := func(sessionOwned bool) {
		if keepDriver {
			return
		}
		if sessionOwned {
			_ = bidi.End(driver, resolvedTimeout)
		}
		stopProbeDriver(driver)
	}

	status, err := bidi.Status(driver, resolvedTimeout)
	if err != nil {
		stopProbeDriver(driver)
		return nil, nil
	}

	if !status.Ready {
		stopProbeDriver(driver)
		return newOccupiedProbeInfo(
			address,
			host,
			port,
			status.Message,
			"",
			wsURL,
			resolvedTimeout,
		), nil
	}

	sessionResult, err := bidi.New(driver, map[string]any{}, nil, resolvedTimeout)
	if err != nil {
		if isMaximumActiveSessionsError(err) {
			stopProbeDriver(driver)
			return newOccupiedProbeInfo(
				address,
				host,
				port,
				status.Message,
				err.Error(),
				wsURL,
				resolvedTimeout,
			), nil
		}
		stopProbeDriver(driver)
		return nil, nil
	}

	sessionOwned := true
	clientWindows := readMapSliceFromAny(mustGetClientWindows(driver, resolvedTimeout))
	contexts := readMapSliceFromAny(mustGetBrowsingContexts(driver, resolvedTimeout))

	info := &ProbeInfo{
		Address:       address,
		Host:          host,
		Port:          port,
		Ready:         status.Ready,
		Message:       status.Message,
		StatusMessage: status.Message,
		ErrorMessage:  "",
		ProbeState:    ProbeStateAttachable,
		WSURL:         wsURL,
		SessionID:     sessionResult.SessionID,
		SessionOwned:  sessionOwned,
		WindowCount:   len(clientWindows),
		TabCount:      len(contexts),
		ClientWindows: cloneMapSlice(clientWindows),
		Contexts:      buildProbeContexts(contexts),
		probeTimeout:  resolvedTimeout,
	}
	if keepDriver {
		info.Driver = driver
	}

	cleanup(sessionOwned)
	return info, nil
}

// CloseProbeInfo 释放未被生命周期对象接管的 live probe 结果。
func CloseProbeInfo(info *ProbeInfo) error {
	if info == nil || info.Driver == nil {
		return nil
	}

	var firstErr error
	if info.SessionOwned {
		if err := bidi.End(info.Driver, resolveProbeTimeout(info.probeTimeout)); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if err := info.Driver.Stop(); err != nil && firstErr == nil {
		firstErr = err
	}

	info.Driver = nil
	info.SessionOwned = false
	return firstErr
}

// CloseProbeInfos 批量关闭 probe 结果中未被选中的 live driver。
func CloseProbeInfos(infos []*ProbeInfo, keepAddress string) error {
	var firstErr error
	for _, info := range infos {
		if info == nil || info.Address == keepAddress {
			continue
		}
		if err := CloseProbeInfo(info); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// ScanLiveProbes 并发扫描端口并保留 attachable 结果中的 live driver。
//
// 返回值同时包含 attachable 与 occupied 结果，供后续自动接管逻辑判断。
func ScanLiveProbes(
	host string,
	startPort int,
	endPort int,
	timeout time.Duration,
	maxWorkers int,
) ([]*ProbeInfo, error) {
	ports, err := buildInclusivePorts(startPort, endPort)
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return []*ProbeInfo{}, nil
	}

	results := scanPorts(resolveProbeHost(host), ports, timeout, maxWorkers, true)
	sortProbeInfos(results)
	return results, nil
}

// FindExistingBrowsers 并发扫描端口范围内可接管的 Firefox BiDi 实例。
func FindExistingBrowsers(
	host string,
	startPort int,
	endPort int,
	timeout time.Duration,
	maxWorkers int,
) ([]*ProbeInfo, error) {
	infos, err := ScanLiveProbes(host, startPort, endPort, timeout, maxWorkers)
	if err != nil {
		return nil, err
	}

	filtered := make([]*ProbeInfo, 0, len(infos))
	for _, info := range infos {
		if info != nil && info.IsAttachable() {
			filtered = append(filtered, info)
		}
	}
	sortProbeInfos(filtered)
	return filtered, nil
}

// FindCandidatePortsByProcess 按 Windows 进程特征发现候选监听端口，但不做 BiDi 探测。
func FindCandidatePortsByProcess(processNamePatterns []string, commandlinePatterns []string) ([]int, error) {
	processRows, err := decodeJSONRows(mustRunWindowsProcessQuery(
		"Get-CimInstance Win32_Process | Select-Object ProcessId, Name, CommandLine | ConvertTo-Json -Compress",
	))
	if err != nil {
		return nil, err
	}

	processNamePatterns = normalizePatterns(processNamePatterns, defaultFirefoxProcessNamePatterns)
	commandlinePatterns = normalizePatterns(commandlinePatterns, defaultFirefoxCommandlinePatterns)

	pids := make(map[int]struct{})
	for _, row := range processRows {
		pid := readIntValue(row["ProcessId"])
		if pid <= 0 {
			continue
		}

		name := strings.ToLower(readStringValue(row["Name"]))
		commandLine := strings.ToLower(readStringValue(row["CommandLine"]))
		if matchesAnyPattern(name, processNamePatterns) || matchesAnyPattern(commandLine, commandlinePatterns) {
			pids[pid] = struct{}{}
		}
	}

	if len(pids) == 0 {
		return []int{}, nil
	}

	listenRows, err := decodeJSONRows(mustRunWindowsProcessQuery(
		"Get-NetTCPConnection -State Listen | Select-Object LocalAddress, LocalPort, OwningProcess | ConvertTo-Json -Compress",
	))
	if err != nil {
		return nil, err
	}

	portSet := make(map[int]struct{})
	for _, row := range listenRows {
		pid := readIntValue(row["OwningProcess"])
		if _, ok := pids[pid]; !ok {
			continue
		}

		localAddress := readStringValue(row["LocalAddress"])
		if !isLocalProbeAddress(localAddress) {
			continue
		}

		port := readIntValue(row["LocalPort"])
		if port < 1 || port > 65535 {
			continue
		}
		portSet[port] = struct{}{}
	}

	ports := make([]int, 0, len(portSet))
	for port := range portSet {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	return ports, nil
}

// FindExistingBrowsersByProcess 先按进程特征发现候选端口，再并发探测 attachable 结果。
func FindExistingBrowsersByProcess(
	host string,
	timeout time.Duration,
	maxWorkers int,
	keepDriver bool,
	processNamePatterns []string,
	commandlinePatterns []string,
) ([]*ProbeInfo, error) {
	ports, err := FindCandidatePortsByProcess(processNamePatterns, commandlinePatterns)
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return []*ProbeInfo{}, nil
	}

	results := scanPorts(resolveProbeHost(host), ports, timeout, maxWorkers, keepDriver)
	filtered := make([]*ProbeInfo, 0, len(results))
	for _, info := range results {
		if info == nil || !info.IsAttachable() {
			continue
		}
		info.ScannedPorts = cloneIntSlice(ports)
		filtered = append(filtered, info)
	}
	sortProbeInfos(filtered)
	return filtered, nil
}

func newOccupiedProbeInfo(
	address string,
	host string,
	port int,
	statusMessage string,
	errorMessage string,
	wsURL string,
	timeout time.Duration,
) *ProbeInfo {
	return &ProbeInfo{
		Address:       address,
		Host:          host,
		Port:          port,
		Ready:         false,
		Message:       statusMessage,
		StatusMessage: statusMessage,
		ErrorMessage:  errorMessage,
		ProbeState:    ProbeStateOccupied,
		WSURL:         wsURL,
		ClientWindows: []map[string]any{},
		Contexts:      []ProbeContextInfo{},
		ScannedPorts:  []int{},
		probeTimeout:  timeout,
	}
}

func mustGetClientWindows(driver *base.BrowserBiDiDriver, timeout time.Duration) any {
	result, err := bidi.GetClientWindows(driver, timeout)
	if err != nil {
		return nil
	}
	return result["clientWindows"]
}

func mustGetBrowsingContexts(driver *base.BrowserBiDiDriver, timeout time.Duration) any {
	maxDepth := 0
	result, err := bidi.GetTree(driver, &maxDepth, "", timeout)
	if err != nil {
		return nil
	}
	return result["contexts"]
}

func buildProbeContexts(raw []map[string]any) []ProbeContextInfo {
	if len(raw) == 0 {
		return []ProbeContextInfo{}
	}

	contexts := make([]ProbeContextInfo, 0, len(raw))
	for _, item := range raw {
		contexts = append(contexts, ProbeContextInfo{
			Context:        readStringValue(item["context"]),
			URL:            readStringValue(item["url"]),
			UserContext:    resolveContextUserContext(item),
			OriginalOpener: item["originalOpener"],
		})
	}
	return contexts
}

func resolveContextUserContext(values map[string]any) string {
	userContext := readStringValue(values["userContext"])
	if userContext == "" {
		return "default"
	}
	return userContext
}

func scanPorts(host string, ports []int, timeout time.Duration, maxWorkers int, keepDriver bool) []*ProbeInfo {
	if len(ports) == 0 {
		return []*ProbeInfo{}
	}

	workers := resolveWorkerCount(maxWorkers, len(ports))
	results := make([]*ProbeInfo, 0, len(ports))
	var (
		resultsMu sync.Mutex
		jobs      = make(chan int)
		workerWG  sync.WaitGroup
	)

	for workerIndex := 0; workerIndex < workers; workerIndex++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			for port := range jobs {
				info, err := probeBiDiAddressForScan(fmt.Sprintf("%s:%d", host, port), timeout, keepDriver)
				if err != nil || info == nil {
					continue
				}

				resultsMu.Lock()
				results = append(results, info)
				resultsMu.Unlock()
			}
		}()
	}

	for _, port := range ports {
		jobs <- port
	}
	close(jobs)
	workerWG.Wait()
	return results
}

func buildInclusivePorts(startPort int, endPort int) ([]int, error) {
	if startPort > endPort {
		return []int{}, nil
	}
	if startPort < 1 || endPort > 65535 {
		return nil, fmt.Errorf("端口范围必须在 1-65535 之间，当前为 [%d, %d]", startPort, endPort)
	}

	ports := make([]int, 0, endPort-startPort+1)
	for port := startPort; port <= endPort; port++ {
		ports = append(ports, port)
	}
	return ports, nil
}

func resolveProbeHost(host string) string {
	if host == "" {
		return support.DefaultHost
	}
	return host
}

func resolveProbeTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return time.Second
	}
	if timeout < time.Second {
		return time.Second
	}
	return timeout
}

func resolveWorkerCount(maxWorkers int, total int) int {
	if total <= 0 {
		return 1
	}
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	if maxWorkers > total {
		return total
	}
	return maxWorkers
}

func splitProbeAddress(address string) (string, int, error) {
	index := strings.LastIndex(address, ":")
	if index <= 0 || index >= len(address)-1 {
		return "", 0, fmt.Errorf("address %q 缺少端口", address)
	}

	host := address[:index]
	port, err := strconv.Atoi(address[index+1:])
	if err != nil {
		return "", 0, err
	}
	if host == "" || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("address %q 非法", address)
	}
	return host, port, nil
}

func isMaximumActiveSessionsError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "maximum number of active sessions") ||
		(strings.Contains(message, "session not created") && strings.Contains(message, "maximum"))
}

func sortProbeInfos(infos []*ProbeInfo) {
	sort.Slice(infos, func(left int, right int) bool {
		return infos[left].Port < infos[right].Port
	})
}

func stopProbeDriver(driver *base.BrowserBiDiDriver) {
	if driver == nil {
		return
	}
	_ = driver.Stop()
}

func decodeJSONRows(payload []byte, err error) ([]map[string]any, error) {
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(payload))
	if trimmed == "" || trimmed == "null" {
		return []map[string]any{}, nil
	}

	switch trimmed[0] {
	case '{':
		var row map[string]any
		if err := json.Unmarshal(payload, &row); err != nil {
			return nil, err
		}
		return []map[string]any{row}, nil
	case '[':
		var rows []map[string]any
		if err := json.Unmarshal(payload, &rows); err != nil {
			return nil, err
		}
		return rows, nil
	default:
		return nil, fmt.Errorf("无法解析 PowerShell JSON 输出: %s", trimmed)
	}
}

func mustRunWindowsProcessQuery(script string) ([]byte, error) {
	return runPowerShellCommand(script)
}

func defaultRunPowerShellCommand(script string) ([]byte, error) {
	command := exec.Command("powershell", "-NoProfile", "-Command", script)
	return command.Output()
}

func normalizePatterns(patterns []string, defaults []string) []string {
	if patterns == nil {
		patterns = defaults
	}
	if len(patterns) == 0 {
		return []string{}
	}

	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		normalized = append(normalized, strings.ToLower(pattern))
	}
	return normalized
}

func matchesAnyPattern(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func isLocalProbeAddress(address string) bool {
	switch address {
	case "127.0.0.1", "::1", "0.0.0.0":
		return true
	default:
		return false
	}
}

func readMapSliceFromAny(value any) []map[string]any {
	switch typed := value.(type) {
	case nil:
		return []map[string]any{}
	case []map[string]any:
		return cloneMapSlice(typed)
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			mapValue, ok := item.(map[string]any)
			if !ok {
				continue
			}
			result = append(result, cloneMap(mapValue))
		}
		return result
	default:
		return []map[string]any{}
	}
}

func cloneMapSlice(values []map[string]any) []map[string]any {
	if len(values) == 0 {
		return []map[string]any{}
	}

	cloned := make([]map[string]any, len(values))
	for index, value := range values {
		cloned[index] = cloneMap(value)
	}
	return cloned
}

func cloneMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}

	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneIntSlice(values []int) []int {
	if len(values) == 0 {
		return []int{}
	}

	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}

func readStringValue(value any) string {
	text, _ := value.(string)
	return text
}

func readIntValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		number, err := strconv.Atoi(typed)
		if err == nil {
			return number
		}
	}
	return 0
}
