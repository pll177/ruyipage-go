// Package elements 提供 FirefoxElement、StaticElement、NoneElement 的内部实现承接。
// 本包只承接元素对象行为与状态封装，不放公开根包导出、raw BiDi 命令或高层 manager 实现。
// 本包可以依赖 internal/units、internal/browser、internal/bidi、internal/base、internal/config、internal/support 与标准库；与 pages 的共享契约必须下沉到 base 或 support。
package elements
