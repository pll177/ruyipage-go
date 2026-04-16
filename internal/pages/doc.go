// Package pages 提供 FirefoxBase、FirefoxPage、FirefoxTab、FirefoxFrame 的内部实现承接。
// 本包只承接公开页面对象背后的状态与行为编排，不放公开根包导出、raw BiDi 命令或 Firefox 进程接入。
// 本包可以依赖 internal/units、internal/browser、internal/bidi、internal/base、internal/config、internal/support 与标准库；与 elements 的共享契约必须下沉到 base 或 support。
package pages
