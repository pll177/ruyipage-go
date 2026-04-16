// Package units 提供高层 manager 与组合单元。
// 本包固定承接 actions、console、contexts、downloads、emulation、events、extensions、listener、network tools、realm tracker、waiter、config manager 等高层能力，不放 raw command 封装。
// 本包可以依赖 internal/browser、internal/adapter、internal/bidi、internal/base、internal/config、internal/support 与标准库，但不得依赖根包；需同时服务 pages 与 elements 时应面向 base 契约避免包环。
package units
