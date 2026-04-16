// Package base 提供 transport、dispatcher、driver、event emitter 与跨层共享基础契约。
// 本包只承接基础设施，不放原始 BiDi 命令、Firefox 生命周期或高层 manager。
// 本包只允许依赖 internal/support 与标准库；pages 与 elements 的共享契约也必须下沉到这里避免包环。
package base
