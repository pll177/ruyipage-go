package ruyipage

import internalbrowser "github.com/pll177/ruyipage-go/internal/browser"

// ProbeContextInfo 表示单个 live probe 上下文快照。
type ProbeContextInfo struct {
	Context        string
	URL            string
	UserContext    string
	OriginalOpener any
}

// ProbeInfo 表示扫描 / 探测到的可接管浏览器信息。
type ProbeInfo struct {
	Address       string
	Host          string
	Port          int
	Ready         bool
	Message       string
	StatusMessage string
	ErrorMessage  string
	ProbeState    string
	WSURL         string
	SessionID     string
	SessionOwned  bool
	WindowCount   int
	TabCount      int
	ClientWindows []map[string]any
	Contexts      []ProbeContextInfo
	ScannedPorts  []int
}

// IsAttachable 返回当前探测结果是否可直接接管。
func (info ProbeInfo) IsAttachable() bool {
	return info.ProbeState == string(internalbrowser.ProbeStateAttachable)
}

// IsOccupied 返回当前探测结果是否为“端口有效但唯一 session 已被占用”。
func (info ProbeInfo) IsOccupied() bool {
	return info.ProbeState == string(internalbrowser.ProbeStateOccupied)
}

func clonePublicProbeInfos(infos []*internalbrowser.ProbeInfo) []ProbeInfo {
	if len(infos) == 0 {
		return []ProbeInfo{}
	}

	cloned := make([]ProbeInfo, 0, len(infos))
	for _, info := range infos {
		if info == nil {
			continue
		}
		cloned = append(cloned, clonePublicProbeInfo(info))
	}
	return cloned
}

func clonePublicProbeInfo(info *internalbrowser.ProbeInfo) ProbeInfo {
	if info == nil {
		return ProbeInfo{}
	}

	return ProbeInfo{
		Address:       info.Address,
		Host:          info.Host,
		Port:          info.Port,
		Ready:         info.Ready,
		Message:       info.Message,
		StatusMessage: info.StatusMessage,
		ErrorMessage:  info.ErrorMessage,
		ProbeState:    string(info.ProbeState),
		WSURL:         info.WSURL,
		SessionID:     info.SessionID,
		SessionOwned:  info.SessionOwned,
		WindowCount:   info.WindowCount,
		TabCount:      info.TabCount,
		ClientWindows: cloneProbeClientWindows(info.ClientWindows),
		Contexts:      cloneProbeContexts(info.Contexts),
		ScannedPorts:  cloneProbePorts(info.ScannedPorts),
	}
}

func cloneProbeContexts(contexts []internalbrowser.ProbeContextInfo) []ProbeContextInfo {
	if len(contexts) == 0 {
		return []ProbeContextInfo{}
	}

	cloned := make([]ProbeContextInfo, 0, len(contexts))
	for _, context := range contexts {
		cloned = append(cloned, ProbeContextInfo{
			Context:        context.Context,
			URL:            context.URL,
			UserContext:    context.UserContext,
			OriginalOpener: context.OriginalOpener,
		})
	}
	return cloned
}

func cloneProbeClientWindows(windows []map[string]any) []map[string]any {
	if len(windows) == 0 {
		return []map[string]any{}
	}

	cloned := make([]map[string]any, len(windows))
	for index, window := range windows {
		row := make(map[string]any, len(window))
		for key, value := range window {
			row[key] = value
		}
		cloned[index] = row
	}
	return cloned
}

func cloneProbePorts(ports []int) []int {
	if len(ports) == 0 {
		return []int{}
	}

	cloned := make([]int, len(ports))
	copy(cloned, ports)
	return cloned
}
