// Change-Requests tree view.
import * as path from "path";
import * as fs from "fs";
import * as vscode from "vscode";
import { runAndCapture, findSpecsFolder, findSpecsRoot, getOutput } from "./engine";

interface CRRecord {
  id: string;
  slug: string;
  dir: string;
  has_index: boolean;
  requirements: number;
  features: number;
  components: number;
  architecture: number;
}

type Node = CRNode | FileNode | StatNode;

class CRNode {
  readonly kind = "cr" as const;
  constructor(public readonly cr: CRRecord) {}
}

class FileNode {
  readonly kind = "file" as const;
  constructor(
    public readonly cr: CRRecord,
    public readonly fullPath: string,
    public readonly relPath: string,
  ) {}
}

class StatNode {
  readonly kind = "stat" as const;
  constructor(
    public readonly cr: CRRecord,
    public readonly label: string,
  ) {}
}

export class CRTreeProvider implements vscode.TreeDataProvider<Node> {
  private _onDidChange = new vscode.EventEmitter<Node | void>();
  readonly onDidChangeTreeData = this._onDidChange.event;

  constructor(private readonly context: vscode.ExtensionContext) {}

  refresh(): void {
    this._onDidChange.fire();
  }

  getTreeItem(node: Node): vscode.TreeItem {
    switch (node.kind) {
      case "cr": {
        const item = new vscode.TreeItem(
          `${node.cr.id} — ${node.cr.slug}`,
          vscode.TreeItemCollapsibleState.Collapsed,
        );
        item.description = countsLabel(node.cr);
        item.iconPath = new vscode.ThemeIcon("git-pull-request");
        item.contextValue = "specs.cr";
        item.tooltip = node.cr.dir;
        return item;
      }
      case "file": {
        const item = new vscode.TreeItem(node.relPath, vscode.TreeItemCollapsibleState.None);
        item.resourceUri = vscode.Uri.file(node.fullPath);
        item.command = {
          command: "vscode.open",
          title: "Open",
          arguments: [item.resourceUri],
        };
        item.contextValue = "specs.cr.file";
        return item;
      }
      case "stat": {
        const item = new vscode.TreeItem(node.label, vscode.TreeItemCollapsibleState.None);
        item.iconPath = new vscode.ThemeIcon("info");
        return item;
      }
    }
  }

  async getChildren(node?: Node): Promise<Node[]> {
    if (!node) {
      return await this.loadRoots();
    }
    if (node.kind === "cr") {
      return this.loadCRChildren(node.cr);
    }
    return [];
  }

  private async loadRoots(): Promise<Node[]> {
    const folder = findSpecsFolder();
    if (!folder) {
      return [];
    }
    const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
    const res = await runAndCapture(this.context, ["cr", "status", "--json"], cwd);
    if (res.exitCode !== 0) {
      getOutput().appendLine(res.stderr);
      return [];
    }
    try {
      const records = JSON.parse(res.stdout || "[]") as CRRecord[];
      return records.map((r) => new CRNode(r));
    } catch (err) {
      getOutput().appendLine(`cr status --json parse error: ${err}`);
      return [];
    }
  }

  private loadCRChildren(cr: CRRecord): Node[] {
    const children: Node[] = [];
    children.push(new StatNode(cr, cr.has_index ? "_index.md ✓" : "_index.md missing"));
    for (const sub of ["requirements", "features", "components", "architecture"]) {
      const dir = path.join(cr.dir, sub);
      if (!fs.existsSync(dir)) {
        continue;
      }
      for (const entry of walkMarkdown(dir, cr.dir)) {
        children.push(new FileNode(cr, entry.full, entry.rel));
      }
    }
    // Always include _index.md if present (for quick access).
    const idx = path.join(cr.dir, "_index.md");
    if (fs.existsSync(idx)) {
      children.unshift(new FileNode(cr, idx, "_index.md"));
    }
    return children;
  }
}

function countsLabel(cr: CRRecord): string {
  return `R:${cr.requirements} F:${cr.features} C:${cr.components} A:${cr.architecture}`;
}

function* walkMarkdown(dir: string, crDir: string): Generator<{ full: string; rel: string }> {
  const stack = [dir];
  while (stack.length > 0) {
    const d = stack.pop()!;
    let entries: fs.Dirent[];
    try {
      entries = fs.readdirSync(d, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const e of entries) {
      const full = path.join(d, e.name);
      if (e.isDirectory()) {
        stack.push(full);
      } else if (e.isFile() && e.name.endsWith(".md") && e.name !== "_index.md") {
        yield { full, rel: path.relative(crDir, full).split(path.sep).join("/") };
      }
    }
  }
}

export function registerCRTree(context: vscode.ExtensionContext): CRTreeProvider {
  const provider = new CRTreeProvider(context);
  context.subscriptions.push(
    vscode.window.registerTreeDataProvider("specs.changeRequests", provider),
    vscode.commands.registerCommand("specs.cr.refresh", () => provider.refresh()),
    vscode.commands.registerCommand("specs.cr.openDir", (node?: Node) => {
      if (node && node.kind === "cr") {
        vscode.commands.executeCommand("revealInExplorer", vscode.Uri.file(node.cr.dir));
      }
    }),
  );
  // Auto-refresh when change-requests/ changes.
  const folder = findSpecsFolder();
  if (folder) {
    const root = findSpecsRoot(folder) ?? folder.uri.fsPath;
    const watcher = vscode.workspace.createFileSystemWatcher(
      new vscode.RelativePattern(root, "change-requests/**"),
    );
    watcher.onDidCreate(() => provider.refresh());
    watcher.onDidDelete(() => provider.refresh());
    watcher.onDidChange(() => provider.refresh());
    context.subscriptions.push(watcher);
  }
  return provider;
}
