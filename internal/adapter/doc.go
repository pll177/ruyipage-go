// Package adapter 提供 Firefox 外围接入与桥接能力。
// 本包固定承接 remote agent、marionette、pref、context bridge 等适配实现。
// 本包只允许依赖 internal/config、internal/base、internal/support 与标准库，不承接公开 API、raw BiDi 命令或高层 manager。
package adapter
