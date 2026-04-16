package ruyipage

import internalunits "github.com/pll177/ruyipage-go/internal/units"

type (
	// LogEntry 表示单条控制台 / javascript 日志。
	LogEntry = internalunits.LogEntry
	// ConsoleListener 表示页面级控制台日志监听器。
	ConsoleListener = internalunits.ConsoleListener
	// DownloadEvent 表示单条下载事件快照。
	DownloadEvent = internalunits.DownloadEvent
	// DownloadsManager 表示页面级下载管理器。
	DownloadsManager = internalunits.DownloadsManager
	// BidiEvent 表示通用 BiDi 事件快照。
	BidiEvent = internalunits.BidiEvent
	// EventTracker 表示通用 BiDi 事件跟踪器。
	EventTracker = internalunits.EventTracker
	// RealmTracker 表示 realm 生命周期跟踪器。
	RealmTracker = internalunits.RealmTracker
)

// Console 返回页面级控制台日志监听器。
func (p *FirefoxBase) Console() *ConsoleListener {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Console()
}

// Downloads 返回页面级下载管理器。
func (p *FirefoxBase) Downloads() *DownloadsManager {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Downloads()
}

// Events 返回通用 BiDi 事件跟踪器。
func (p *FirefoxBase) Events() *EventTracker {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Events()
}

// Realms 返回 realm 生命周期跟踪器。
func (p *FirefoxBase) Realms() *RealmTracker {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Realms()
}
