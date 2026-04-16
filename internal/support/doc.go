// Package support 提供 internal 层通用支持代码。
// 本包固定承接 locator、tools、web、cookies、BiDi 值转换、settings 等公共辅助。
// 本包位于内部依赖最底层，只允许依赖标准库，不反向依赖 config、base、bidi、adapter、browser、units、pages、elements 或根包。
package support
