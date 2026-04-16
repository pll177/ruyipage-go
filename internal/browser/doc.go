// Package browser 提供 Firefox 生命周期、attach、probe、scan 与连接复用。
// 本包固定承接启动、连接、退出、按端口或进程探测以及已打开浏览器接管流程。
// 本包可以依赖 internal/adapter、internal/bidi、internal/base、internal/config、internal/support 与标准库，不承接公开 page/element 类型或高层 manager。
package browser
