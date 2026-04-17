#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""基于 examples/w3c_bidi/w3c_bidi_apis.json 生成简要对比表。"""

from __future__ import annotations

import json
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
SOURCE_PATH = SCRIPT_DIR / "w3c_bidi_apis.json"
OUTPUT_PATH = SCRIPT_DIR / "w3c_bidi_comparison.md"


RUYIPAGE_GO_COMMANDS = {
    "browser": 7,
    "browsingContext": 13,
    "emulation": 11,
    "input": 3,
    "network": 13,
    "script": 6,
    "session": 5,
    "storage": 3,
    "webExtension": 2,
}

RUYIPAGE_GO_EVENTS = {
    "browsingContext": 14,
    "input": 1,
    "log": 1,
    "network": 5,
    "script": 3,
}


def main() -> None:
    data = json.loads(SOURCE_PATH.read_text(encoding="utf-8"))
    lines: list[str] = [
        "# WebDriver BiDi API 对比摘要",
        "",
        f"- W3C 命令数: {data['total_commands']}",
        f"- W3C 事件数: {data['total_events']}",
        "",
        "| 模块 | W3C 命令 | Go 参考命令 | W3C 事件 | Go 参考事件 |",
        "| --- | ---: | ---: | ---: | ---: |",
    ]

    modules = sorted(set(data["modules"]) | set(RUYIPAGE_GO_COMMANDS) | set(RUYIPAGE_GO_EVENTS))
    for module in modules:
        w3c_commands = len(data["commands"].get(module, []))
        go_commands = RUYIPAGE_GO_COMMANDS.get(module, 0)
        w3c_events = len(data["events"].get(module, []))
        go_events = RUYIPAGE_GO_EVENTS.get(module, 0)
        lines.append(f"| {module} | {w3c_commands} | {go_commands} | {w3c_events} | {go_events} |")

    OUTPUT_PATH.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"已生成: {OUTPUT_PATH}")


if __name__ == "__main__":
    main()
