package ruyipage

import (
	"fmt"
	"time"

	internalbrowser "github.com/pll177/ruyipage-go/internal/browser"
	internalpages "github.com/pll177/ruyipage-go/internal/pages"
	"github.com/pll177/ruyipage-go/internal/support"
)

const (
	defaultAutoAttachEndPort       = 65535
	defaultFindExistingEndPort     = 9322
	defaultAutoAttachTimeout       = 200 * time.Millisecond
	defaultFindExistingTimeout     = 500 * time.Millisecond
	defaultAutoAttachMaxWorkers    = 64
	defaultFindExistingMaxWorkers  = 32
	defaultAttachTabIndex          = 1
	defaultAttachProcessMaxWorkers = 32
)

var (
	newFirefoxPageForEntry                = NewFirefoxPage
	pageFromLiveProbeInfo                 = pageFromProbeInfo
	scanLiveProbesForEntry                = internalbrowser.ScanLiveProbes
	closeProbeInfosForEntry               = internalbrowser.CloseProbeInfos
	findExistingBrowsersByProcessForEntry = func(host string, timeout time.Duration, maxWorkers int, keepDriver bool) ([]*internalbrowser.ProbeInfo, error) {
		return internalbrowser.FindExistingBrowsersByProcess(host, timeout, maxWorkers, keepDriver, nil, nil)
	}
	findCandidatePortsFromProcessForEntry = func() ([]int, error) {
		return internalbrowser.FindCandidatePortsByProcess(nil, nil)
	}
	probeBiDiAddressForEntry           = internalbrowser.ProbeBiDiAddress
	createFirefoxFromProbeInfoForEntry = internalbrowser.CreateFirefoxFromProbeInfo
	newFirefoxPageFromBrowserForEntry  = internalpages.NewFirefoxPageFromBrowser
)

// Launch 使用 quick_start 默认预设启动 FirefoxPage。
func Launch() (*FirefoxPage, error) {
	opts := NewFirefoxOptions()
	opts.QuickStart(DefaultFirefoxQuickStartOptions())
	return newFirefoxPageForEntry(opts)
}

// Attach 连接到已启动浏览器；未传地址时默认使用 127.0.0.1:9222。
func Attach(address ...string) (*FirefoxPage, error) {
	resolvedAddress := DefaultAddress
	if len(address) > 0 && address[0] != "" {
		resolvedAddress = address[0]
	}

	opts := NewFirefoxOptions().
		WithAddress(resolvedAddress).
		ExistingOnly(true)
	return newFirefoxPageForEntry(opts)
}

// AttachExistingBrowser 接管一个已启动浏览器，并按 tab_index / latest_tab 切到目标标签页。
func AttachExistingBrowser(address string, tabIndex int, latestTab bool) (*FirefoxPage, error) {
	page, err := Attach(resolveAttachAddress(address))
	if err != nil {
		return nil, err
	}

	contextID := resolveAttachContextID(page.TabIDs(), tabIndex, latestTab)
	if contextID == "" {
		return page, nil
	}
	if err := selectFirefoxPageContext(page, contextID); err != nil {
		return nil, err
	}
	return page, nil
}

// AutoAttachExistingBrowser 优先直连指定地址，失败后扫描端口范围自动接管浏览器。
func AutoAttachExistingBrowser(
	address string,
	host string,
	startPort int,
	endPort int,
	timeout time.Duration,
	maxWorkers int,
	tabIndex int,
	latestTab bool,
) (*FirefoxPage, error) {
	resolvedAddress := resolveAttachAddress(address)
	resolvedHost := resolveEntryHost(host)
	resolvedStartPort := resolveStartPort(startPort)
	resolvedEndPort := resolveEndPort(endPort, defaultAutoAttachEndPort)
	resolvedTimeout := resolveEntryTimeout(timeout, defaultAutoAttachTimeout)
	resolvedWorkers := resolveEntryWorkers(maxWorkers, defaultAutoAttachMaxWorkers)
	resolvedTabIndex := resolveAttachTabIndex(tabIndex)

	errorsList := make([]string, 0, 4)
	if address != "" {
		page, err := AttachExistingBrowser(resolvedAddress, resolvedTabIndex, latestTab)
		if err == nil {
			return page, nil
		}
		errorsList = append(errorsList, fmt.Sprintf("%s -> %v", resolvedAddress, err))
	}

	infos, err := scanLiveProbesForEntry(
		resolvedHost,
		resolvedStartPort,
		resolvedEndPort,
		resolvedTimeout,
		resolvedWorkers,
	)
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, support.NewBrowserConnectError(
			"没有发现可接管的 Firefox 浏览器，请检查调试端口是否开启，或扩大扫描端口范围。",
			nil,
		)
	}

	for _, info := range infos {
		if info == nil {
			continue
		}
		if !info.IsAttachable() {
			if info.IsOccupied() {
				errorsList = append(errorsList, formatOccupiedProbeDetail(info))
			}
			continue
		}

		page, pageErr := pageFromLiveProbeInfo(info, resolvedTabIndex, latestTab)
		if pageErr == nil {
			_ = closeProbeInfosForEntry(infos, info.Address)
			return page, nil
		}
		errorsList = append(errorsList, fmt.Sprintf("%s -> %v", info.Address, pageErr))
	}

	_ = closeProbeInfosForEntry(infos, "")
	return nil, buildAutoAttachError(
		"发现了可探测端口，但没有可真正接管的 Firefox 会话。这通常表示指纹浏览器已被自身或其他客户端占用了唯一 BiDi session。",
		errorsList,
	)
}

// FindExistingBrowsers 按端口范围扫描当前机器上可接管的 Firefox 浏览器。
func FindExistingBrowsers(
	host string,
	startPort int,
	endPort int,
	timeout time.Duration,
	maxWorkers int,
) ([]ProbeInfo, error) {
	infos, err := scanLiveProbesForEntry(
		resolveEntryHost(host),
		resolveStartPort(startPort),
		resolveEndPort(endPort, defaultFindExistingEndPort),
		resolveEntryTimeout(timeout, defaultFindExistingTimeout),
		resolveEntryWorkers(maxWorkers, defaultFindExistingMaxWorkers),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = closeProbeInfosForEntry(infos, "")
	}()

	filtered := make([]*internalbrowser.ProbeInfo, 0, len(infos))
	for _, info := range infos {
		if info != nil && info.IsAttachable() {
			filtered = append(filtered, info)
		}
	}
	return clonePublicProbeInfos(filtered), nil
}

// FindExistingBrowsersByProcess 按进程特征发现可接管的 Firefox 浏览器。
func FindExistingBrowsersByProcess(
	host string,
	timeout time.Duration,
	maxWorkers int,
) ([]ProbeInfo, error) {
	infos, err := findExistingBrowsersByProcessForEntry(
		resolveEntryHost(host),
		resolveEntryTimeout(timeout, defaultFindExistingTimeout),
		resolveEntryWorkers(maxWorkers, defaultAttachProcessMaxWorkers),
		false,
	)
	if err != nil {
		return nil, err
	}
	return clonePublicProbeInfos(infos), nil
}

// AutoAttachExistingBrowserByProcess 按进程特征自动探测并接管已启动浏览器。
func AutoAttachExistingBrowserByProcess(
	host string,
	timeout time.Duration,
	maxWorkers int,
	tabIndex int,
	latestTab bool,
) (*FirefoxPage, error) {
	candidatePorts, err := findCandidatePortsFromProcessForEntry()
	if err != nil {
		return nil, err
	}
	if len(candidatePorts) == 0 {
		return nil, support.NewBrowserConnectError(
			"未从进程特征中发现 Firefox 调试端口，请确认浏览器已启动并启用 --remote-debugging-port。",
			nil,
		)
	}

	resolvedHost := resolveEntryHost(host)
	resolvedTimeout := resolveEntryTimeout(timeout, defaultAutoAttachTimeout)
	resolvedWorkers := resolveEntryWorkers(maxWorkers, defaultAttachProcessMaxWorkers)
	resolvedTabIndex := resolveAttachTabIndex(tabIndex)

	infos, err := findExistingBrowsersByProcessForEntry(
		resolvedHost,
		resolvedTimeout,
		resolvedWorkers,
		true,
	)
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		occupiedInfos := make([]*internalbrowser.ProbeInfo, 0, len(candidatePorts))
		for _, port := range candidatePorts {
			info, probeErr := probeBiDiAddressForEntry(
				fmt.Sprintf("%s:%d", resolvedHost, port),
				resolvedTimeout,
				false,
			)
			if probeErr != nil || info == nil || !info.IsOccupied() {
				continue
			}
			occupiedInfos = append(occupiedInfos, info)
		}
		if len(occupiedInfos) > 0 {
			details := make([]string, 0, len(occupiedInfos))
			for _, info := range occupiedInfos {
				details = append(details, formatOccupiedProbeDetail(info))
			}
			return nil, buildAutoAttachError(
				"已发现 Firefox 调试端口，但其唯一 BiDi session 已被占用，当前无法接管。",
				details,
			)
		}
		return nil, support.NewBrowserConnectError(
			"已发现候选调试端口，但未检测到可接管的 Firefox BiDi 会话。",
			nil,
		)
	}

	errorsList := make([]string, 0, len(infos))
	for _, info := range infos {
		if info == nil {
			continue
		}
		page, pageErr := pageFromLiveProbeInfo(info, resolvedTabIndex, latestTab)
		if pageErr == nil {
			_ = closeProbeInfosForEntry(infos, info.Address)
			return page, nil
		}
		errorsList = append(errorsList, fmt.Sprintf("%s -> %v", info.Address, pageErr))
	}

	_ = closeProbeInfosForEntry(infos, "")
	return nil, buildAutoAttachError(
		"按进程特征发现了候选端口，但未能完成接管。",
		errorsList,
	)
}

// FindCandidatePortsFromProcess 按进程特征返回候选监听端口，但不做 BiDi 探测。
func FindCandidatePortsFromProcess() ([]int, error) {
	ports, err := findCandidatePortsFromProcessForEntry()
	if err != nil {
		return nil, err
	}
	return cloneProbePorts(ports), nil
}

func pageFromProbeInfo(info *internalbrowser.ProbeInfo, tabIndex int, latestTab bool) (*FirefoxPage, error) {
	if info == nil {
		return nil, support.NewBrowserConnectError("探测结果为空，无法接管浏览器。", nil)
	}

	firefox, err := createFirefoxFromProbeInfoForEntry(info)
	if err != nil {
		return nil, err
	}

	contextID := resolveAttachContextID(firefox.TabIDs(), tabIndex, latestTab)
	if contextID == "" {
		contextID, err = firefox.NewTab("", false)
		if err != nil {
			return nil, err
		}
	}

	firefoxPageRegistryMu.Lock()
	existing := firefoxPageRegistry[firefox.Address()]
	firefoxPageRegistryMu.Unlock()
	if existing != nil {
		existing.purgeClosedTabs(firefox.TabIDs())
		existing.mu.Lock()
		existing.browser = newFirefoxFromInner(firefox)
		existing.mu.Unlock()
		if err := selectFirefoxPageContext(existing, contextID); err != nil {
			return nil, err
		}
		return existing, nil
	}

	innerPage, err := newFirefoxPageFromBrowserForEntry(firefox, contextID)
	if err != nil {
		return nil, err
	}
	page := newFirefoxPageFromInner(innerPage, firefox.Address())

	firefoxPageRegistryMu.Lock()
	firefoxPageRegistry[firefox.Address()] = page
	firefoxPageRegistryMu.Unlock()
	return page, nil
}

func selectFirefoxPageContext(page *FirefoxPage, contextID string) error {
	if page == nil || page.inner == nil || contextID == "" {
		return nil
	}

	if browser := page.Browser(); browser != nil && browser.inner != nil {
		if err := browser.inner.ActivateTab(contextID); err != nil {
			return err
		}
	}
	return page.inner.SetContextID(contextID)
}

func resolveAttachContextID(tabIDs []string, tabIndex int, latestTab bool) string {
	if len(tabIDs) == 0 {
		return ""
	}
	if latestTab {
		return tabIDs[len(tabIDs)-1]
	}

	index := 0
	if tabIndex > 0 {
		index = tabIndex - 1
	}
	if index >= 0 && index < len(tabIDs) {
		return tabIDs[index]
	}
	return tabIDs[0]
}

func resolveAttachAddress(address string) string {
	if address == "" {
		return DefaultAddress
	}
	return address
}

func resolveAttachTabIndex(tabIndex int) int {
	if tabIndex <= 0 {
		return defaultAttachTabIndex
	}
	return tabIndex
}

func resolveEntryHost(host string) string {
	if host == "" {
		return DefaultHost
	}
	return host
}

func resolveStartPort(port int) int {
	if port <= 0 {
		return DefaultPort
	}
	return port
}

func resolveEndPort(port int, fallback int) int {
	if port <= 0 {
		return fallback
	}
	return port
}

func resolveEntryTimeout(timeout time.Duration, fallback time.Duration) time.Duration {
	if timeout <= 0 {
		return fallback
	}
	return timeout
}

func resolveEntryWorkers(workers int, fallback int) int {
	if workers <= 0 {
		return fallback
	}
	return workers
}

func formatOccupiedProbeDetail(info *internalbrowser.ProbeInfo) string {
	if info == nil {
		return ""
	}

	switch {
	case info.StatusMessage != "" && info.ErrorMessage != "":
		return fmt.Sprintf("%s -> %s (%s)", info.Address, info.StatusMessage, info.ErrorMessage)
	case info.StatusMessage != "":
		return fmt.Sprintf("%s -> %s", info.Address, info.StatusMessage)
	case info.ErrorMessage != "":
		return fmt.Sprintf("%s -> %s", info.Address, info.ErrorMessage)
	default:
		return fmt.Sprintf("%s -> Firefox BiDi 会话已被占用", info.Address)
	}
}

func buildAutoAttachError(summary string, details []string) error {
	if len(details) == 0 {
		return support.NewBrowserConnectError(summary, nil)
	}

	limit := 3
	if len(details) < limit {
		limit = len(details)
	}
	return support.NewBrowserConnectError(
		fmt.Sprintf("%s 失败详情: %s", summary, joinProbeDetails(details[:limit])),
		nil,
	)
}

func joinProbeDetails(details []string) string {
	switch len(details) {
	case 0:
		return ""
	case 1:
		return details[0]
	}

	result := details[0]
	for _, detail := range details[1:] {
		result += "；" + detail
	}
	return result
}
