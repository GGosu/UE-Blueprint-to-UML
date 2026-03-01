mermaid.initialize({startOnLoad: false, theme: 'dark'});

// ── Viewer pan/zoom state ─────────────────────────────────────────────────────
let scale = 1, panX = 0, panY = 0;
let isDragging = false, dragStartX, dragStartY, dragStartPanX, dragStartPanY;

function applyTransform() {
    const area = document.getElementById('viewer-zoom-area');
    if (area) area.style.transform = `translate(${panX}px, ${panY}px) scale(${scale})`;
}

function resetTransform() {
    scale = 1;
    panX = 0;
    panY = 0;
    const area = document.getElementById('viewer-zoom-area');
    if (area) area.style.transform = 'none';
}

function fitDiagram() {
    const canvas = document.getElementById('viewer-canvas');
    const area = document.getElementById('viewer-zoom-area');
    const svg = area?.querySelector('svg');

    if (!canvas || !area) return;
    area.style.transform = 'none';
    let {width: w, height: h} = area.getBoundingClientRect();

    if ((!w || !h) && svg) {
        const vb = svg.viewBox?.baseVal;
        const mw = parseFloat(svg.style.maxWidth);
        if (mw > 0 && vb?.width > 0 && vb?.height > 0) {
            w = mw;
            h = mw * (vb.height / vb.width);
        } else if (vb?.width > 0 && vb?.height > 0) {
            w = vb.width;
            h = vb.height;
        }
    }

    if (!w || !h) return;

    const pad = 32;
    scale = Math.min(
        (canvas.clientWidth - pad) / w,
        (canvas.clientHeight - pad) / h
    );
    panX = (canvas.clientWidth - w * scale) / 2;
    panY = (canvas.clientHeight - h * scale) / 2;
    applyTransform();
}

function zoomIn() {
    scale = Math.min(scale * 1.25, 30);
    applyTransform();
}

function zoomOut() {
    scale = Math.max(scale / 1.25, 0.04);
    applyTransform();
}

function viewerFullscreen() {
    const pane = document.getElementById('viewer-pane');
    if (!document.fullscreenElement) {
        pane?.requestFullscreen();
    } else {
        document.exitFullscreen();
    }
}

document.addEventListener('fullscreenchange', () => setTimeout(fitDiagram, 80));

// ── Convert ───────────────────────────────────────────────────────────────────
async function convertBlueprint(text) {
    const spinner = document.getElementById('topbar-spinner');
    if (spinner) spinner.classList.add('active');

    try {
        const resp = await fetch('/convert', {
            method: 'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body: 'blueprint=' + encodeURIComponent(text),
        });
        const data = await resp.json();

        if (data.error) {
            showCodeError(data.error);
            resetTransform();
            document.getElementById('result').innerHTML = '';
            document.getElementById('empty-hint').style.display = 'flex';
            return;
        }

        showCode(data.mermaid);
        await renderDiagram(data.mermaid);
    } catch (e) {
        showCodeError('Network error: ' + e.message);
    } finally {
        if (spinner) spinner.classList.remove('active');
    }
}

// ── Render diagram ────────────────────────────────────────────────────────────
async function renderDiagram(mermaidSrc) {
    resetTransform();

    const result = document.getElementById('result');
    const hint = document.getElementById('empty-hint');

    result.innerHTML = `<pre class="mermaid">${escHtml(mermaidSrc)}</pre>`;
    if (hint) hint.style.display = 'none';

    await mermaid.run({nodes: result.querySelectorAll('.mermaid')});

    await new Promise(r => requestAnimationFrame(r));
    fitDiagram();
}

// ── Code view (left panel) ────────────────────────────────────────────────────
function showCode(src) {
    const view = document.getElementById('code-view');
    const lines = src.split('\n');
    const nums = lines.map((_, i) => `<span>${i + 1}</span>`).join('');
    view.innerHTML =
        `<div class="code-gutter">${nums}</div>` +
        `<pre class="code-content">${escHtml(src)}</pre>`;
    view.dataset.src = src;
}

function showCodeError(msg) {
    const view = document.getElementById('code-view');
    view.innerHTML = `<p class="error">${escHtml(msg)}</p>`;
    view.dataset.src = '';
}

function escHtml(s) {
    return s.replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

// ── Global Ctrl+V → paste Blueprint ──────────────────────────────────────────
document.addEventListener('paste', (e) => {
    const tag = document.activeElement?.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA') return;
    const text = e.clipboardData.getData('text/plain').trim();
    if (!text) return;
    e.preventDefault();
    convertBlueprint(text);
});

// ── Global Ctrl+C → copy Mermaid source ──────────────────────────────────────
document.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'c') {
        if (!window.getSelection()?.toString()) {
            e.preventDefault();
            copyMermaidSource();
        }
    }
});

// ── Wheel → zoom toward cursor ────────────────────────────────────────────────
const viewerCanvas = document.getElementById('viewer-canvas');

viewerCanvas.addEventListener('wheel', function (e) {
    e.preventDefault();
    const factor = e.deltaY < 0 ? 1.12 : 1 / 1.12;
    const newScale = Math.min(Math.max(scale * factor, 0.04), 30);
    const rect = this.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    panX = mx - (mx - panX) * (newScale / scale);
    panY = my - (my - panY) * (newScale / scale);
    scale = newScale;
    applyTransform();
}, {passive: false});

// ── Mouse drag → pan ──────────────────────────────────────────────────────────
viewerCanvas.addEventListener('mousedown', function (e) {
    if (e.button !== 0) return;
    e.preventDefault();
    isDragging = true;
    dragStartX = e.clientX;
    dragStartY = e.clientY;
    dragStartPanX = panX;
    dragStartPanY = panY;
    this.classList.add('dragging');
});

document.addEventListener('mousemove', function (e) {
    if (!isDragging) return;
    panX = dragStartPanX + (e.clientX - dragStartX);
    panY = dragStartPanY + (e.clientY - dragStartY);
    applyTransform();
});

document.addEventListener('mouseup', function () {
    if (!isDragging) return;
    isDragging = false;
    viewerCanvas.classList.remove('dragging');
});

// ── Download SVG ──────────────────────────────────────────────────────────────
function downloadDiagramSvg() {
    const svg = document.querySelector('#result svg');
    if (!svg) {
        alert('No diagram rendered yet.');
        return;
    }
    const clone = svg.cloneNode(true);
    clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg');

    const bg = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
    bg.setAttribute('width', '100%');
    bg.setAttribute('height', '100%');
    bg.setAttribute('fill', '#0b0b18');
    clone.insertBefore(bg, clone.firstChild);

    const blob = new Blob([new XMLSerializer().serializeToString(clone)],
        {type: 'image/svg+xml'});
    const a = document.createElement('a');
    a.download = 'blueprint-diagram.svg';
    a.href = URL.createObjectURL(blob);
    a.click();
}

// ── Copy Mermaid source ───────────────────────────────────────────────────────
async function copyMermaidSource() {
    const src = document.getElementById('code-view')?.dataset?.src?.trim();
    if (!src) {
        alert('No diagram yet — paste a Blueprint first.');
        return;
    }

    try {
        await navigator.clipboard.writeText(src);
    } catch {
        const ta = document.createElement('textarea');
        ta.value = src;
        ta.style.cssText = 'position:fixed;opacity:0;pointer-events:none';
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
    }

    const btn = document.getElementById('copy-source-btn');
    if (btn) {
        const orig = btn.textContent;
        btn.textContent = '✓ Copied';
        setTimeout(() => {
            btn.textContent = orig;
        }, 1500);
    }
}
