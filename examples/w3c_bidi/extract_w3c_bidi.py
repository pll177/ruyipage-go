#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""从 W3C WebDriver BiDi 规范网页提取命令与事件目录。"""

from __future__ import annotations

import io
import json
import re
import sys
from pathlib import Path

if sys.platform == "win32":
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")

from ruyipage import FirefoxPage


SCRIPT_DIR = Path(__file__).resolve().parent
OUTPUT_PATH = SCRIPT_DIR / "w3c_bidi_apis.json"


def main() -> None:
    page = FirefoxPage()
    try:
        print("正在访问 W3C WebDriver BiDi 规范...")
        page.get("https://w3c.github.io/webdriver-bidi/")
        page.wait(3)

        toc_items = page.run_js(
            """
            (() => {
                const items = [];
                document.querySelectorAll("ol.toc li a").forEach((link) => {
                    const href = link.getAttribute("href");
                    const text = link.textContent.trim();
                    if (href && text) {
                        items.push({
                            href,
                            text,
                            id: href.replace("#", "")
                        });
                    }
                });
                return items;
            })()
            """
        )

        commands: dict[str, list[str]] = {}
        events: dict[str, list[str]] = {}
        modules: set[str] = set()

        for item in toc_items:
            text = item["text"]
            cmd_match = re.search(r"The\\s+(\\w+)\\.(\\w+)\\s+Command", text)
            if cmd_match:
                module = cmd_match.group(1)
                command = f"{module}.{cmd_match.group(2)}"
                modules.add(module)
                commands.setdefault(module, []).append(command)

            evt_match = re.search(r"The\\s+(\\w+)\\.(\\w+)\\s+Event", text)
            if evt_match:
                module = evt_match.group(1)
                event = f"{module}.{evt_match.group(2)}"
                modules.add(module)
                events.setdefault(module, []).append(event)

        result = {
            "modules": sorted(modules),
            "commands": {module: sorted(items) for module, items in commands.items()},
            "events": {module: sorted(items) for module, items in events.items()},
            "total_commands": sum(len(items) for items in commands.values()),
            "total_events": sum(len(items) for items in events.values()),
        }
        OUTPUT_PATH.write_text(json.dumps(result, indent=2, ensure_ascii=False), encoding="utf-8")
        print(f"已保存到: {OUTPUT_PATH}")
    finally:
        page.quit()


if __name__ == "__main__":
    main()
