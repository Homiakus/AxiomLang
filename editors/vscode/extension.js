const vscode = require('vscode');
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

let activePanel = null;
let debounceTimer = null;

function activate(context) {
    console.log('[AxiomLang] Extension activated.');

    // Register command to show flowchart panel
    const showCmd = vscode.commands.registerCommand('axiom.showFlowchart', () => {
        openFlowchartWebview(context);
    });

    // Register export command
    const exportCmd = vscode.commands.registerCommand('axiom.exportFlowchartSVG', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            vscode.window.showWarningMessage('No active Axiom document found.');
            return;
        }

        const lspPath = findLSPBinary(context);
        const code = editor.document.getText();

        try {
            const data = await requestFlowchartFromLSP(lspPath, code);
            if (!data.charts || data.charts.length === 0) {
                vscode.window.showInformationMessage('No functions or script blocks found in this Axiom file.');
                return;
            }

            const defaultUri = vscode.Uri.file(editor.document.fileName.replace(/\.axl$/, '_flow.svg'));
            const uri = await vscode.window.showSaveDialog({
                defaultUri: defaultUri,
                filters: { 'SVG Images': ['svg'] }
            });

            if (uri) {
                fs.writeFileSync(uri.fsPath, data.charts[0].svg, 'utf-8');
                vscode.window.showInformationMessage(`Flowchart saved to ${path.basename(uri.fsPath)}!`);
            }
        } catch (e) {
            vscode.window.showErrorMessage(`Failed to export flowchart: ${e.message}`);
        }
    });

    // React organically when user edits or changes active file
    const changeDocSub = vscode.workspace.onDidChangeTextDocument((event) => {
        const editor = vscode.window.activeTextEditor;
        if (editor && event.document === editor.document && isAxiomFile(event.document)) {
            triggerLiveFlowchartUpdate(context);
        }
    });

    const changeEditorSub = vscode.window.onDidChangeActiveTextEditor((editor) => {
        if (editor && isAxiomFile(editor.document) && activePanel) {
            triggerLiveFlowchartUpdate(context);
        }
    });

    // Highlight node when user cursor moves in code
    const cursorSub = vscode.window.onDidChangeTextEditorSelection((event) => {
        if (activePanel && isAxiomFile(event.textEditor.document)) {
            const pos = event.selections[0].active;
            activePanel.webview.postMessage({
                command: 'cursorMoved',
                line: pos.line + 1,
                character: pos.character + 1
            });
        }
    });

    context.subscriptions.push(showCmd, exportCmd, changeDocSub, changeEditorSub, cursorSub);
}

function isAxiomFile(doc) {
    if (!doc) return false;
    return doc.languageId === 'axiom' || doc.fileName.endsWith('.axl') || doc.fileName.endsWith('.axiom');
}

function findLSPBinary(context) {
    const configPath = vscode.workspace.getConfiguration('axiom').get('lspPath');
    if (configPath && fs.existsSync(configPath)) {
        return configPath;
    }

    // Try workspace bin/axiom_lsp
    const folders = vscode.workspace.workspaceFolders;
    if (folders && folders.length > 0) {
        const wsBin = path.join(folders[0].uri.fsPath, 'bin', 'axiom_lsp');
        if (fs.existsSync(wsBin)) {
            return wsBin;
        }
    }

    // Fallback to /home/den/main/30-39 Development/AxiomLang/bin/axiom_lsp
    const defaultAxiomBin = '/home/den/main/30-39 Development/AxiomLang/bin/axiom_lsp';
    if (fs.existsSync(defaultAxiomBin)) {
        return defaultAxiomBin;
    }

    return 'axiom_lsp';
}

function requestFlowchartFromLSP(lspBinary, sourceCode) {
    return new Promise((resolve, reject) => {
        let child;
        try {
            child = spawn(lspBinary, [], { stdio: ['pipe', 'pipe', 'pipe'] });
        } catch (err) {
            return reject(err);
        }

        const req = {
            jsonrpc: "2.0",
            id: 1,
            method: "axiom/flowchart",
            params: { text: sourceCode }
        };

        const jsonStr = JSON.stringify(req);
        const packet = `Content-Length: ${Buffer.byteLength(jsonStr, 'utf8')}\r\n\r\n${jsonStr}`;

        let outputBuffer = Buffer.alloc(0);

        child.stdout.on('data', (chunk) => {
            outputBuffer = Buffer.concat([outputBuffer, chunk]);
            const headerEnd = outputBuffer.indexOf('\r\n\r\n');
            if (headerEnd !== -1) {
                const header = outputBuffer.slice(0, headerEnd).toString('utf8');
                const m = header.match(/Content-Length:\s*(\d+)/i);
                if (m) {
                    const cl = parseInt(m[1], 10);
                    const bodyStart = headerEnd + 4;
                    if (outputBuffer.length >= bodyStart + cl) {
                        const body = outputBuffer.slice(bodyStart, bodyStart + cl).toString('utf8');
                        child.kill();
                        try {
                            const res = JSON.parse(body);
                            if (res.error) {
                                reject(new Error(res.error.message || 'LSP error'));
                            } else {
                                resolve(res.result);
                            }
                        } catch (e) {
                            reject(e);
                        }
                    }
                }
            }
        });

        child.stderr.on('data', (errData) => {
            console.error('[Axiom LSP Stderr]:', errData.toString());
        });

        child.on('error', (err) => {
            reject(err);
        });

        child.on('close', (code) => {
            if (code !== 0 && outputBuffer.length === 0) {
                reject(new Error(`axiom_lsp exited with code ${code}`));
            }
        });

        child.stdin.write(packet);
        child.stdin.end();
    });
}

function triggerLiveFlowchartUpdate(context) {
    const delay = vscode.workspace.getConfiguration('axiom').get('liveDiagramDelay') || 300;
    if (debounceTimer) {
        clearTimeout(debounceTimer);
    }

    debounceTimer = setTimeout(async () => {
        if (!activePanel) return;

        const editor = vscode.window.activeTextEditor;
        if (!editor || !isAxiomFile(editor.document)) return;

        const lspPath = findLSPBinary(context);
        const code = editor.document.getText();

        try {
            const data = await requestFlowchartFromLSP(lspPath, code);
            if (activePanel) {
                activePanel.webview.postMessage({
                    command: 'updateDiagrams',
                    data: data,
                    fileName: path.basename(editor.document.fileName)
                });
            }
        } catch (e) {
            console.warn('[Axiom Flowchart] Update failed:', e);
        }
    }, delay);
}

function openFlowchartWebview(context) {
    if (activePanel) {
        activePanel.reveal(vscode.ViewColumn.Beside);
        triggerLiveFlowchartUpdate(context);
        return;
    }

    activePanel = vscode.window.createWebviewPanel(
        'axiomFlowchart',
        '⚡ Axiom Live Flowchart',
        vscode.ViewColumn.Beside,
        {
            enableScripts: true,
            retainContextWhenHidden: true
        }
    );

    activePanel.webview.html = getWebviewContent();

    activePanel.webview.onDidReceiveMessage(
        message => {
            switch (message.command) {
                case 'jumpToCode':
                    jumpToCodeLocation(message.line, message.col);
                    break;
                case 'requestRefresh':
                    triggerLiveFlowchartUpdate(context);
                    break;
            }
        },
        undefined,
        context.subscriptions
    );

    activePanel.onDidDispose(() => {
        activePanel = null;
    }, null, context.subscriptions);

    triggerLiveFlowchartUpdate(context);
}

function jumpToCodeLocation(line, col) {
    const editor = vscode.window.activeTextEditor;
    if (!editor || !line) return;

    const zeroLine = Math.max(0, line - 1);
    const zeroCol = Math.max(0, (col || 1) - 1);
    const pos = new vscode.Position(zeroLine, zeroCol);
    editor.selection = new vscode.Selection(pos, pos);
    editor.revealRange(new vscode.Range(pos, pos), vscode.TextEditorRevealType.InCenter);
}

function getWebviewContent() {
    return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Axiom Live Flowchart</title>
<style>
  :root {
    --bg-color: #0b0f19;
    --card-bg: #111827;
    --accent-blue: #38bdf8;
    --accent-green: #34d399;
    --accent-red: #f87171;
    --text-primary: #f8fafc;
    --text-secondary: #94a3b8;
    --border-color: #1e293b;
  }
  body {
    margin: 0;
    padding: 0;
    background-color: var(--bg-color);
    color: var(--text-primary);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
  }
  header {
    background: var(--card-bg);
    border-bottom: 1px solid var(--border-color);
    padding: 12px 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  }
  .title-group {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .status-badge {
    background: rgba(56, 189, 248, 0.15);
    color: var(--accent-blue);
    border: 1px solid rgba(56, 189, 248, 0.3);
    border-radius: 9999px;
    padding: 2px 10px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.5px;
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .pulse-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent-blue);
    box-shadow: 0 0 8px var(--accent-blue);
    animation: pulse 2s infinite ease-in-out;
  }
  @keyframes pulse {
    0%, 100% { opacity: 0.4; transform: scale(0.9); }
    50% { opacity: 1; transform: scale(1.3); }
  }
  .tabs {
    display: flex;
    gap: 8px;
    padding: 8px 20px;
    background: rgba(17, 24, 39, 0.6);
    border-bottom: 1px solid var(--border-color);
    overflow-x: auto;
  }
  .tab-btn {
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-secondary);
    padding: 4px 12px;
    border-radius: 6px;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }
  .tab-btn.active {
    background: rgba(56, 189, 248, 0.12);
    color: var(--accent-blue);
    border-color: rgba(56, 189, 248, 0.4);
    font-weight: 600;
  }
  #viewer-container {
    flex: 1;
    overflow: auto;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding: 30px;
    position: relative;
  }
  .diagram-wrapper {
    background: radial-gradient(circle at 50% 0%, rgba(30, 41, 59, 0.5) 0%, rgba(15, 23, 42, 0.2) 100%);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 20px 40px rgba(0,0,0,0.5);
    max-width: 100%;
    transition: transform 0.2s ease;
  }
  .diagram-wrapper svg {
    display: block;
    max-width: 100%;
    height: auto;
  }
  .node-active rect, .node-active polygon, .node-active path {
    filter: drop-shadow(0 0 12px #38bdf8) !important;
    stroke: #38bdf8 !important;
    stroke-width: 3px !important;
  }
  .svg-clickable {
    cursor: pointer;
    transition: transform 0.15s ease;
  }
  .svg-clickable:hover {
    filter: brightness(1.2);
  }
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-secondary);
    text-align: center;
  }
  .controls {
    display: flex;
    gap: 8px;
  }
  .btn-control {
    background: #1e293b;
    border: 1px solid #334155;
    color: var(--text-primary);
    padding: 4px 10px;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
  }
  .btn-control:hover {
    background: #334155;
  }
</style>
</head>
<body>
<header>
  <div class="title-group">
    <div class="status-badge"><span class="pulse-dot"></span> LIVE CFG</div>
    <span id="doc-title" style="font-weight: 600; font-size: 13px;">No file open</span>
  </div>
  <div class="controls">
    <button class="btn-control" onclick="zoomIn()">Zoom +</button>
    <button class="btn-control" onclick="zoomOut()">Zoom -</button>
    <button class="btn-control" onclick="resetZoom()">Reset</button>
    <button class="btn-control" onclick="refreshDiagram()">Refresh</button>
  </div>
</header>
<div class="tabs" id="chart-tabs"></div>
<div id="viewer-container">
  <div id="empty-view" class="empty-state">
    <h3>⚡ Organically Building Flowchart...</h3>
    <p>Write Axiom functions or statements to see real-time control flow diagrams.</p>
  </div>
  <div id="diagram-view" class="diagram-wrapper" style="display: none;"></div>
</div>

<script>
  const vscode = acquireVsCodeApi();
  let currentCharts = [];
  let activeChartIndex = 0;
  let zoomLevel = 1.0;
  let activeCursorLine = null;

  function refreshDiagram() {
    vscode.postMessage({ command: 'requestRefresh' });
  }

  function zoomIn() {
    zoomLevel += 0.15;
    applyZoom();
  }
  function zoomOut() {
    if (zoomLevel > 0.3) zoomLevel -= 0.15;
    applyZoom();
  }
  function resetZoom() {
    zoomLevel = 1.0;
    applyZoom();
  }
  function applyZoom() {
    const view = document.getElementById('diagram-view');
    view.style.transform = 'scale(' + zoomLevel + ')';
    view.style.transformOrigin = 'top center';
  }

  window.addEventListener('message', event => {
    const msg = event.data;
    switch (msg.command) {
      case 'updateDiagrams':
        handleUpdateDiagrams(msg.data, msg.fileName);
        break;
      case 'cursorMoved':
        highlightNodeForLine(msg.line);
        break;
    }
  });

  function handleUpdateDiagrams(data, fileName) {
    if (fileName) {
      document.getElementById('doc-title').textContent = fileName;
    }
    if (!data || !data.charts || data.charts.length === 0) {
      document.getElementById('empty-view').style.display = 'flex';
      document.getElementById('diagram-view').style.display = 'none';
      document.getElementById('chart-tabs').innerHTML = '';
      return;
    }

    currentCharts = data.charts;
    if (activeChartIndex >= currentCharts.length) {
      activeChartIndex = 0;
    }

    renderTabs();
    renderCurrentChart();
  }

  function renderTabs() {
    const tabsContainer = document.getElementById('chart-tabs');
    tabsContainer.innerHTML = '';
    currentCharts.forEach((c, idx) => {
      const btn = document.createElement('button');
      btn.className = 'tab-btn' + (idx === activeChartIndex ? ' active' : '');
      btn.textContent = c.title || ('Chart ' + (idx + 1));
      btn.onclick = () => {
        activeChartIndex = idx;
        renderTabs();
        renderCurrentChart();
      };
      tabsContainer.appendChild(btn);
    });
  }

  function renderCurrentChart() {
    const chart = currentCharts[activeChartIndex];
    if (!chart) return;

    document.getElementById('empty-view').style.display = 'none';
    const diagramView = document.getElementById('diagram-view');
    diagramView.style.display = 'block';
    diagramView.innerHTML = chart.svg;

    // Attach click and hover behaviors on SVG elements
    const svgEl = diagramView.querySelector('svg');
    if (svgEl && chart.nodes) {
      chart.nodes.forEach(node => {
        const textElements = svgEl.querySelectorAll('text');
        textElements.forEach(t => {
          if (node.label && t.textContent.includes(node.label.slice(0, 15))) {
            const group = t.previousElementSibling;
            if (group) {
              group.classList.add('svg-clickable');
              group.title = 'Click to jump to line ' + (node.line || '');
              group.onclick = () => {
                if (node.line) {
                  vscode.postMessage({ command: 'jumpToCode', line: node.line, col: node.col });
                }
              };
            }
          }
        });
      });
    }

    if (activeCursorLine) {
      highlightNodeForLine(activeCursorLine);
    }
  }

  function highlightNodeForLine(line) {
    activeCursorLine = line;
    const chart = currentCharts[activeChartIndex];
    if (!chart || !chart.nodes) return;

    const matchedNode = chart.nodes.find(n => n.line === line);
    const diagramView = document.getElementById('diagram-view');
    const allClickable = diagramView.querySelectorAll('.svg-clickable');
    allClickable.forEach(el => el.classList.remove('node-active'));

    if (matchedNode) {
      const svgEl = diagramView.querySelector('svg');
      if (svgEl) {
        const textElements = svgEl.querySelectorAll('text');
        textElements.forEach(t => {
          if (matchedNode.label && t.textContent.includes(matchedNode.label.slice(0, 15))) {
            const group = t.previousElementSibling;
            if (group) {
              group.classList.add('node-active');
            }
          }
        });
      }
    }
  }
</script>
</body>
</html>`;
}

function deactivate() {
    if (activePanel) {
        activePanel.dispose();
    }
}

module.exports = {
    activate,
    deactivate
};
