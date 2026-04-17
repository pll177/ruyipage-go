package pages

const actionVisualScript = `() => {
    if (typeof window === "undefined" || typeof document === "undefined") {
        return;
    }
    if (window.__ruyiAV) {
        return;
    }

    const CANVAS_ID = "__ruyi_av_canvas__";
    const DOT_ID = "__ruyi_av_dot__";
    const COORD_ID = "__ruyi_av_coord__";
    const HIGHLIGHT_ID = "__ruyi_av_highlight__";

    let dot = document.getElementById(DOT_ID);
    if (!dot) {
        dot = document.createElement("div");
        dot.id = DOT_ID;
        dot.style.cssText = "position:fixed;width:14px;height:14px;border-radius:50%;" +
            "background:rgba(255,50,50,0.5);border:2px solid rgba(255,50,50,0.85);" +
            "pointer-events:none;z-index:2147483646;transform:translate(-50%,-50%);display:none;";
        document.documentElement.appendChild(dot);
    }

    let coord = document.getElementById(COORD_ID);
    if (!coord) {
        coord = document.createElement("div");
        coord.id = COORD_ID;
        coord.style.cssText = "position:fixed;pointer-events:none;z-index:2147483646;" +
            "font:11px/1.2 monospace;color:#fff;background:rgba(0,0,0,0.65);" +
            "padding:2px 6px;border-radius:3px;display:none;white-space:nowrap;";
        document.documentElement.appendChild(coord);
    }

    let highlight = document.getElementById(HIGHLIGHT_ID);
    if (!highlight) {
        highlight = document.createElement("div");
        highlight.id = HIGHLIGHT_ID;
        highlight.style.cssText = "position:fixed;display:none;pointer-events:none;z-index:2147483646;" +
            "border:3px solid rgba(255,205,86,0.98);border-radius:10px;" +
            "background:rgba(255,205,86,0.12);box-shadow:0 0 0 2px rgba(255,255,255,0.20),0 14px 32px rgba(255,205,86,0.28);";
        document.documentElement.appendChild(highlight);
    }
    const highlightLabel = document.createElement("div");
    highlightLabel.style.cssText = "position:absolute;left:0;top:-28px;padding:3px 8px;border-radius:999px;" +
        "font:11px/1.2 monospace;color:#111827;background:rgba(255,205,86,0.96);white-space:nowrap;";
    highlight.appendChild(highlightLabel);
    let highlightTimer = null;

    let canvas = document.getElementById(CANVAS_ID);
    if (!canvas) {
        canvas = document.createElement("canvas");
        canvas.id = CANVAS_ID;
        canvas.style.cssText = "position:fixed;top:0;left:0;width:100%;height:100%;" +
            "pointer-events:none;z-index:2147483645;";
        document.documentElement.appendChild(canvas);
    }
    const resizeCanvas = () => {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
    };
    resizeCanvas();
    window.addEventListener("resize", resizeCanvas);
    const ctx = canvas.getContext("2d");

    let trail = [];
    const MAX_TRAIL = 200;
    let fadeTimer = null;
    let fadeOpacity = 1.0;
    let moveQueue = [];
    let moveRaf = 0;

    const drawTrail = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        if (trail.length < 2) {
            return;
        }
        const len = trail.length;
        ctx.lineWidth = 5;
        ctx.lineCap = "round";
        ctx.lineJoin = "round";
        for (let i = 1; i < len; i += 1) {
            const age = 1 - i / len;
            const alpha = (0.15 + 0.55 * (i / len)) * fadeOpacity;
            const r = Math.round(80 + 175 * age);
            const g = Math.round(60 + 140 * (1 - age));
            const b = Math.round(200 + 55 * (1 - age));
            ctx.strokeStyle = "rgba(" + r + "," + g + "," + b + "," + alpha.toFixed(2) + ")";
            ctx.beginPath();
            ctx.moveTo(trail[i - 1][0], trail[i - 1][1]);
            ctx.lineTo(trail[i][0], trail[i][1]);
            ctx.stroke();
        }
    };

    const startFadeOut = () => {
        if (fadeTimer) {
            clearInterval(fadeTimer);
        }
        fadeOpacity = 1.0;
        fadeTimer = setInterval(() => {
            fadeOpacity -= 0.025;
            if (fadeOpacity <= 0) {
                fadeOpacity = 0;
                clearInterval(fadeTimer);
                fadeTimer = null;
                trail = [];
                ctx.clearRect(0, 0, canvas.width, canvas.height);
                return;
            }
            drawTrail();
        }, 40);
    };

    const moveDot = (x, y) => {
        dot.style.left = x + "px";
        dot.style.top = y + "px";
        dot.style.display = "block";
        coord.style.left = x + 16 + "px";
        coord.style.top = y + 16 + "px";
        coord.textContent = "(" + x + ", " + y + ")";
        coord.style.display = "block";
    };

    const pumpMoves = () => {
        moveRaf = 0;
        if (!moveQueue.length) {
            startFadeOut();
            return;
        }
        if (fadeTimer) {
            clearInterval(fadeTimer);
            fadeTimer = null;
        }
        fadeOpacity = 1.0;

        const batch = Math.min(moveQueue.length, 3);
        for (let i = 0; i < batch; i += 1) {
            const pt = moveQueue.shift();
            trail.push(pt);
            if (trail.length > MAX_TRAIL) {
                trail.shift();
            }
            moveDot(pt[0], pt[1]);
        }
        drawTrail();
        moveRaf = window.requestAnimationFrame(pumpMoves);
    };

    const renderMoves = (points) => {
        if (!points || !points.length) {
            return;
        }
        for (let i = 0; i < points.length; i += 1) {
            moveQueue.push(points[i]);
        }
        if (!moveRaf) {
            moveRaf = window.requestAnimationFrame(pumpMoves);
        }
    };

    const renderClick = (x, y, button) => {
        const color = button === 2 ? "255,60,60" : button === 1 ? "60,60,255" : "60,200,60";
        moveDot(x, y);

        const ring = document.createElement("div");
        ring.style.cssText = "position:fixed;pointer-events:none;z-index:2147483647;" +
            "border:3px solid rgba(" + color + ",0.92);border-radius:50%;" +
            "width:14px;height:14px;left:" + x + "px;top:" + y + "px;" +
            "transform:translate(-50%,-50%);";
        document.documentElement.appendChild(ring);
        const ch = document.createElement("div");
        ch.style.cssText = "position:fixed;pointer-events:none;z-index:2147483647;" +
            "left:" + (x - 18) + "px;top:" + (y - 1) + "px;width:36px;height:2px;" +
            "background:rgba(" + color + ",0.68);";
        const cv = document.createElement("div");
        cv.style.cssText = "position:fixed;pointer-events:none;z-index:2147483647;" +
            "left:" + (x - 1) + "px;top:" + (y - 18) + "px;width:2px;height:36px;" +
            "background:rgba(" + color + ",0.68);";
        document.documentElement.appendChild(ch);
        document.documentElement.appendChild(cv);

        let size = 14;
        let opacity = 0.92;
        const anim = setInterval(() => {
            size += 3.2;
            opacity -= 0.03;
            if (opacity <= 0) {
                clearInterval(anim);
                ring.remove();
                ch.remove();
                cv.remove();
                return;
            }
            ring.style.width = size + "px";
            ring.style.height = size + "px";
            ring.style.borderColor = "rgba(" + color + "," + opacity.toFixed(2) + ")";
        }, 20);
    };

    const renderHighlight = (rect, label) => {
        if (!rect) {
            return;
        }
        highlight.style.display = "block";
        highlight.style.left = Math.max(0, rect.x) + "px";
        highlight.style.top = Math.max(0, rect.y) + "px";
        highlight.style.width = Math.max(0, rect.width) + "px";
        highlight.style.height = Math.max(0, rect.height) + "px";
        highlightLabel.textContent = label || "target";
        if (highlightTimer) {
            clearTimeout(highlightTimer);
        }
        highlightTimer = setTimeout(() => {
            highlight.style.display = "none";
        }, 900);
    };

    window.__ruyiAV = {
        moves: renderMoves,
        click: renderClick,
        highlight: renderHighlight,
    };
}`
