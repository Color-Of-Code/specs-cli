// Visualize webview rendering the traceability graph as
// Mermaid via the bundled mermaid.min.js.
import * as path from "path";
import * as fs from "fs";
import * as vscode from "vscode";
import { runAndCapture, findSpecsFolder, findSpecsRoot, getOutput } from "./engine";

let currentPanel: vscode.WebviewPanel | undefined;

export function registerVisualizePanel(context: vscode.ExtensionContext): void {
  context.subscriptions.push(
    vscode.commands.registerCommand("specs.visualize.preview", () => openOrRefresh(context)),
    vscode.commands.registerCommand("specs.visualize.refresh", () => refresh(context)),
  );

  // Auto-refresh when model files change while panel is open.
  const folder = findSpecsFolder();
  if (folder) {
    const root = findSpecsRoot(folder) ?? folder.uri.fsPath;
    const watcher = vscode.workspace.createFileSystemWatcher(
      new vscode.RelativePattern(root, "model/**/*.md"),
    );
    const debounced = debounce(() => {
      if (currentPanel) {
        refresh(context);
      }
    }, 500);
    watcher.onDidCreate(debounced);
    watcher.onDidChange(debounced);
    watcher.onDidDelete(debounced);
    context.subscriptions.push(watcher);
  }
}

async function openOrRefresh(context: vscode.ExtensionContext): Promise<void> {
  if (currentPanel) {
    currentPanel.reveal(vscode.ViewColumn.Beside);
    await refresh(context);
    return;
  }
  const folder = findSpecsFolder();
  if (!folder) {
    vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    return;
  }
  const mediaRoot = vscode.Uri.joinPath(context.extensionUri, "media");
  const panel = vscode.window.createWebviewPanel(
    "specs.visualize",
    "Specs: Traceability",
    vscode.ViewColumn.Beside,
    {
      enableScripts: true,
      retainContextWhenHidden: true,
      localResourceRoots: [mediaRoot],
    },
  );
  currentPanel = panel;
  panel.onDidDispose(() => {
    currentPanel = undefined;
  });
  panel.webview.onDidReceiveMessage(async (msg: { type: string; payload?: string }) => {
    if (msg.type === "export-dot") {
      await exportFile(context, "dot");
    } else if (msg.type === "export-mermaid") {
      await exportFile(context, "mermaid");
    } else if (msg.type === "refresh") {
      await refresh(context);
    }
  });
  await refresh(context);
}

async function refresh(context: vscode.ExtensionContext): Promise<void> {
  if (!currentPanel) {
    return;
  }
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  const res = await runAndCapture(
    context,
    ["visualize", "traceability", "--format", "mermaid", "--out", "-"],
    cwd,
  );
  if (res.exitCode !== 0) {
    getOutput().appendLine(res.stderr);
    currentPanel.webview.html = errorHtml(res.stderr || "visualize failed");
    return;
  }
  currentPanel.webview.html = renderHtml(currentPanel.webview, context, res.stdout);
}

async function exportFile(
  context: vscode.ExtensionContext,
  format: "dot" | "mermaid",
): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  const ext = format === "dot" ? "dot" : "md";
  const target = await vscode.window.showSaveDialog({
    defaultUri: vscode.Uri.joinPath(folder.uri, `traceability.${ext}`),
    filters: format === "dot" ? { "Graphviz DOT": ["dot"] } : { "Mermaid markdown": ["md", "mmd"] },
  });
  if (!target) {
    return;
  }
  const res = await runAndCapture(
    context,
    ["visualize", "traceability", "--format", format, "--out", target.fsPath],
    cwd,
  );
  if (res.exitCode !== 0) {
    vscode.window.showErrorMessage(`visualize failed: ${res.stderr}`);
    return;
  }
  vscode.window.showInformationMessage(`Wrote ${target.fsPath}`);
}

function renderHtml(
  webview: vscode.Webview,
  context: vscode.ExtensionContext,
  mermaidSrc: string,
): string {
  const mediaRoot = vscode.Uri.joinPath(context.extensionUri, "media");
  const mermaidUri = webview.asWebviewUri(vscode.Uri.joinPath(mediaRoot, "mermaid.min.js"));
  const nonce = randomNonce();
  const csp = [
    `default-src 'none'`,
    `style-src ${webview.cspSource} 'unsafe-inline'`,
    `script-src 'nonce-${nonce}'`,
    `img-src ${webview.cspSource} data:`,
    `font-src ${webview.cspSource}`,
  ].join("; ");
  const safeMermaid = escapeHtml(mermaidSrc);
  const fallbackInline = !fs.existsSync(
    path.join(context.extensionPath, "media", "mermaid.min.js"),
  );
  const fallbackBanner = fallbackInline
    ? `<div class="banner">mermaid.min.js not bundled \u2014 the diagram will not render. See extension/README for setup.</div>`
    : "";

  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta http-equiv="Content-Security-Policy" content="${csp}">
<style>
  body { font-family: var(--vscode-font-family); padding: 12px; color: var(--vscode-foreground); }
  .toolbar { margin-bottom: 12px; display: flex; gap: 8px; }
  button { background: var(--vscode-button-background); color: var(--vscode-button-foreground); border: none; padding: 4px 12px; cursor: pointer; }
  button:hover { background: var(--vscode-button-hoverBackground); }
  .banner { background: var(--vscode-inputValidation-warningBackground); color: var(--vscode-inputValidation-warningForeground); padding: 6px 10px; margin-bottom: 12px; }
  pre.source { display: none; }
  .mermaid svg { max-width: 100%; height: auto; }
</style>
</head>
<body>
${fallbackBanner}
<div class="toolbar">
  <button id="refresh">Refresh</button>
  <button id="export-mermaid">Export Mermaid</button>
  <button id="export-dot">Export DOT</button>
</div>
<div class="mermaid">
${safeMermaid}
</div>
<pre class="source" id="source">${safeMermaid}</pre>
${fallbackInline ? "" : `<script nonce="${nonce}" src="${mermaidUri}"></script>`}
<script nonce="${nonce}">
  const vscode = acquireVsCodeApi();
  document.getElementById('refresh').addEventListener('click', () => vscode.postMessage({ type: 'refresh' }));
  document.getElementById('export-mermaid').addEventListener('click', () => vscode.postMessage({ type: 'export-mermaid' }));
  document.getElementById('export-dot').addEventListener('click', () => vscode.postMessage({ type: 'export-dot' }));
  if (typeof mermaid !== 'undefined') {
    mermaid.initialize({ startOnLoad: true, theme: 'default', securityLevel: 'strict' });
  }
</script>
</body>
</html>`;
}

function errorHtml(message: string): string {
  return `<!doctype html><html><body><pre style="color: var(--vscode-errorForeground)">${escapeHtml(
    message,
  )}</pre></body></html>`;
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function randomNonce(): string {
  let s = "";
  const cs = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < 24; i++) {
    s += cs.charAt(Math.floor(Math.random() * cs.length));
  }
  return s;
}

function debounce<T extends (...args: unknown[]) => void>(fn: T, ms: number): T {
  let t: NodeJS.Timeout | undefined;
  return ((...args: Parameters<T>) => {
    if (t) {
      clearTimeout(t);
    }
    t = setTimeout(() => fn(...args), ms);
  }) as T;
}
