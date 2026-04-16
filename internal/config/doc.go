// Package config 提供 FirefoxOptions 背后的内部配置链与默认值承接。
// 本包只放配置归一化、默认值和内部选项表示，不承载浏览器生命周期、page/element 或 raw BiDi 命令。
// 本包只允许依赖 internal/support 与标准库，供 browser、units、pages、elements 与根包向下复用。
package config
