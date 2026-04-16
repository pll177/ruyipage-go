// Package bidi 提供 Firefox BiDi 原始协议命令封装。
// 本包固定承接 session、browsingContext、script、input、network 等命令与参数结果结构。
// 本包只允许依赖 internal/base、internal/support 与标准库，不承接浏览器生命周期、公开页面对象或高层 manager。
package bidi
