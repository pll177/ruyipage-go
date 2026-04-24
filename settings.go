package ruyipage

import "github.com/pll177/ruyipage-go/internal/support"

type (
	// SettingsValues 表示全局默认行为与超时基线。
	// 具体字段含义见 internal/support/settings.go 中每个字段的中文说明。
	SettingsValues = support.SettingsValues
)

// Settings 是全局可变设置基线，建议在创建浏览器或页面对象前统一配置。
var Settings = support.Settings
