// Package ruyipage 是 ruyiPage Go 版唯一公开根包。
// 根包只暴露公共类型、入口函数、公共小类型、常量和错误公开面。
// transport、dispatcher、raw BiDi 命令、Firefox 生命周期、manager 与工具实现固定落在 internal/*。
// 后续根包只允许承接已冻结的公开文件族，不新增公开子包，也不把 internal 实现上浮到根包。
package ruyipage
