package ruyipage

import internalunits "github.com/pll177/ruyipage-go/internal/units"

type (
	// ContextInfo 表示单个 browsingContext 快照。
	ContextInfo = internalunits.ContextInfo
	// ContextTree 表示 browsingContext.getTree 的高层结果。
	ContextTree = internalunits.ContextTree
	// ContextManager 表示 browsingContext / user context / client window 的高层管理器。
	ContextManager = internalunits.ContextManager
	// ExtensionManager 表示 WebExtension 安装与卸载管理器。
	ExtensionManager = internalunits.ExtensionManager
	// EmulationManager 表示环境与设备仿真管理器。
	EmulationManager = internalunits.EmulationManager
)

// Contexts 返回上下文管理器。
func (p *FirefoxBase) Contexts() *ContextManager {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Contexts()
}

// Extensions 返回扩展管理器。
func (p *FirefoxBase) Extensions() *ExtensionManager {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Extensions()
}

// Emulation 返回仿真管理器。
func (p *FirefoxBase) Emulation() *EmulationManager {
	if p == nil || p.inner == nil {
		return nil
	}
	return p.inner.Emulation()
}
