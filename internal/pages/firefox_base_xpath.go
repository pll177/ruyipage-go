package pages

const textFindFunction = `(text, rootNode) => {
    const root = rootNode || document.body || document.documentElement;
    const results = [];
    const skip = new Set(["SCRIPT", "STYLE", "NOSCRIPT", "META", "LINK"]);
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_ELEMENT);
    let node = root;
    if (node && node.nodeType === Node.ELEMENT_NODE && !skip.has(node.tagName)) {
        const content = (node.textContent || "").trim();
        if (content.includes(text)) {
            results.push(node);
        }
    }
    while ((node = walker.nextNode())) {
        if (!node || skip.has(node.tagName)) {
            continue;
        }
        const content = (node.textContent || "").trim();
        if (content.includes(text)) {
            results.push(node);
        }
    }
    return results;
}`

const xpathPickerBridgeScript = `(source) => {
    if (typeof window === "undefined" || window.top !== window) {
        return false;
    }

    const injectWindow = (targetWindow) => {
        try {
            if (!targetWindow || typeof targetWindow.eval !== "function") {
                return;
            }
            targetWindow.eval("(" + source + ")()");
        } catch (error) {
        }
    };

    const bindFrame = (frame) => {
        if (!frame || frame.__ruyiXPathPickerBound) {
            return;
        }
        frame.__ruyiXPathPickerBound = true;

        const inject = () => {
            try {
                const childWindow = frame.contentWindow;
                if (!childWindow || childWindow === window) {
                    return;
                }
                injectWindow(childWindow);
                scanWindow(childWindow);
            } catch (error) {
            }
        };

        const tryInjectReady = () => {
            let attempts = 0;
            const timer = window.setInterval(() => {
                attempts += 1;
                try {
                    const childWindow = frame.contentWindow;
                    const childDocument = childWindow && childWindow.document;
                    if (childWindow && childDocument && childDocument.readyState !== "loading") {
                        inject();
                        window.clearInterval(timer);
                        return;
                    }
                } catch (error) {
                    window.clearInterval(timer);
                    return;
                }
                if (attempts >= 20) {
                    window.clearInterval(timer);
                }
            }, 150);
        };

        frame.addEventListener("load", inject);
        inject();
        tryInjectReady();
    };

    const observeWindow = (targetWindow) => {
        try {
            const targetDocument = targetWindow.document;
            if (!targetDocument || targetDocument.__ruyiXPathPickerObserverBound) {
                return;
            }
            let pending = false;
            const observer = new MutationObserver(() => {
                if (pending) {
                    return;
                }
                pending = true;
                targetWindow.setTimeout(() => {
                    pending = false;
                    scanWindow(targetWindow);
                }, 50);
            });
            observer.observe(targetDocument.documentElement || targetDocument, {
                childList: true,
                subtree: true,
            });
            targetDocument.__ruyiXPathPickerObserverBound = true;
        } catch (error) {
        }
    };

    const scanWindow = (targetWindow) => {
        try {
            const frames = Array.from(targetWindow.document.querySelectorAll("iframe"));
            frames.forEach(bindFrame);
            observeWindow(targetWindow);
        } catch (error) {
        }
    };

    window.__ruyiXPathPickerInjectIntoFrames = () => scanWindow(window);
    scanWindow(window);
    return true;
}`

const xpathPickerScript = `() => {
    if (typeof window === "undefined" || typeof document === "undefined") {
        return false;
    }

    const PANEL_ID = "__ruyi_xpath_picker_panel__";
    const HIGHLIGHT_ID = "__ruyi_xpath_picker_highlight__";

    let isTopWindow = false;
    let topWindowRef = window;
    try {
        isTopWindow = window.top === window;
        topWindowRef = isTopWindow ? window : window.top;
    } catch (error) {
        isTopWindow = true;
        topWindowRef = window;
    }

    const state = topWindowRef.__ruyiXPathPicker__ || {
        mode: "idle",
        collapsed: false,
        activeTab: "info",
        hoverData: null,
        selectedData: null,
        panel: null,
        watchdogBound: false,
    };
    if (!state.group || typeof state.group !== "object") {
        state.group = {
            items: [],
            candidates: [],
            selectedKey: "",
            anchorName: "",
            statusText: "",
            commonXPath: "",
            commonCss: "",
            code: "",
            diagnostics: [],
        };
    }
    topWindowRef.__ruyiXPathPicker__ = state;

    const localState = window.__ruyiXPathPickerLocal__ || {
        hoverElement: null,
        selectedElement: null,
        highlight: null,
        handlersBound: false,
        boundDocument: null,
        moveHandler: null,
        clickHandler: null,
        scrollHandler: null,
        resizeHandler: null,
    };
    window.__ruyiXPathPickerLocal__ = localState;

    function normalizeText(text) {
        return String(text || "").replace(/\s+/g, " ").trim();
    }

    function escapeHTML(text) {
        return String(text || "")
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;");
    }

    function escapeAttribute(text) {
        return escapeHTML(text).replace(/"/g, "&quot;");
    }

    function quoteCode(value) {
        return JSON.stringify(String(value || ""));
    }

    function escapeCSSValue(text) {
        return '"' + String(text || "")
            .replace(/\\/g, "\\\\")
            .replace(/"/g, '\\"') + '"';
    }

    function isElementNode(node) {
        return !!node && node.nodeType === 1 && typeof node.tagName === "string";
    }

    function isShadowRootNode(root) {
        return !!root && root.nodeType === 11 && !!root.host && isElementNode(root.host);
    }

    function countCSSMatches(doc, selector) {
        try {
            return doc.querySelectorAll(selector).length;
        } catch (error) {
            return Number.POSITIVE_INFINITY;
        }
    }

    function ensureStyles() {
        if (document.getElementById("__ruyi_xpath_picker_style__")) {
            return;
        }
        const style = document.createElement("style");
        style.id = "__ruyi_xpath_picker_style__";
        style.textContent =
            "#" + PANEL_ID + "{position:fixed;right:16px;bottom:16px;width:min(340px,calc(100vw - 24px));max-height:min(70vh,560px);overflow:auto;padding:14px;border-radius:16px;border:1px solid rgba(255,255,255,.16);background:rgba(15,23,42,.62);color:#e5eefb;box-shadow:0 18px 42px rgba(2,6,23,.34);backdrop-filter:blur(16px) saturate(140%);-webkit-backdrop-filter:blur(16px) saturate(140%);font:12px/1.5 Inter,'Segoe UI',Arial,sans-serif;z-index:2147483647;transition:width .18s ease,padding .18s ease,transform .18s ease;}" +
            "#" + PANEL_ID + "[data-collapsed='true']{width:auto;max-height:none;overflow:visible;padding:10px 12px;border-radius:999px;}" +
            "#" + PANEL_ID + " *{box-sizing:border-box;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__header{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:12px;}" +
            "#" + PANEL_ID + "[data-collapsed='true'] .ruyi-xpath-picker__header{margin-bottom:0;gap:10px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__title{font-size:13px;font-weight:700;letter-spacing:.02em;color:#f8fafc;}" +
            "#" + PANEL_ID + "[data-collapsed='true'] .ruyi-xpath-picker__title{font-size:12px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__badge{display:inline-flex;align-items:center;padding:3px 8px;border-radius:999px;background:rgba(96,165,250,.18);color:#bfdbfe;font-size:11px;font-weight:600;white-space:nowrap;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__intro{margin:0 0 12px;color:rgba(226,232,240,.82);}" +
            "#" + PANEL_ID + "[data-collapsed='true'] .ruyi-xpath-picker__intro," +
            "#" + PANEL_ID + "[data-collapsed='true'] .ruyi-xpath-picker__meta," +
            "#" + PANEL_ID + "[data-collapsed='true'] .ruyi-xpath-picker__actions{display:none;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__tabs{display:flex;gap:8px;margin-bottom:12px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__tab{appearance:none;border:1px solid rgba(148,163,184,.18);border-radius:999px;padding:6px 10px;background:rgba(15,23,42,.24);color:#cbd5e1;cursor:pointer;font-size:11px;font-weight:700;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__tab[data-active='true']{background:rgba(59,130,246,.22);color:#eff6ff;border-color:rgba(96,165,250,.35);}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__meta{display:grid;gap:10px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__field{padding:10px 11px;border-radius:12px;background:rgba(15,23,42,.34);border:1px solid rgba(148,163,184,.14);}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__field-header{display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:6px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__label{display:block;margin-bottom:4px;color:#93c5fd;font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:.04em;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__value{color:#f8fafc;word-break:break-word;white-space:pre-wrap;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__value[data-code='true']{font-family:Consolas,'SFMono-Regular',monospace;font-size:11px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__code-block{padding:12px;border-radius:12px;background:rgba(2,6,23,.5);border:1px solid rgba(148,163,184,.16);color:#e2e8f0;white-space:pre-wrap;word-break:break-word;font-family:Consolas,'SFMono-Regular',monospace;font-size:11px;line-height:1.6;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__hint{margin:0;color:rgba(191,219,254,.88);font-size:11px;line-height:1.6;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__copy{appearance:none;border:1px solid rgba(148,163,184,.18);border-radius:999px;padding:4px 8px;background:rgba(148,163,184,.12);color:#cbd5e1;cursor:pointer;font-size:10px;font-weight:700;line-height:1;white-space:nowrap;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__copy[data-copied='true']{background:rgba(34,197,94,.18);color:#dcfce7;border-color:rgba(34,197,94,.28);}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__actions{display:flex;gap:8px;margin-top:14px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__header-actions{display:inline-flex;align-items:center;gap:8px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button{appearance:none;border:0;border-radius:10px;padding:9px 12px;cursor:pointer;font-size:12px;font-weight:700;transition:transform .16s ease,background .16s ease,opacity .16s ease;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button:hover{transform:translateY(-1px);}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button[disabled]{opacity:.45;cursor:not-allowed;transform:none;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button--primary{background:rgba(59,130,246,.92);color:#eff6ff;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button--secondary{background:rgba(15,23,42,.26);color:#e2e8f0;border:1px solid rgba(148,163,184,.18);}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button--tiny{padding:6px 9px;font-size:11px;border-radius:999px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button--ghost," +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button--icon{background:rgba(148,163,184,.18);color:#e2e8f0;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__button--icon{min-width:34px;padding:8px 10px;border-radius:999px;line-height:1;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__group-list{display:grid;gap:8px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__group-item{display:flex;align-items:flex-start;justify-content:space-between;gap:10px;padding:10px 11px;border-radius:12px;background:rgba(15,23,42,.34);border:1px solid rgba(148,163,184,.14);}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__group-item-main{min-width:0;flex:1;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__group-item-title{font-weight:700;color:#f8fafc;word-break:break-word;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__group-item-sub{margin-top:4px;color:#cbd5e1;font-size:11px;word-break:break-word;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__strategy-list{display:grid;gap:8px;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__strategy-item{padding:10px 11px;border-radius:12px;background:rgba(15,23,42,.34);border:1px solid rgba(148,163,184,.14);cursor:pointer;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__strategy-item[data-active='true']{border-color:rgba(96,165,250,.55);background:rgba(30,41,59,.72);box-shadow:0 0 0 1px rgba(96,165,250,.18) inset;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__strategy-title{font-weight:700;color:#f8fafc;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__strategy-sub{margin-top:4px;color:#cbd5e1;font-size:11px;word-break:break-word;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__strategy-meta{display:flex;gap:8px;flex-wrap:wrap;margin-top:6px;font-size:10px;color:#93c5fd;}" +
            "#" + HIGHLIGHT_ID + "{position:absolute;display:none;border-radius:12px;border:2px solid rgba(96,165,250,.95);background:rgba(96,165,250,.12);box-shadow:0 0 0 1px rgba(255,255,255,.28),0 10px 30px rgba(37,99,235,.18);pointer-events:none;z-index:2147483646;}" +
            "@media (max-width:640px){" +
            "#" + PANEL_ID + "{right:12px;bottom:12px;width:calc(100vw - 24px);max-height:58vh;}" +
            "#" + PANEL_ID + " .ruyi-xpath-picker__actions{flex-direction:column;}" +
            "}";
        (document.head || document.documentElement).appendChild(style);
    }

    function ensurePanel() {
        if (!isTopWindow) {
            return null;
        }
        let panel = document.getElementById(PANEL_ID);
        if (!panel) {
            panel = document.createElement("div");
            panel.id = PANEL_ID;
            panel.setAttribute("data-collapsed", "false");
            panel.innerHTML = [
                '<div class="ruyi-xpath-picker__header">',
                '  <div class="ruyi-xpath-picker__title">XPath Picker</div>',
                '  <div class="ruyi-xpath-picker__header-actions">',
                '    <div class="ruyi-xpath-picker__badge" data-role="status">待选择</div>',
                '    <button type="button" class="ruyi-xpath-picker__button ruyi-xpath-picker__button--icon" data-action="toggle" aria-label="收起 XPath Picker">-</button>',
                "  </div>",
                "</div>",
                '<p class="ruyi-xpath-picker__intro" data-role="intro">移动鼠标可预览目标，点击后锁定当前元素。</p>',
                '<div class="ruyi-xpath-picker__tabs">',
                '  <button type="button" class="ruyi-xpath-picker__tab" data-tab="info" data-active="true">Info</button>',
                '  <button type="button" class="ruyi-xpath-picker__tab" data-tab="ruyipage" data-active="false">ruyiPage代码生成</button>',
                '  <button type="button" class="ruyi-xpath-picker__tab" data-tab="group" data-active="false">元素组</button>',
                "</div>",
                '<div class="ruyi-xpath-picker__meta" data-role="meta"></div>',
                '<div class="ruyi-xpath-picker__actions">',
                '  <button type="button" class="ruyi-xpath-picker__button ruyi-xpath-picker__button--primary" data-action="unlock" disabled>继续选择</button>',
                '  <button type="button" class="ruyi-xpath-picker__button ruyi-xpath-picker__button--secondary" data-action="capture-group" disabled>捕获相似元素</button>',
                '  <button type="button" class="ruyi-xpath-picker__button ruyi-xpath-picker__button--secondary" data-action="pause">暂停选择</button>',
                '  <button type="button" class="ruyi-xpath-picker__button ruyi-xpath-picker__button--ghost" data-action="toggle">收起</button>',
                "</div>",
            ].join("");
            document.documentElement.appendChild(panel);

            const unlockButton = panel.querySelector('[data-action="unlock"]');
            const captureGroupButton = panel.querySelector('[data-action="capture-group"]');
            const pauseButton = panel.querySelector('[data-action="pause"]');
            const toggleButtons = panel.querySelectorAll('[data-action="toggle"]');
            const tabs = panel.querySelectorAll("[data-tab]");

            if (unlockButton) {
                unlockButton.addEventListener("click", (event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    unlockSelection();
                });
            }
            if (pauseButton) {
                pauseButton.addEventListener("click", (event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    togglePaused();
                });
            }
            if (captureGroupButton) {
                captureGroupButton.addEventListener("click", (event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    captureGroup();
                });
            }
            toggleButtons.forEach((button) => {
                button.addEventListener("click", (event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    toggleCollapsed();
                });
            });
            tabs.forEach((button) => {
                button.addEventListener("click", (event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    state.activeTab = button.getAttribute("data-tab") || "info";
                    syncTopUI();
                });
            });
            panel.addEventListener("click", (event) => {
                const strategyButton = event.target && event.target.closest ? event.target.closest("[data-group-strategy-key]") : null;
                if (strategyButton) {
                    event.preventDefault();
                    event.stopPropagation();
                    selectGroupStrategy(strategyButton.getAttribute("data-group-strategy-key") || "");
                    return;
                }
                const removeButton = event.target && event.target.closest ? event.target.closest("[data-remove-group-index]") : null;
                if (removeButton) {
                    event.preventDefault();
                    event.stopPropagation();
                    removeGroupItem(Number.parseInt(removeButton.getAttribute("data-remove-group-index") || "-1", 10));
                    return;
                }
                const clearGroupButton = event.target && event.target.closest ? event.target.closest("[data-clear-group]") : null;
                if (clearGroupButton) {
                    event.preventDefault();
                    event.stopPropagation();
                    clearGroup();
                    return;
                }
                const copyButton = event.target && event.target.closest ? event.target.closest("[data-copy-value]") : null;
                if (!copyButton) {
                    return;
                }
                event.preventDefault();
                event.stopPropagation();
                copyText(copyButton.getAttribute("data-copy-value") || "", copyButton);
            });
        }
        state.panel = panel;
        return panel;
    }

    function ensureHighlight() {
        let highlight = document.getElementById(HIGHLIGHT_ID);
        if (!highlight) {
            highlight = document.createElement("div");
            highlight.id = HIGHLIGHT_ID;
            document.documentElement.appendChild(highlight);
        }
        localState.highlight = highlight;
        return highlight;
    }

    function getElementName(element) {
        if (!element || !element.tagName) {
            return "";
        }
        const tag = element.tagName.toLowerCase();
        const id = element.id ? "#" + element.id : "";
        const nameAttr = element.getAttribute("name");
        const ariaLabel = element.getAttribute("aria-label");
        const dataTestId = element.getAttribute("data-testid") || element.getAttribute("data-test") || element.getAttribute("data-qa");
        const className = typeof element.className === "string"
            ? element.className.trim().split(/\s+/).filter(Boolean).slice(0, 2).map((item) => "." + item).join("")
            : "";
        const hints = [nameAttr && ("name=" + nameAttr), ariaLabel && ("aria=" + ariaLabel), dataTestId && ("data=" + dataTestId)]
            .filter(Boolean)
            .slice(0, 1);
        return [tag + id + className].concat(hints).join(" ");
    }

    function getVisibleText(element) {
        if (!element) {
            return "";
        }
        const text = normalizeText(element.innerText || element.textContent || "");
        return text.length > 160 ? text.slice(0, 157) + "..." : text;
    }

    function getClosedShadowRoot(host) {
        if (!host || typeof window.__ruyiGetClosedShadowRoot !== "function") {
            return null;
        }
        try {
            return window.__ruyiGetClosedShadowRoot(host);
        } catch (error) {
            return null;
        }
    }

    function getShadowMode(root) {
        try {
            if (root && typeof root.mode === "string" && root.mode) {
                return root.mode;
            }
        } catch (error) {
        }
        return "open";
    }

    function getNextShadowRoot(element, currentRoot) {
        try {
            if (element && typeof element.getRootNode === "function") {
                const root = element.getRootNode();
                if (isShadowRootNode(root) && root !== currentRoot) {
                    return root;
                }
            }
        } catch (error) {
        }
        const closedRoot = getClosedShadowRoot(element);
        if (closedRoot && closedRoot !== currentRoot) {
            return closedRoot;
        }
        return null;
    }

    function getDeepestShadowElement(root, clientX, clientY) {
        if (!root || typeof root.elementFromPoint !== "function" || typeof clientX !== "number" || typeof clientY !== "number") {
            return null;
        }
        let currentRoot = root;
        let candidate = null;
        for (let depth = 0; depth < 8; depth += 1) {
            let next = null;
            try {
                next = currentRoot.elementFromPoint(clientX, clientY);
            } catch (error) {
                break;
            }
            if (!isElementNode(next)) {
                break;
            }
            candidate = next;
            const nestedRoot = getNextShadowRoot(next, currentRoot);
            if (!nestedRoot) {
                break;
            }
            currentRoot = nestedRoot;
        }
        return candidate;
    }

    function escapeXPathLiteral(value) {
        const text = String(value || "");
        if (!text.includes('"')) {
            return '"' + text + '"';
        }
        if (!text.includes("'")) {
            return "'" + text + "'";
        }
        return "concat(" + text.split('"').map((part, index, parts) => {
            const pieces = [];
            if (part) {
                pieces.push('"' + part + '"');
            }
            if (index < parts.length - 1) {
                pieces.push("'\"'");
            }
            return pieces.join(", ");
        }).filter(Boolean).join(", ") + ")";
    }

    function getXPathNodeName(element) {
        if (!element || !element.tagName) {
            return "*";
        }
        const tagName = element.tagName.toLowerCase();
        const namespace = element.namespaceURI || "";
        if (namespace && namespace !== "http://www.w3.org/1999/xhtml") {
            const localName = typeof element.localName === "string" ? element.localName : tagName;
            return "*[local-name()=" + escapeXPathLiteral(localName) + "]";
        }
        return tagName;
    }

    function getSiblingIndex(element) {
        let index = 1;
        let sibling = element.previousElementSibling;
        while (sibling) {
            if ((sibling.namespaceURI || "") === (element.namespaceURI || "") && sibling.tagName === element.tagName) {
                index += 1;
            }
            sibling = sibling.previousElementSibling;
        }
        return index;
    }

    function countMatches(doc, expr) {
        try {
            return doc.evaluate("count(" + expr + ")", doc, null, XPathResult.NUMBER_TYPE, null).numberValue;
        } catch (error) {
            return Number.POSITIVE_INFINITY;
        }
    }

    function buildSegmentWithIndex(element) {
        return getXPathNodeName(element) + "[" + getSiblingIndex(element) + "]";
    }

    function buildAncestorRelativeXPath(element, maxDepth) {
        const segments = [];
        let current = element;
        let depth = 0;
        while (current && current.nodeType === Node.ELEMENT_NODE && depth < maxDepth) {
            segments.unshift(buildSegmentWithIndex(current));
            const candidate = "//" + segments.join("/");
            if (countMatches(current.ownerDocument, candidate) === 1) {
                return candidate;
            }
            current = current.parentElement;
            depth += 1;
        }
        return "";
    }

    function getStableAttributeSelector(element) {
        const attrs = ["data-testid", "data-test", "data-qa", "name", "aria-label", "placeholder", "type", "role", "title"];
        for (const attr of attrs) {
            const value = normalizeText(element.getAttribute(attr));
            if (!value) {
                continue;
            }
            const expr = "//" + getXPathNodeName(element) + "[@" + attr + "=" + escapeXPathLiteral(value) + "]";
            if (countMatches(element.ownerDocument, expr) === 1) {
                return expr;
            }
        }
        return "";
    }

    function getAbsoluteXPath(element) {
        if (!element || element.nodeType !== Node.ELEMENT_NODE) {
            return "";
        }
        const segments = [];
        let current = element;
        while (current && current.nodeType === Node.ELEMENT_NODE) {
            segments.unshift(buildSegmentWithIndex(current));
            current = current.parentElement;
        }
        return "/" + segments.join("/");
    }

    function getRelativeXPath(element) {
        if (!element || element.nodeType !== Node.ELEMENT_NODE) {
            return "";
        }
        if (element.id) {
            return "//*[@id=" + escapeXPathLiteral(element.id) + "]";
        }
        const stableAttr = getStableAttributeSelector(element);
        if (stableAttr) {
            return stableAttr;
        }

        const ownText = normalizeText(Array.from(element.childNodes)
            .filter((node) => node.nodeType === Node.TEXT_NODE)
            .map((node) => node.textContent || "")
            .join(" "));
        if (ownText && ownText.length <= 80) {
            const expr = "//" + getXPathNodeName(element) + "[normalize-space(text())=" + escapeXPathLiteral(ownText) + "]";
            if (countMatches(element.ownerDocument, expr) === 1) {
                return expr;
            }
        }

        const ancestorRelative = buildAncestorRelativeXPath(element, 5);
        if (ancestorRelative) {
            return ancestorRelative;
        }

        return getAbsoluteXPath(element);
    }

    function getElementLocator(element) {
        if (!element || !element.tagName) {
            return "";
        }
        if (element.id) {
            return "#" + element.id;
        }
        const tag = element.tagName.toLowerCase();
        const attrs = ["data-testid", "data-test", "data-qa", "name", "aria-label", "title"];
        for (const attr of attrs) {
            const value = normalizeText(element.getAttribute(attr));
            if (!value) {
                continue;
            }
            const selector = tag + "[" + attr + "=" + escapeCSSValue(value) + "]";
            if (countCSSMatches(element.ownerDocument, selector) === 1) {
                return selector;
            }
        }
        const xpath = getRelativeXPath(element);
        return xpath ? "xpath:" + xpath : "";
    }

    function isGenericFrameLabel(label) {
        const normalized = normalizeText(label);
        return !normalized || normalized === "iframe" || normalized === "frame" || normalized === "cross-origin-frame";
    }

    function isSafeCSSID(label) {
        return /^[A-Za-z_][A-Za-z0-9_:-]*$/.test(String(label || ""));
    }

    function getFrameLocator(frame) {
        if (!isElementNode(frame)) {
            return "";
        }
        const tag = frame.tagName ? frame.tagName.toLowerCase() : "iframe";
        if (frame.id) {
            return "#" + frame.id;
        }

        const attrs = ["name", "title"];
        for (const attr of attrs) {
            const value = normalizeText(frame.getAttribute(attr));
            if (!value) {
                continue;
            }
            const selector = tag + "[" + attr + "=" + escapeCSSValue(value) + "]";
            if (countCSSMatches(frame.ownerDocument, selector) === 1) {
                return selector;
            }
        }

        const src = normalizeText(frame.getAttribute("src"));
        if (src) {
            const exactSelector = tag + "[src=" + escapeCSSValue(src) + "]";
            if (countCSSMatches(frame.ownerDocument, exactSelector) === 1) {
                return exactSelector;
            }
            const srcWithoutQuery = src.split(/[?#]/)[0];
            const srcFile = srcWithoutQuery.split("/").filter(Boolean).pop() || srcWithoutQuery;
            if (srcFile) {
                const containsSelector = tag + "[src*=" + escapeCSSValue(srcFile) + "]";
                if (countCSSMatches(frame.ownerDocument, containsSelector) === 1) {
                    return containsSelector;
                }
            }
        }

        const xpath = getRelativeXPath(frame);
        return xpath ? "xpath:" + xpath : "";
    }

    function getFrameInfo(frame) {
        if (!isElementNode(frame)) {
            return { label: "iframe", locator: "", index: -1, id: "", name: "", title: "", src: "" };
        }
        const id = normalizeText(frame.id);
        const name = normalizeText(frame.getAttribute("name"));
        const title = normalizeText(frame.getAttribute("title"));
        const src = normalizeText(frame.getAttribute("src"));
        let index = -1;
        try {
            index = Array.from(frame.ownerDocument.querySelectorAll("iframe")).indexOf(frame);
        } catch (error) {
        }
        return {
            label: id || name || title || frame.tagName.toLowerCase() || "iframe",
            locator: getFrameLocator(frame),
            index: index,
            id: id,
            name: name,
            title: title,
            src: src,
        };
    }

    function encodeFrameCodeLocator(locator) {
        return locator ? "locator:" + String(locator) : "";
    }

    function encodeFrameCodeIndex(index) {
        return typeof index === "number" && index >= 0 ? "index:" + String(index) : "";
    }

    function getFrameCodeEntry(frame) {
        const frameInfo = getFrameInfo(frame);
        if (frameInfo.locator) {
            return encodeFrameCodeLocator(frameInfo.locator);
        }
        if (frameInfo.index >= 0) {
            return encodeFrameCodeIndex(frameInfo.index);
        }
        return "";
    }

    function getFrameContextPath() {
        const path = [];
        let currentWindow = window;
        while (currentWindow && currentWindow !== currentWindow.top) {
            try {
                const frame = currentWindow.frameElement;
                if (!frame) {
                    break;
                }
                path.unshift(getFrameInfo(frame));
                currentWindow = currentWindow.parent;
            } catch (error) {
                path.unshift({ label: "cross-origin-frame", locator: "", index: -1 });
                break;
            }
        }
        return path;
    }

    function getFrameCodePath() {
        const path = [];
        let currentWindow = window;
        while (currentWindow && currentWindow !== currentWindow.top) {
            try {
                const frame = currentWindow.frameElement;
                if (!frame) {
                    break;
                }
                path.unshift(getFrameCodeEntry(frame));
                currentWindow = currentWindow.parent;
            } catch (error) {
                path.unshift("");
                break;
            }
        }
        return path;
    }

    function getHostSelector(host) {
        return getElementLocator(host);
    }

    function getShadowPath(element) {
        const chain = [];
        let current = element;
        let depth = 0;
        while (current && typeof current.getRootNode === "function" && depth < 12) {
            const root = current.getRootNode();
            if (!isShadowRootNode(root)) {
                break;
            }
            const host = root.host;
            chain.unshift({
                mode: getShadowMode(root),
                locator: getHostSelector(host),
                label: getElementName(host) || "host",
            });
            current = host;
            depth += 1;
        }
        return chain;
    }

    function getViewportOffsetToTop() {
        let left = 0;
        let top = 0;
        let currentWindow = window;
        while (currentWindow && currentWindow !== currentWindow.top) {
            try {
                const frame = currentWindow.frameElement;
                if (!frame) {
                    break;
                }
                const rect = frame.getBoundingClientRect();
                left += rect.left;
                top += rect.top;
                currentWindow = currentWindow.parent;
            } catch (error) {
                break;
            }
        }
        return { left: left, top: top };
    }

    function getElementCenter(element) {
        const rect = element.getBoundingClientRect();
        const topOffset = getViewportOffsetToTop();
        return {
            x: Math.round(rect.left + rect.width / 2 + window.scrollX),
            y: Math.round(rect.top + rect.height / 2 + window.scrollY),
            topViewportLeft: rect.left + topOffset.left,
            topViewportTop: rect.top + topOffset.top,
            rect: rect,
        };
    }

    function getElementContext(element) {
        const framePath = getFrameContextPath();
        const parts = [];
        if (framePath.length) {
            parts.push("iframe: " + framePath.map((item) => item.label).join(" > "));
        } else {
            parts.push("main document");
        }

        const shadowPath = getShadowPath(element);
        if (shadowPath.length) {
            const last = shadowPath[shadowPath.length - 1];
            parts.push("shadow(" + last.mode + "): " + (last.label || "host"));
        } else {
            const closedRoot = getClosedShadowRoot(element);
            if (closedRoot) {
                parts.push("shadow(closed-host): " + (getElementName(element) || "host"));
            }
        }
        return parts.join(" | ");
    }

    function collectElementData(element) {
        const center = getElementCenter(element);
        const shadowPath = getShadowPath(element);
        const closedShadowHost = shadowPath.length === 0 && getClosedShadowRoot(element)
            ? { locator: getElementLocator(element), label: getElementName(element) || "host" }
            : null;
        return {
            tag: element.tagName ? element.tagName.toLowerCase() : "",
            name: getElementName(element),
            text: getVisibleText(element),
            absoluteXPath: getAbsoluteXPath(element),
            relativeXPath: getRelativeXPath(element),
            centerX: center.x,
            centerY: center.y,
            context: getElementContext(element),
            framePath: getFrameContextPath(),
            frameCodePath: getFrameCodePath(),
            shadowPath: shadowPath,
            closedShadowHost: closedShadowHost,
            topViewportLeft: center.topViewportLeft,
            topViewportTop: center.topViewportTop,
            width: center.rect.width,
            height: center.rect.height,
        };
    }

    function updateHighlight(element) {
        const highlight = ensureHighlight();
        if (!element || !document.documentElement.contains(element)) {
            highlight.style.display = "none";
            return;
        }
        const rect = element.getBoundingClientRect();
        highlight.style.display = "block";
        highlight.style.borderColor = state.mode === "locked"
            ? "rgba(59, 130, 246, 0.98)"
            : "rgba(34, 197, 94, 0.95)";
        highlight.style.background = state.mode === "locked"
            ? "rgba(96, 165, 250, 0.12)"
            : "rgba(34, 197, 94, 0.10)";
        highlight.style.left = rect.left + window.scrollX + "px";
        highlight.style.top = rect.top + window.scrollY + "px";
        highlight.style.width = Math.max(rect.width, 0) + "px";
        highlight.style.height = Math.max(rect.height, 0) + "px";
    }

    function getEventElement(event) {
        if (!event) {
            return null;
        }
        const path = typeof event.composedPath === "function" ? event.composedPath() : null;
        let target = null;
        if (Array.isArray(path)) {
            for (const item of path) {
                if (isElementNode(item)) {
                    target = item;
                    break;
                }
            }
        }
        if (!target && isElementNode(event.target)) {
            target = event.target;
        }
        if (!isElementNode(target)) {
            return null;
        }
        const closedRoot = getClosedShadowRoot(target);
        const deepTarget = getDeepestShadowElement(closedRoot, event.clientX, event.clientY);
        return isElementNode(deepTarget) ? deepTarget : target;
    }

    function updateTopHighlightFromData(data) {
        if (!isTopWindow) {
            return;
        }
        const highlight = ensureHighlight();
        if (!data) {
            highlight.style.display = "none";
            return;
        }
        highlight.style.display = "block";
        highlight.style.borderColor = state.mode === "locked"
            ? "rgba(59, 130, 246, 0.98)"
            : "rgba(34, 197, 94, 0.95)";
        highlight.style.background = state.mode === "locked"
            ? "rgba(96, 165, 250, 0.12)"
            : "rgba(34, 197, 94, 0.10)";
        highlight.style.left = Math.max(data.topViewportLeft + topWindowRef.scrollX, 0) + "px";
        highlight.style.top = Math.max(data.topViewportTop + topWindowRef.scrollY, 0) + "px";
        highlight.style.width = Math.max(data.width || 0, 0) + "px";
        highlight.style.height = Math.max(data.height || 0, 0) + "px";
    }

    function getDisplayData() {
        if (state.mode === "locked" || state.mode === "paused") {
            return state.selectedData;
        }
        return state.hoverData;
    }

    function getStatusText() {
        if (state.mode === "locked") {
            return "已锁定";
        }
        if (state.mode === "paused") {
            return "已暂停";
        }
        return "待选择";
    }

    function copyText(text, button) {
        const value = String(text || "");
        const markCopied = () => {
            if (!button) {
                return;
            }
            button.setAttribute("data-copied", "true");
            const original = button.getAttribute("data-copy-label") || "复制";
            button.textContent = "已复制";
            topWindowRef.setTimeout(() => {
                button.removeAttribute("data-copied");
                button.textContent = original;
            }, 1200);
        };

        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(value).then(markCopied).catch(() => {
                const textarea = document.createElement("textarea");
                textarea.value = value;
                document.body.appendChild(textarea);
                textarea.select();
                try {
                    document.execCommand("copy");
                    markCopied();
                } catch (error) {
                }
                textarea.remove();
            });
            return;
        }

        const textarea = document.createElement("textarea");
        textarea.value = value;
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand("copy");
            markCopied();
        } catch (error) {
        }
        textarea.remove();
    }

    function containsClosedShadow(shadowPath) {
        return Array.isArray(shadowPath) && shadowPath.some((item) => item && item.mode === "closed");
    }

    function getFrameCodeLocator(frameInfo) {
        if (!frameInfo) {
            return "";
        }
        if (typeof frameInfo === "string") {
            const label = normalizeText(frameInfo);
            if (isGenericFrameLabel(label)) {
                return "";
            }
            if (isSafeCSSID(label)) {
                return "#" + label;
            }
            return "iframe[title=" + escapeCSSValue(label) + "]";
        }
        if (typeof frameInfo !== "object") {
            return "";
        }
        if (frameInfo.locator) {
            return String(frameInfo.locator);
        }
        if (frameInfo.id) {
            return "#" + String(frameInfo.id);
        }
        if (frameInfo.name) {
            return "iframe[name=" + escapeCSSValue(frameInfo.name) + "]";
        }
        if (frameInfo.title) {
            return "iframe[title=" + escapeCSSValue(frameInfo.title) + "]";
        }
        if (frameInfo.src) {
            return "iframe[src=" + escapeCSSValue(frameInfo.src) + "]";
        }
        if (!isGenericFrameLabel(frameInfo.label)) {
            if (isSafeCSSID(frameInfo.label)) {
                return "#" + String(frameInfo.label);
            }
            return "iframe[title=" + escapeCSSValue(frameInfo.label) + "]";
        }
        return "";
    }

    function getFrameCodeIndex(frameInfo) {
        if (!frameInfo || typeof frameInfo !== "object") {
            return -1;
        }
        return typeof frameInfo.index === "number" ? frameInfo.index : -1;
    }

    function decodeFrameCodeLocator(frameCode) {
        if (typeof frameCode !== "string" || !frameCode.startsWith("locator:")) {
            return "";
        }
        return frameCode.slice("locator:".length);
    }

    function decodeFrameCodeIndex(frameCode) {
        if (typeof frameCode !== "string" || !frameCode.startsWith("index:")) {
            return -1;
        }
        const value = Number.parseInt(frameCode.slice("index:".length), 10);
        return Number.isInteger(value) && value >= 0 ? value : -1;
    }

    function isUsableFrameCodeEntry(frameCode) {
        return !!decodeFrameCodeLocator(frameCode) || decodeFrameCodeIndex(frameCode) >= 0;
    }

    function getLegacyFrameCodeEntries(framePath) {
        if (!Array.isArray(framePath)) {
            return [];
        }
        return framePath.map((frameInfo) => {
            const locator = getFrameCodeLocator(frameInfo);
            if (locator) {
                return encodeFrameCodeLocator(locator);
            }
            const index = getFrameCodeIndex(frameInfo);
            if (index >= 0) {
                return encodeFrameCodeIndex(index);
            }
            return "";
        });
    }

    function mergeFrameCodeEntries(primaryEntries, fallbackEntries) {
        const size = Math.max(primaryEntries.length, fallbackEntries.length);
        const merged = [];
        for (let index = 0; index < size; index += 1) {
            const primary = typeof primaryEntries[index] === "string" ? primaryEntries[index] : "";
            if (isUsableFrameCodeEntry(primary)) {
                merged.push(primary);
                continue;
            }
            const fallback = typeof fallbackEntries[index] === "string" ? fallbackEntries[index] : "";
            merged.push(fallback);
        }
        return merged;
    }

    function getContextFrameCodeEntries(context) {
        const text = String(context || "");
        const prefix = "iframe: ";
        const start = text.indexOf(prefix);
        if (start < 0) {
            return [];
        }
        const end = text.indexOf(" | ", start);
        const frameSection = (end >= 0 ? text.slice(start, end) : text.slice(start)).trim();
        if (!frameSection.startsWith(prefix)) {
            return [];
        }
        return frameSection.slice(prefix.length).split(" > ").map((rawLabel) => {
            const label = normalizeText(rawLabel);
            if (isGenericFrameLabel(label)) {
                return "";
            }
            if (isSafeCSSID(label)) {
                return encodeFrameCodeLocator("#" + label);
            }
            return encodeFrameCodeLocator("iframe[title=" + escapeCSSValue(label) + "]");
        });
    }

    function getFrameCodeEntries(data) {
        const frameCodePathEntries = data && Array.isArray(data.frameCodePath)
            ? data.frameCodePath.map((entry) => typeof entry === "string" ? entry : "")
            : [];
        const legacyEntries = getLegacyFrameCodeEntries(data ? data.framePath : null);
        const contextEntries = getContextFrameCodeEntries(data ? data.context : "");
        if (frameCodePathEntries.length > 0) {
            return mergeFrameCodeEntries(
                mergeFrameCodeEntries(frameCodePathEntries, legacyEntries),
                contextEntries
            );
        }
        if (legacyEntries.length > 0) {
            return mergeFrameCodeEntries(legacyEntries, contextEntries);
        }
        return contextEntries;
    }

    function getElementKeyFromData(data) {
        if (!data) {
            return "";
        }
        return String(data.absoluteXPath || data.relativeXPath || ((data.tag || "element") + "@" + String(data.centerX || 0) + "," + String(data.centerY || 0)));
    }

    function getElementKey(element) {
        if (!isElementNode(element)) {
            return "";
        }
        return getAbsoluteXPath(element) || getRelativeXPath(element) || (getElementName(element) + "@" + String(element.tagName || ""));
    }

    function evaluateXPathAll(doc, expr) {
        if (!doc || !expr) {
            return [];
        }
        try {
            const snapshot = doc.evaluate(expr, doc, null, XPathResult.ORDERED_NODE_SNAPSHOT_TYPE, null);
            const results = [];
            for (let index = 0; index < snapshot.snapshotLength; index += 1) {
                const item = snapshot.snapshotItem(index);
                if (isElementNode(item)) {
                    results.push(item);
                }
            }
            return results;
        } catch (error) {
            return [];
        }
    }

    function getCSSLocatorOnly(element) {
        if (!isElementNode(element) || !element.tagName) {
            return "";
        }
        if (element.id) {
            return "#" + element.id;
        }
        const tag = element.tagName.toLowerCase();
        const attrs = ["data-testid", "data-test", "data-qa", "name", "aria-label", "title", "role"];
        for (let index = 0; index < attrs.length; index += 1) {
            const attr = attrs[index];
            const value = normalizeText(element.getAttribute(attr));
            if (!value) {
                continue;
            }
            const selector = tag + "[" + attr + "=" + escapeCSSValue(value) + "]";
            if (countCSSMatches(element.ownerDocument, selector) === 1) {
                return selector;
            }
        }
        return "";
    }

    function dedupeElements(elements) {
        const seen = {};
        const unique = [];
        (Array.isArray(elements) ? elements : []).forEach((element) => {
            if (!isElementNode(element)) {
                return;
            }
            const key = getElementKey(element);
            if (!key || seen[key]) {
                return;
            }
            seen[key] = true;
            unique.push(element);
        });
        return unique;
    }

    function sameItemKeys(left, right) {
        if (!Array.isArray(left) || !Array.isArray(right) || left.length !== right.length) {
            return false;
        }
        for (let index = 0; index < left.length; index += 1) {
            if (left[index] !== right[index]) {
                return false;
            }
        }
        return true;
    }

    function makeCandidateKey(kind, value) {
        return String(kind || "") + ":" + String(value || "");
    }

    function addGroupCandidate(store, candidate) {
        if (!candidate || !candidate.value || !Array.isArray(candidate.elements)) {
            return;
        }
        const elements = dedupeElements(candidate.elements);
        if (elements.length < 2) {
            return;
        }
        const key = makeCandidateKey(candidate.kind, candidate.value);
        if (store[key]) {
            return;
        }
        const itemKeys = elements.map((element) => getElementKey(element)).filter(Boolean).sort();
        store[key] = {
            key: key,
            kind: candidate.kind,
            strategy: candidate.strategy,
            value: candidate.value,
            elements: elements,
            count: elements.length,
            depth: candidate.depth || 0,
            score: candidate.score || 0,
            itemKeys: itemKeys,
        };
    }

    function findBestCandidateByKind(candidates, kind, itemKeys) {
        let matched = null;
        let fallback = null;
        (Array.isArray(candidates) ? candidates : []).forEach((candidate) => {
            if (!candidate || candidate.kind !== kind) {
                return;
            }
            if (!fallback || candidate.score > fallback.score) {
                fallback = candidate;
            }
            if (Array.isArray(itemKeys) && sameItemKeys(candidate.itemKeys || [], itemKeys)) {
                if (!matched || candidate.score > matched.score) {
                    matched = candidate;
                }
            }
        });
        return matched || fallback;
    }

    function buildGroupCode(referenceData, group) {
        if (!referenceData) {
            return "// 锁定一个元素后，这里会生成元素组的 ruyiPage Go 示例代码";
        }

        const lines = [];
        let currentVar = "page";
        lines.push("// ruyiPage generated group snippet");
        lines.push("// add: import time");

        getFrameCodeEntries(referenceData).forEach((frameCode, index) => {
            const frameVar = "frame" + String(index + 1);
            const frameLocator = decodeFrameCodeLocator(frameCode);
            if (frameLocator) {
                lines.push(frameVar + ", _ := " + currentVar + ".GetFrame(" + quoteCode(frameLocator) + ")");
                currentVar = frameVar;
                return;
            }
            const frameIndex = decodeFrameCodeIndex(frameCode);
            if (frameIndex >= 0) {
                lines.push(frameVar + ", _ := " + currentVar + ".GetFrame(" + String(frameIndex) + ")");
                currentVar = frameVar;
                return;
            }
        });

        const shadowPath = referenceData.shadowPath || [];
        shadowPath.forEach((shadow, index) => {
            const hostVar = "shadowHost" + String(index + 1);
            const rootVar = "shadowRoot" + String(index + 1);
            const hostLocator = shadow && shadow.locator ? shadow.locator : "";
            if (hostLocator) {
                lines.push(hostVar + ", _ := " + currentVar + ".Ele(" + quoteCode(hostLocator) + ", 1, 5*time.Second)");
            } else {
                lines.push("// 无法稳定还原 shadow host 定位，请手动确认");
                lines.push(hostVar + " := /* locate host manually */");
            }
            if (shadow && shadow.mode === "closed") {
                lines.push(rootVar + ", _ := " + hostVar + ".ClosedShadowRoot()");
            } else {
                lines.push(rootVar + ", _ := " + hostVar + ".ShadowRoot()");
            }
            currentVar = rootVar;
        });

        if (group && group.commonXPath) {
            lines.push("items, _ := " + currentVar + ".Eles(" + quoteCode("xpath:" + group.commonXPath) + ", 5*time.Second)");
        } else if (group && group.commonCss) {
            lines.push("items, _ := " + currentVar + ".Eles(" + quoteCode(group.commonCss) + ", 5*time.Second)");
        } else {
            lines.push("// 暂未生成稳定的组选择器，请先在“元素组”页签里选择策略");
            lines.push("items := []*ruyipage.FirefoxElement{}");
        }
        lines.push("for _, item := range items {");
        lines.push('    text, _ := item.Text()');
        lines.push('    fmt.Println(text)');
        lines.push("}");
        return lines.join("\n");
    }

    function buildManualUnionXPath(items) {
        const parts = [];
        (Array.isArray(items) ? items : []).forEach((item) => {
            const expr = String(item && (item.absoluteXPath || item.relativeXPath) || "");
            if (expr) {
                parts.push(expr);
            }
        });
        return parts.join(" | ");
    }

    function collectGroupItems(elements) {
        return dedupeElements(elements).map((element) => collectElementData(element));
    }

    function applyGroupSelection(candidate, candidates, anchorElement) {
        const chosen = candidate || null;
        const groupItems = chosen ? collectGroupItems(chosen.elements) : [];
        const itemKeys = groupItems.map((item) => getElementKeyFromData(item)).filter(Boolean).sort();
        const bestXPath = findBestCandidateByKind(candidates, "xpath", itemKeys);
        const bestCss = findBestCandidateByKind(candidates, "css", itemKeys);
        const group = {
            items: groupItems,
            candidates: Array.isArray(candidates) ? candidates : [],
            selectedKey: chosen ? chosen.key : "",
            anchorName: anchorElement ? (getElementName(anchorElement) || "") : "",
            statusText: groupItems.length >= 2 ? "已捕获 " + String(groupItems.length) + " 个相似元素" : "未推断到稳定的相似元素组",
            commonXPath: bestXPath ? bestXPath.value : "",
            commonCss: bestCss ? bestCss.value : "",
            diagnostics: Array.isArray(candidates) ? candidates.map((item) => ({
                kind: item.kind,
                strategy: item.strategy,
                value: item.value,
                count: item.count,
            })) : [],
            code: "",
        };
        group.code = buildGroupCode(state.selectedData || state.hoverData, group);
        state.group = group;
        state.activeTab = "group";
    }

    function buildClassXPath(tagName, className) {
        return getXPathNodeName({ tagName: tagName, namespaceURI: "http://www.w3.org/1999/xhtml", localName: tagName }) +
            "[contains(concat(' ', normalize-space(@class), ' '), " + escapeXPathLiteral(" " + className + " ") + ")]";
    }

    function buildGroupCandidates(selectedElement) {
        const store = {};
        const diagnostics = [];
        if (!isElementNode(selectedElement) || !selectedElement.ownerDocument) {
            return { anchorElement: null, candidates: [], diagnostics: diagnostics };
        }

        const tagName = selectedElement.tagName.toLowerCase();
        let anchorElement = null;
        let current = selectedElement.parentElement;
        let depth = 0;
        while (current && depth < 5) {
            const anchorXPath = getRelativeXPath(current) || getAbsoluteXPath(current);
            const anchorCss = getCSSLocatorOnly(current);
            if (!anchorElement && anchorXPath) {
                anchorElement = current;
            }
            const children = Array.from(current.children).filter((item) => isElementNode(item));
            const sameTagChildren = children.filter((item) => item.tagName && item.tagName.toLowerCase() === tagName);
            if (anchorXPath && sameTagChildren.length >= 2) {
                addGroupCandidate(store, {
                    kind: "xpath",
                    strategy: "同父同tag",
                    value: anchorXPath + "/" + tagName,
                    elements: sameTagChildren,
                    depth: depth,
                    score: sameTagChildren.length * 100 + 50 - depth * 20,
                });
                if (anchorCss) {
                    addGroupCandidate(store, {
                        kind: "css",
                        strategy: "同父同tag",
                        value: anchorCss + " > " + tagName,
                        elements: sameTagChildren,
                        depth: depth,
                        score: sameTagChildren.length * 100 + 48 - depth * 20,
                    });
                }
            }

            const attrs = ["data-testid", "data-test", "data-qa", "data-group", "data-list", "role"];
            attrs.forEach((attr) => {
                const attrValue = normalizeText(selectedElement.getAttribute(attr));
                if (!anchorXPath || !attrValue) {
                    return;
                }
                const matches = sameTagChildren.filter((item) => normalizeText(item.getAttribute(attr)) === attrValue);
                if (matches.length < 2) {
                    return;
                }
                addGroupCandidate(store, {
                    kind: "xpath",
                    strategy: "属性 " + attr,
                    value: anchorXPath + "/" + tagName + "[@" + attr + "=" + escapeXPathLiteral(attrValue) + "]",
                    elements: matches,
                    depth: depth,
                    score: matches.length * 100 + 90 - depth * 20,
                });
                if (anchorCss) {
                    addGroupCandidate(store, {
                        kind: "css",
                        strategy: "属性 " + attr,
                        value: anchorCss + " > " + tagName + "[" + attr + "=" + escapeCSSValue(attrValue) + "]",
                        elements: matches,
                        depth: depth,
                        score: matches.length * 100 + 88 - depth * 20,
                    });
                }
            });

            const classList = typeof selectedElement.className === "string"
                ? selectedElement.className.trim().split(/\s+/).filter(Boolean).slice(0, 2)
                : [];
            classList.forEach((className) => {
                if (!anchorXPath || !className) {
                    return;
                }
                const matches = sameTagChildren.filter((item) => item.classList && item.classList.contains(className));
                if (matches.length < 2) {
                    return;
                }
                addGroupCandidate(store, {
                    kind: "xpath",
                    strategy: "class " + className,
                    value: anchorXPath + "/" + tagName + "[contains(concat(' ', normalize-space(@class), ' '), " + escapeXPathLiteral(" " + className + " ") + ")]",
                    elements: matches,
                    depth: depth,
                    score: matches.length * 100 + 72 - depth * 20,
                });
                if (anchorCss) {
                    addGroupCandidate(store, {
                        kind: "css",
                        strategy: "class " + className,
                        value: anchorCss + " > " + tagName + "." + className,
                        elements: matches,
                        depth: depth,
                        score: matches.length * 100 + 70 - depth * 20,
                    });
                }
            });

            const descendants = dedupeElements(Array.from(current.querySelectorAll(tagName)));
            if (anchorXPath && descendants.length >= 2 && descendants.length <= 24) {
                addGroupCandidate(store, {
                    kind: "xpath",
                    strategy: "祖先后代同tag",
                    value: anchorXPath + "//" + tagName,
                    elements: descendants,
                    depth: depth,
                    score: descendants.length * 100 + 20 - depth * 20,
                });
                if (anchorCss) {
                    addGroupCandidate(store, {
                        kind: "css",
                        strategy: "祖先后代同tag",
                        value: anchorCss + " " + tagName,
                        elements: descendants,
                        depth: depth,
                        score: descendants.length * 100 + 18 - depth * 20,
                    });
                }
            }

            current = current.parentElement;
            depth += 1;
        }

        const candidates = Object.keys(store).map((key) => store[key]).sort((left, right) => right.score - left.score);
        diagnostics.push({
            kind: "meta",
            strategy: "候选策略数",
            value: String(candidates.length),
            count: candidates.length,
        });
        return {
            anchorElement: anchorElement || selectedElement.parentElement,
            candidates: candidates,
            diagnostics: diagnostics,
        };
    }

    function captureGroup() {
        const seed = localState.selectedElement;
        if (!isElementNode(seed)) {
            state.group.statusText = "先锁定一个元素，再点击“捕获相似元素”";
            state.activeTab = "group";
            syncTopUI();
            return;
        }
        const result = buildGroupCandidates(seed);
        if (!result.candidates.length) {
            state.group = {
                items: [],
                candidates: [],
                selectedKey: "",
                anchorName: result.anchorElement ? getElementName(result.anchorElement) : "",
                statusText: "未找到稳定的相似元素组，可尝试换一个列表项重新捕获",
                commonXPath: "",
                commonCss: "",
                code: buildGroupCode(state.selectedData || state.hoverData, null),
                diagnostics: result.diagnostics || [],
            };
            state.activeTab = "group";
            syncTopUI();
            return;
        }
        applyGroupSelection(result.candidates[0], result.candidates, result.anchorElement);
        syncTopUI();
    }

    function selectGroupStrategy(key) {
        if (!key || !state.group || !Array.isArray(state.group.candidates)) {
            return;
        }
        const candidate = state.group.candidates.find((item) => item && item.key === key);
        if (!candidate) {
            return;
        }
        applyGroupSelection(candidate, state.group.candidates, localState.selectedElement ? localState.selectedElement.parentElement : null);
        syncTopUI();
    }

    function removeGroupItem(index) {
        if (!state.group || !Array.isArray(state.group.items) || index < 0 || index >= state.group.items.length) {
            return;
        }
        const nextItems = state.group.items.filter((_, itemIndex) => itemIndex !== index);
        if (!nextItems.length) {
            clearGroup();
            return;
        }
        state.group.items = nextItems;
        state.group.selectedKey = "";
        state.group.commonXPath = buildManualUnionXPath(nextItems);
        state.group.commonCss = "";
        state.group.statusText = nextItems.length >= 2
            ? "已移除 1 项，当前保留 " + String(nextItems.length) + " 个元素"
            : "当前仅剩 1 个元素，组选择器已退化为手工并集";
        state.group.diagnostics = (state.group.diagnostics || []).concat([{
            kind: "meta",
            strategy: "manual removal",
            value: "已改为并集 XPath",
            count: nextItems.length,
        }]);
        state.group.code = buildGroupCode(state.selectedData || state.hoverData, state.group);
        syncTopUI();
    }

    function clearGroup() {
        state.group = {
            items: [],
            candidates: [],
            selectedKey: "",
            anchorName: "",
            statusText: "",
            commonXPath: "",
            commonCss: "",
            code: buildGroupCode(state.selectedData || state.hoverData, null),
            diagnostics: [],
        };
        syncTopUI();
    }

    function renderStrategyList(group) {
        const candidates = group && Array.isArray(group.candidates) ? group.candidates : [];
        if (!candidates.length) {
            return renderField("策略候选", "锁定一个列表项后，点击“捕获相似元素”生成组策略。", {});
        }
        return [
            '<section class="ruyi-xpath-picker__field">',
            '  <div class="ruyi-xpath-picker__field-header">',
            '    <span class="ruyi-xpath-picker__label">策略候选</span>',
            "  </div>",
            '  <div class="ruyi-xpath-picker__strategy-list">',
            candidates.map((candidate) => [
                '<button type="button" class="ruyi-xpath-picker__strategy-item" data-group-strategy-key="' + escapeAttribute(candidate.key || "") + '" data-active="' + String((group.selectedKey || "") === (candidate.key || "")) + '">',
                '  <div class="ruyi-xpath-picker__strategy-title">' + escapeHTML(String(candidate.kind || "").toUpperCase() + " / " + String(candidate.strategy || "")) + "</div>",
                '  <div class="ruyi-xpath-picker__strategy-sub">' + escapeHTML(candidate.value || "") + "</div>",
                '  <div class="ruyi-xpath-picker__strategy-meta"><span>数量 ' + escapeHTML(String(candidate.count || 0)) + "</span><span>深度 " + escapeHTML(String(candidate.depth || 0)) + "</span></div>",
                "</button>",
            ].join("")).join(""),
            "  </div>",
            "</section>",
        ].join("");
    }

    function renderGroupItems(group) {
        const items = group && Array.isArray(group.items) ? group.items : [];
        const headerAction = items.length
            ? '<button type="button" class="ruyi-xpath-picker__copy" data-clear-group="true">清空组</button>'
            : "";
        return [
            '<section class="ruyi-xpath-picker__field">',
            '  <div class="ruyi-xpath-picker__field-header">',
            '    <span class="ruyi-xpath-picker__label">元素组结果</span>',
            headerAction,
            "  </div>",
            items.length ? (
                '<div class="ruyi-xpath-picker__group-list">' +
                items.map((item, index) => [
                    '<div class="ruyi-xpath-picker__group-item">',
                    '  <div class="ruyi-xpath-picker__group-item-main">',
                    '    <div class="ruyi-xpath-picker__group-item-title">' + escapeHTML(String(index + 1) + ". " + (item.name || item.tag || "element")) + "</div>",
                    '    <div class="ruyi-xpath-picker__group-item-sub">' + escapeHTML(item.text || item.relativeXPath || item.absoluteXPath || "-") + "</div>",
                    "  </div>",
                    '  <button type="button" class="ruyi-xpath-picker__button ruyi-xpath-picker__button--secondary ruyi-xpath-picker__button--tiny" data-remove-group-index="' + String(index) + '">移除</button>',
                    "</div>",
                ].join("")).join("") +
                "</div>"
            ) : '<div class="ruyi-xpath-picker__value">暂无元素组结果</div>',
            "</section>",
        ].join("");
    }

    function renderGroupPanel(group) {
        return [
            renderField("状态", group && group.statusText ? group.statusText : "先锁定一个元素，再点击“捕获相似元素”。", {}),
            renderField("组锚点", group && group.anchorName ? group.anchorName : "未解析到组锚点", {}),
            renderField("组 XPath", group && group.commonXPath ? group.commonXPath : "未生成稳定组 XPath", { isCode: true, canCopy: !!(group && group.commonXPath), copyLabel: "复制组XPath" }),
            renderField("组 CSS", group && group.commonCss ? group.commonCss : "未生成稳定组 CSS", { isCode: true, canCopy: !!(group && group.commonCss), copyLabel: "复制组CSS" }),
            renderStrategyList(group),
            renderGroupItems(group),
            renderCodeField(group && group.code ? group.code : buildGroupCode(state.selectedData || state.hoverData, group)),
        ].join("");
    }

    function buildRuyiPageCode(data) {
        if (!data) {
            return "// 点击一个元素后，这里会生成 ruyiPage Go 示例代码";
        }

        const lines = [];
        let currentVar = "page";

        lines.push("// ruyiPage generated snippet");
        lines.push("// add: import time");

        getFrameCodeEntries(data).forEach((frameCode, index) => {
            const frameVar = "frame" + String(index + 1);
            const frameLocator = decodeFrameCodeLocator(frameCode);
            if (frameLocator) {
                lines.push(frameVar + ", _ := " + currentVar + ".GetFrame(" + quoteCode(frameLocator) + ")");
                currentVar = frameVar;
                return;
            }
            const frameIndex = decodeFrameCodeIndex(frameCode);
            if (frameIndex >= 0) {
                lines.push(frameVar + ", _ := " + currentVar + ".GetFrame(" + String(frameIndex) + ")");
                currentVar = frameVar;
                return;
            }
            lines.push("// 无法稳定还原第 " + String(index + 1) + " 层 iframe 定位，请手动确认");
        });

        const shadowPath = data.shadowPath || [];
        shadowPath.forEach((shadow, index) => {
            const hostVar = "shadowHost" + String(index + 1);
            const rootVar = "shadowRoot" + String(index + 1);
            const hostLocator = shadow && shadow.locator ? shadow.locator : "";
            if (hostLocator) {
                lines.push(hostVar + ", _ := " + currentVar + ".Ele(" + quoteCode(hostLocator) + ", 1, 5*time.Second)");
            } else {
                lines.push("// 无法稳定还原 shadow host 定位，请手动确认");
                lines.push(hostVar + " := /* locate host manually */");
            }
            if (shadow && shadow.mode === "closed") {
                lines.push(rootVar + ", _ := " + hostVar + ".ClosedShadowRoot()");
            } else {
                lines.push(rootVar + ", _ := " + hostVar + ".ShadowRoot()");
            }
            currentVar = rootVar;
        });

        const primarySelector = data.relativeXPath || data.absoluteXPath || "";
        if (primarySelector) {
            lines.push("target, _ := " + currentVar + ".Ele(" + quoteCode("xpath:" + primarySelector) + ", 1, 5*time.Second)");
        } else {
            lines.push("// 无法生成 XPath，建议手动补充 selector");
            lines.push("target, _ := " + currentVar + ".Ele(" + quoteCode(data.name || data.tag || "") + ", 1, 5*time.Second)");
        }

        if (shadowPath.length) {
            lines.push("");
            lines.push("// 也可改成 host.WithShadow(" + quoteCode(containsClosedShadow(shadowPath) ? "closed" : "open") + ") 形式");
        }

        if (data.closedShadowHost && (!data.shadowPath || !data.shadowPath.length)) {
            lines.push("");
            lines.push("// 当前命中的是 closed shadow host，页面提供了 __ruyiGetClosedShadowRoot 调试桥");
            if (data.closedShadowHost.locator) {
                lines.push("host, _ := " + currentVar + ".Ele(" + quoteCode(data.closedShadowHost.locator) + ", 1, 5*time.Second)");
                lines.push("shadowRoot, _ := host.ClosedShadowRoot()");
                lines.push("// 再继续在 shadowRoot 下查找目标元素");
            } else {
                lines.push("// host 需要手动定位后再调用 ClosedShadowRoot()");
            }
        }

        if (String(data.context || "").includes("shadow(") && (!data.shadowPath || !data.shadowPath.length) && !data.closedShadowHost) {
            lines.push("");
            lines.push("// 注意：当前命中元素位于 shadow 场景，但未能还原 host 链。");
            lines.push("// closed shadow 需要页面提供 __ruyiGetClosedShadowRoot 调试桥后，才能稳定生成访问代码。");
        }

        return lines.join("\n");
    }

    function renderField(label, value, options) {
        const settings = options || {};
        const copyLabel = settings.copyLabel || "复制";
        const canCopy = !!settings.canCopy && value !== "";
        const codeAttr = settings.isCode ? ' data-code="true"' : "";
        return [
            '<section class="ruyi-xpath-picker__field">',
            '  <div class="ruyi-xpath-picker__field-header">',
            '    <span class="ruyi-xpath-picker__label">' + escapeHTML(label) + "</span>",
            canCopy ? '    <button type="button" class="ruyi-xpath-picker__copy" data-copy-label="' + escapeAttribute(copyLabel) + '" data-copy-value="' + escapeAttribute(value) + '">' + escapeHTML(copyLabel) + "</button>" : "",
            "  </div>",
            '  <div class="ruyi-xpath-picker__value"' + codeAttr + ">" + escapeHTML(value || "-") + "</div>",
            "</section>",
        ].join("");
    }

    function renderCodeField(code) {
        return [
            '<section class="ruyi-xpath-picker__field">',
            '  <div class="ruyi-xpath-picker__field-header">',
            '    <span class="ruyi-xpath-picker__label">ruyiPage代码生成</span>',
            '    <button type="button" class="ruyi-xpath-picker__copy" data-copy-label="复制代码" data-copy-value="' + escapeAttribute(code) + '">复制代码</button>',
            "  </div>",
            '  <div class="ruyi-xpath-picker__code-block">' + escapeHTML(code) + "</div>",
            "</section>",
        ].join("");
    }

    function syncTopUI() {
        if (isTopWindow) {
            renderFields();
            updateTopHighlightFromData(getDisplayData());
            return;
        }
        try {
            if (topWindowRef && typeof topWindowRef.__ruyiInitXPathPicker === "function") {
                topWindowRef.__ruyiInitXPathPicker();
            }
        } catch (error) {
        }
    }

    function renderFields() {
        if (!isTopWindow) {
            return;
        }
        const panel = ensurePanel();
        const meta = panel.querySelector('[data-role="meta"]');
        const intro = panel.querySelector('[data-role="intro"]');
        const status = panel.querySelector('[data-role="status"]');
        const unlockButton = panel.querySelector('[data-action="unlock"]');
        const captureGroupButton = panel.querySelector('[data-action="capture-group"]');
        const pauseButton = panel.querySelector('[data-action="pause"]');
        const toggleButtons = panel.querySelectorAll('[data-action="toggle"]');
        const tabs = panel.querySelectorAll("[data-tab]");

        panel.setAttribute("data-collapsed", state.collapsed ? "true" : "false");
        toggleButtons.forEach((button) => {
            const isIcon = button.classList.contains("ruyi-xpath-picker__button--icon");
            button.textContent = state.collapsed ? (isIcon ? "+" : "展开") : (isIcon ? "-" : "收起");
            button.setAttribute("aria-label", state.collapsed ? "展开 XPath Picker" : "收起 XPath Picker");
        });
        tabs.forEach((button) => {
            button.setAttribute("data-active", button.getAttribute("data-tab") === state.activeTab ? "true" : "false");
        });

        const data = getDisplayData();
        if (status) {
            status.textContent = getStatusText();
        }
        if (unlockButton) {
            unlockButton.disabled = state.mode !== "locked";
        }
        if (captureGroupButton) {
            captureGroupButton.disabled = !state.selectedData;
        }
        if (pauseButton) {
            pauseButton.textContent = state.mode === "paused" ? "恢复选择" : "暂停选择";
        }

        if (!data) {
            if (intro) {
                intro.textContent = state.mode === "paused"
                    ? "当前已暂停选择，点击“恢复选择”后可继续检查页面元素。"
                    : "移动鼠标可预览目标，点击页面元素后会锁定当前结果。";
            }
            meta.innerHTML = state.activeTab === "group" ? renderGroupPanel(state.group) : "";
            return;
        }

        if (intro) {
            if (state.activeTab === "group") {
                intro.textContent = "“元素组”页签会根据当前锁定元素推断相似元素列表，并允许切换组策略。";
            } else {
                intro.textContent = state.mode === "locked"
                    ? "当前结果已锁定，点击“继续选择”后可重新选择其他元素。"
                    : state.mode === "paused"
                        ? "当前已暂停选择，保留最近一次锁定结果。"
                        : "当前为预览态，点击元素后会锁定此结果。";
            }
        }

        if (state.activeTab === "ruyipage") {
            meta.innerHTML = renderCodeField(buildRuyiPageCode(data));
            return;
        }
        if (state.activeTab === "group") {
            meta.innerHTML = renderGroupPanel(state.group);
            return;
        }

        meta.innerHTML = [
            renderField("Tag", data.tag || "-", {}),
            renderField("Name", data.name || "-", {}),
            renderField("Text", data.text || "-", {}),
            renderField("XPath (absolute)", data.absoluteXPath || "-", { isCode: true, canCopy: !!data.absoluteXPath, copyLabel: "复制绝对XPath" }),
            renderField("XPath (relative)", data.relativeXPath || "-", { isCode: true, canCopy: !!data.relativeXPath, copyLabel: "复制相对XPath" }),
            renderField("Center", "(" + data.centerX + ", " + data.centerY + ")", {}),
            renderField("Context", data.context || "-", {}),
        ].join("");
    }

    function unlockSelection() {
        state.mode = "idle";
        state.selectedData = null;
        state.hoverData = null;
        state.activeTab = "info";
        localState.selectedElement = null;
        localState.hoverElement = null;
        clearGroup();
        if (localState.highlight) {
            localState.highlight.style.display = "none";
        }
        syncTopUI();
    }

    function toggleCollapsed(forceValue) {
        state.collapsed = typeof forceValue === "boolean" ? forceValue : !state.collapsed;
        syncTopUI();
    }

    function togglePaused() {
        state.mode = state.mode === "paused" ? "idle" : "paused";
        if (state.mode === "idle") {
            state.hoverData = null;
            localState.hoverElement = null;
            updateHighlight(null);
        } else if (localState.selectedElement && document.documentElement.contains(localState.selectedElement)) {
            updateHighlight(localState.selectedElement);
        } else {
            updateHighlight(null);
        }
        syncTopUI();
    }

    function isPickerNode(node) {
        return !!(node && node.closest && node.closest("#" + PANEL_ID + ", #" + HIGHLIGHT_ID));
    }

    function handleMove(event) {
        const target = getEventElement(event);
        if (state.mode !== "idle" || isPickerNode(target)) {
            return;
        }
        if (!isElementNode(target)) {
            return;
        }
        localState.hoverElement = target;
        state.hoverData = collectElementData(target);
        updateHighlight(target);
        syncTopUI();
    }

    function handleClick(event) {
        const target = getEventElement(event);
        if (state.mode !== "idle" || isPickerNode(target)) {
            return;
        }
        if (!isElementNode(target)) {
            return;
        }
        event.preventDefault();
        event.stopPropagation();
        if (typeof event.stopImmediatePropagation === "function") {
            event.stopImmediatePropagation();
        }
        localState.selectedElement = target;
        state.selectedData = collectElementData(target);
        clearGroup();
        state.mode = "locked";
        updateHighlight(target);
        syncTopUI();
    }

    function handleViewportChange() {
        if (state.mode === "locked" && localState.selectedElement && document.documentElement.contains(localState.selectedElement)) {
            updateHighlight(localState.selectedElement);
            syncTopUI();
            return;
        }
        if (state.mode === "paused" && localState.selectedElement && document.documentElement.contains(localState.selectedElement)) {
            updateHighlight(localState.selectedElement);
            syncTopUI();
            return;
        }
        if (state.mode === "idle" && localState.hoverElement && document.documentElement.contains(localState.hoverElement)) {
            updateHighlight(localState.hoverElement);
            syncTopUI();
            return;
        }
        if (state.mode !== "locked") {
            updateHighlight(null);
            syncTopUI();
        }
    }

    function bindEvents() {
        if (localState.handlersBound && localState.boundDocument === document) {
            return;
        }
        if (localState.handlersBound && localState.boundDocument && localState.moveHandler) {
            try {
                localState.boundDocument.removeEventListener("mousemove", localState.moveHandler, true);
                localState.boundDocument.removeEventListener("click", localState.clickHandler, true);
                window.removeEventListener("scroll", localState.scrollHandler, true);
                window.removeEventListener("resize", localState.resizeHandler, true);
            } catch (error) {
            }
            localState.handlersBound = false;
        }

        localState.moveHandler = handleMove;
        localState.clickHandler = handleClick;
        localState.scrollHandler = handleViewportChange;
        localState.resizeHandler = handleViewportChange;
        document.addEventListener("mousemove", localState.moveHandler, true);
        document.addEventListener("click", localState.clickHandler, true);
        window.addEventListener("scroll", localState.scrollHandler, true);
        window.addEventListener("resize", localState.resizeHandler, true);
        localState.handlersBound = true;
        localState.boundDocument = document;
    }

    function bindWatchdog() {
        if (!isTopWindow || state.watchdogBound) {
            return;
        }
        const restoreUI = () => {
            try {
                ensureStyles();
                ensurePanel();
                ensureHighlight();
                syncTopUI();
                if (typeof topWindowRef.__ruyiXPathPickerInjectIntoFrames === "function") {
                    topWindowRef.__ruyiXPathPickerInjectIntoFrames();
                }
            } catch (error) {
            }
        };

        window.addEventListener("pageshow", restoreUI, true);
        window.addEventListener("load", restoreUI, true);
        document.addEventListener("visibilitychange", () => {
            if (!document.hidden) {
                restoreUI();
            }
        }, true);

        const observer = new MutationObserver(() => {
            const panelExists = !!document.getElementById(PANEL_ID);
            const highlightExists = !!document.getElementById(HIGHLIGHT_ID);
            const styleExists = !!document.getElementById("__ruyi_xpath_picker_style__");
            if (!panelExists || !highlightExists || !styleExists) {
                restoreUI();
            }
        });
        observer.observe(document.documentElement || document, {
            childList: true,
            subtree: true,
        });

        window.setInterval(() => {
            const panelExists = !!document.getElementById(PANEL_ID);
            const highlightExists = !!document.getElementById(HIGHLIGHT_ID);
            const styleExists = !!document.getElementById("__ruyi_xpath_picker_style__");
            if (!panelExists || !highlightExists || !styleExists) {
                restoreUI();
            }
        }, 1200);

        state.watchdogBound = true;
    }

    function init() {
        if (["idle", "locked", "paused"].indexOf(state.mode) === -1) {
            state.mode = "idle";
        }
        ensureStyles();
        if (isTopWindow) {
            ensurePanel();
        }
        ensureHighlight();
        bindEvents();
        bindWatchdog();
        syncTopUI();
    }

    window.__ruyiInitXPathPicker = init;
    topWindowRef.__ruyiInitXPathPicker = topWindowRef.__ruyiInitXPathPicker || init;
    init();
    return true;
}`
